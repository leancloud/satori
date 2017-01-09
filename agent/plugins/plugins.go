package plugins

import (
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

	"github.com/leancloud/satori/agent/g"
	"github.com/leancloud/satori/common/model"
	"github.com/toolkits/file"
	"github.com/toolkits/sys"
)

type Plugin struct {
	FilePath string
	Step     int
	Params   []model.PluginParam

	killSwitch chan struct{}
}

func (this *Plugin) Run() {
	debug := g.Config().Debug
	if debug {
		log.Printf("Starting plugin scheduler for %s/%d/%s", this.FilePath, this.Step, this.Params)
	}

	if this.killSwitch != nil {
		this.Kill()
	}

	ticker := time.NewTicker(time.Duration(this.Step) * time.Second)
	this.killSwitch = make(chan struct{})

	go func() {
		for {
			select {
			case <-ticker.C:
				this.RunOnce()
			case <-this.killSwitch:
				ticker.Stop()
				return
			}
		}
	}()
}

func (this *Plugin) reportFailure(subject string, desc string) {
	hostname, _ := g.Hostname()
	now := time.Now().Unix()
	m := []*model.MetricValue{
		&model.MetricValue{
			Endpoint:  hostname,
			Metric:    ".satori.agent.plugin." + subject,
			Step:      0,
			Timestamp: now,
			Tags: map[string]string{
				"file": this.FilePath,
			},
			Desc: desc,
		},
	}
	g.SendToTransfer(m)
}

func (this *Plugin) RunOnce() {
	cfg := g.Config().Plugin
	timeout := this.Step*1000 - 500
	fpath := filepath.Join(cfg.CheckoutPath, cfg.Subdir, this.FilePath)

	if !file.IsExist(fpath) {
		log.Println("no such plugin:", fpath)
		return
	}

	debug := g.Config().Debug
	if debug {
		log.Println(fpath, "running...")
	}

	cmd := exec.Command(fpath)
	if this.Params != nil {
		var stdin bytes.Buffer
		s, err := json.Marshal(this.Params)
		if err != nil {
			log.Println("Error marshalling params for metric plugin: %s", this.FilePath)
			return
		}
		stdin.Write(s)
		cmd.Stdin = &stdin
	}
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.Start()

	err, isTimeout := sys.CmdRunWithTimeout(cmd, time.Duration(timeout)*time.Millisecond)
	errStr := stderr.String()

	if errStr != "" {
		logFile := filepath.Join(cfg.LogDir, this.FilePath+".stderr.log")
		if _, err = file.WriteString(logFile, errStr); err != nil {
			log.Printf("[ERROR] write log to %s fail, error: %s\n", logFile, err)
		}
		this.reportFailure("error", errStr)
	}

	if isTimeout {
		// has been killed
		if err == nil && debug {
			log.Println("[INFO] timeout and kill process", fpath, "successfully")
			this.reportFailure("timeout", "")
		}

		if err != nil {
			log.Println("[ERROR] kill process", fpath, "occur error:", err)
			this.reportFailure("cant-kill", "")
		}

		return
	}

	if err != nil {
		log.Println("[ERROR] exec plugin", fpath, "fail. error:", err)
		this.reportFailure("error", err.Error())
		return
	}

	// exec successfully
	data := stdout.Bytes()
	if len(data) == 0 {
		if debug {
			log.Println("[DEBUG] stdout of", fpath, "is blank")
		}
		this.reportFailure("no-stdout", "")
		return
	}

	var metrics []*model.MetricValue
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		s := stdout.String()
		log.Printf("[ERROR] json.Unmarshal stdout of %s fail. error:%s stdout: \n%s\n", fpath, err, s)
		this.reportFailure("bad-format", err.Error()+"\n\n"+stdout.String())
		return
	}

	g.SendToTransfer(metrics)
}

func (this *Plugin) Kill() {
	if this.killSwitch != nil {
		debug := g.Config().Debug
		if debug {
			log.Printf("Stopping plugin scheduler for %s/%d/%s", this.FilePath, this.Step, this.Params)
		}
		close(this.killSwitch)
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
		var metric string = ""
		var step int = -1

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
		log.Printf("Reschedule for %s plugins", len(L))
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
