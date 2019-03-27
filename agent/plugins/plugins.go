package plugins

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/toolkits/file"

	"github.com/leancloud/satori/agent/g"
	"github.com/leancloud/satori/common/model"
)

func safeClose(c chan struct{}) {
	defer func() { _ = recover() }()
	close(c)
}

func closed(c chan struct{}) bool {
	select {
	case <-c:
		return true
	default:
		return false
	}
}

type Plugin struct {
	FilePath string
	Step     int
	Params   []model.PluginParam

	proc       *exec.Cmd
	stdout     *bufio.Reader
	stdoutPipe io.ReadCloser
	stderr     *bufio.Reader
	stderrPipe io.ReadCloser
	finished   chan struct{}
	killSwitch chan struct{}
}

func (p *Plugin) Run() {
	debug := g.Config().Debug
	if debug {
		log.Printf("Starting plugin scheduler for %s/%d/%s", p.FilePath, p.Step, p.Params)
	}

	p.Kill()

	dur := p.Step
	if dur <= 0 {
		dur = 1
	}

	ticker := time.NewTicker(time.Duration(dur) * time.Second)

	go func() {
		s := make(chan struct{})
		p.killSwitch = s
		for {
			<-ticker.C
			if closed(s) {
				return
			}

			p.RunOnce()
		}
	}()
}

func (p *Plugin) reportFailure(subject string, desc string) {
	hostname, _ := g.Hostname()
	now := time.Now().Unix()
	m := []*model.MetricValue{
		&model.MetricValue{
			Endpoint:  hostname,
			Metric:    ".satori.agent.plugin." + subject,
			Value:     1,
			Step:      1,
			Timestamp: now,
			Tags: map[string]string{
				"file": p.FilePath,
			},
			Desc: desc,
		},
	}
	g.SendToTransfer(m)
}

func (p *Plugin) setupPipes() {
	p.stdoutPipe, _ = p.proc.StdoutPipe()
	p.stdout = bufio.NewReader(p.stdoutPipe)
	p.stderrPipe, _ = p.proc.StderrPipe()
	p.stderr = bufio.NewReader(p.stderrPipe)
	p.finished = make(chan struct{})
}

func (p *Plugin) teardownPipes() {
	if p.stdout != nil {
		_ = p.stdoutPipe.Close()
		p.stdoutPipe = nil
		p.stdout = nil
	}
	if p.stderr != nil {
		_ = p.stderrPipe.Close()
		p.stderrPipe = nil
		p.stderr = nil
	}
}

func (p *Plugin) sendStdout() {
	for {
		if p.stdout == nil {
			return
		}
		s, err := p.stdout.ReadBytes('\n')
		if err != nil {
			p.teardownPipes()
			safeClose(p.finished)
			return
		}
		var metrics []*model.MetricValue
		err = json.Unmarshal(s, &metrics)
		if err != nil {
			log.Printf("[ERROR] json.Unmarshal stdout of %s fail. error:%s stdout: \n%s\n", p.FilePath, err, s)
			p.reportFailure("bad-format", err.Error()+"\n\n"+string(s))
			p.teardownPipes()
			safeClose(p.finished)
			return
		}
		g.SendToTransfer(metrics)
	}
}

func (p *Plugin) sendStderr() {
	if p.stderr != nil {
		s, _ := ioutil.ReadAll(p.stderr)
		if len(s) > 0 {
			p.reportFailure("error", string(s))
		}
	}
}

func (p *Plugin) setupCommand() error {
	cfg := g.Config().Plugin
	fpath := filepath.Join(cfg.CheckoutPath, cfg.Subdir, p.FilePath)
	if !file.IsExist(fpath) {
		return fmt.Errorf("No such plugin: %s", p.FilePath)
	}

	cmd := exec.Command(fpath)
	if p.Params != nil {
		var stdin bytes.Buffer
		s, err := json.Marshal(p.Params)
		if err != nil {
			return fmt.Errorf("Error marshalling params for metric plugin: %s, %s", p.FilePath, err)
		}
		stdin.Write(s)
		cmd.Stdin = &stdin
	}
	p.proc = cmd
	return nil
}

