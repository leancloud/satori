package plugins

import (
	"bufio"
	"bytes"
	"encoding/json"
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

	killSwitch chan struct{}
}

func (p *Plugin) Run() {
	debug := g.Config().Debug
	if debug {
		log.Printf("Starting plugin scheduler for %s/%d/%s", p.FilePath, p.Step, p.Params)
	}

	p.Kill()
	p.killSwitch = make(chan struct{})

	dur := p.Step
	if dur <= 0 {
		dur = 1
	}

	ticker := time.NewTicker(time.Duration(dur) * time.Second)
	go func() {
		s := p.killSwitch
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

func (p *Plugin) RunOnce() {
	cfg := g.Config().Plugin
	fpath := filepath.Join(cfg.CheckoutPath, cfg.Subdir, p.FilePath)

	if !file.IsExist(fpath) {
		log.Println("no such plugin:", fpath)
		return
	}

	debug := g.Config().Debug
	if debug {
		log.Println(fpath, "running...")
	}

	cmd := exec.Command(fpath)
	if p.Params != nil {
		var stdin bytes.Buffer
		s, err := json.Marshal(p.Params)
		if err != nil {
			log.Println("Error marshalling params for metric plugin: %s", p.FilePath)
			return
		}
		stdin.Write(s)
		cmd.Stdin = &stdin
	}
	stdoutPipe, _ := cmd.StdoutPipe()
	stdout := bufio.NewReader(stdoutPipe)
	stderrPipe, _ := cmd.StderrPipe()
	stderr := bufio.NewReader(stderrPipe)
	err := cmd.Start()

	if err != nil {
		log.Println("[ERROR] exec plugin", fpath, "fail. error:", err)
		p.reportFailure("error", err.Error())
		return
	}

	go func() {
		for {
			s, err := stdout.ReadBytes('\n')
			if err != nil {
				return
			}
			var metrics []*model.MetricValue
			err = json.Unmarshal(s, &metrics)
			if err != nil {
				log.Printf("[ERROR] json.Unmarshal stdout of %s fail. error:%s stdout: \n%s\n", fpath, err, s)
				p.reportFailure("bad-format", err.Error()+"\n\n"+string(s))
				stderrPipe.Close()
				stdoutPipe.Close()
				go func() {
					proc := cmd.Process
					time.Sleep(time.Second * 5)
					proc.Kill()
				}()
				return
			}
			g.SendToTransfer(metrics)
		}
	}()

	go func() {
		s, _ := ioutil.ReadAll(stderr)
		if len(s) > 0 {
			p.reportFailure("error", string(s))
		}
	}()

	finished := make(chan error, 1)
	go func() {
		finished <- cmd.Wait()
	}()

	var timeout <-chan time.Time
	if p.Step > 0 {
		t := p.Step*1000 - 500
		timeout = time.After(time.Duration(t) * time.Millisecond)
	} else {
		// Long running plugin
		timeout = make(chan time.Time)
	}

	killSwitch := p.killSwitch

	select {
	case <-finished:
		break
	case <-timeout:
		cmd.Process.Kill()
		log.Println("[INFO] Plugin timed out, terminating: ", fpath)
		p.reportFailure("timeout", "")
	case <-killSwitch:
		cmd.Process.Kill()
		log.Println("[INFO] Plugin was asked to terminate: ", fpath)
	}
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