func (p *Plugin) terminateCommand() {
	debug := g.Config().Debug
	if p.proc != nil {
		cmd := p.proc
		p.proc = nil
		p.teardownPipes()
		if err := cmd.Process.Kill(); err != nil && debug {
			log.Printf("Error killing proc: %s\n", err)
		}
		_ = cmd.Wait()
	}
}

func (p *Plugin) makeTimeoutChan() <-chan time.Time {
	if p.Step > 0 {
		// Periodically called plugin
		t := p.Step*1000 - 500
		return time.After(time.Duration(t) * time.Millisecond)
	}

	// Long running plugin
	return make(chan time.Time)
}

func (p *Plugin) RunOnce() {
	debug := g.Config().Debug

	if err := p.setupCommand(); err != nil {
		log.Printf("Can't setup plugin %s command: %s\n", p.FilePath, err)
		p.reportFailure("error", err.Error())
		return
	}
	p.setupPipes()
	if err := p.proc.Start(); err != nil {
		log.Printf("Can't start plugin %s process: %s\n", p.FilePath, err)
		p.reportFailure("error", err.Error())
		return
	}

	if debug {
		log.Println(p.FilePath, "running...")
	}

	go p.sendStdout()
	go p.sendStderr()

	select {
	case <-p.finished:
		break
	case <-p.makeTimeoutChan():
		log.Println("[INFO] Plugin timed out, terminating: ", p.FilePath)
		p.reportFailure("timeout", "")
	case <-p.killSwitch:
		log.Println("[INFO] Plugin was asked to terminate: ", p.FilePath)
	}
	p.terminateCommand()
}

func (p *Plugin) Kill() {
	if p.killSwitch != nil {
		debug := g.Config().Debug
		if debug {
			log.Printf("Stopping plugin scheduler for %s/%d/%s", p.FilePath, p.Step, p.Params)
		}
		close(p.killSwitch)
	}
}

var (
	PluginsLock = new(sync.RWMutex)
	Plugins     = make([]*Plugin, 0)
)

// key: sys/ntp/60_ntp.py
func getDirPlugins(relativePath string, L *[]*Plugin) {
	if relativePath == "" || relativePath == "_metric" {
		return
	}

	cfg := g.Config().Plugin
	dir := filepath.Join(cfg.CheckoutPath, cfg.Subdir, relativePath)

	if !file.IsExist(dir) || file.IsFile(dir) {
		return
	}

	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Println("can not list files under", dir)
		return
	}

	for _, f := range fs {
		if f.IsDir() {
			continue
		}

		filename := f.Name()
		arr := strings.Split(filename, "_")
		if len(arr) < 2 {
			continue
		}

		// filename should be: $step_$xx
		var step int
		step, err = strconv.Atoi(arr[0])
		if err != nil {
			continue
		}

		fpath := filepath.Join(relativePath, filename)
		plugin := &Plugin{
			FilePath: fpath,
			Step:     step,
			Params:   nil,
		}
		*L = append(*L, plugin)
	}
}

func getMetricPlugins(metrics []model.PluginParam, L *[]*Plugin) {
	type Group struct {
		Metric string
		Step   int
	}

	g := make(map[Group][]model.PluginParam)
	for _, p := range metrics {
		var metric string
		step := -1

		if m, ok := p["_metric"]; ok {
			if s, ok := m.(string); ok {
				metric = s
			}
		}

		if c, ok := p["_step"]; ok {
			if s, ok := c.(float64); ok {
				step = int(s)
			}
		}

		if metric == "" || step == -1 {
			continue
		}

		t := Group{metric, step}
		g[t] = append(g[t], p)
	}

	for k, v := range g {
		*L = append(*L, &Plugin{
			FilePath: filepath.Join("_metric", k.Metric),
			Step:     k.Step,
			Params:   v,
		})
	}
}

func RunPlugins(dirs []string, metrics []model.PluginParam) {
	L := make([]*Plugin, 0)

	for _, d := range dirs {
		getDirPlugins(d, &L)
	}

	getMetricPlugins(metrics, &L)

	debug := g.Config().Debug
	if debug {
		log.Printf("Reschedule for %d plugins", len(L))
	}

	PluginsLock.Lock()
	defer PluginsLock.Unlock()

	for _, p := range Plugins {
		p.Kill()
	}

	Plugins = L
	for _, p := range L {
		p.Run()
	}
}
