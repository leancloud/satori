package plugins

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/leancloud/satori/agent/g"
	"github.com/leancloud/satori/common/model"
	"github.com/toolkits/file"
)

func reportFailure(subject string, desc string) {
	hostname, _ := g.Hostname()
	now := time.Now().Unix()
	m := []*model.MetricValue{
		&model.MetricValue{
			Endpoint:  hostname,
			Metric:    ".satori.agent.plugin." + subject,
			Value:     1,
			Step:      1,
			Timestamp: now,
			Tags:      map[string]string{},
			Desc:      desc,
		},
	}
	g.SendToTransfer(m)
}

func GetCurrentPluginVersion() (string, error) {
	cfg := g.Config().Plugin
	if !cfg.Enabled {
		return "", fmt.Errorf("plugin-not-enabled")
	}

	pluginDir := cfg.CheckoutPath
	if !file.IsExist(pluginDir) {
		reportFailure("plugin-dir-does-not-exist", "")
		return "", fmt.Errorf("plugin-dir-does-not-exist")
	}

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = pluginDir

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		reportFailure("git-fail", err.Error()+"\n"+stderr.String())
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

var updateInflight bool = false
var lastPluginUpdate int64 = 0

func UpdatePlugin(ver string) error {
	debug := g.Config().Debug
	cfg := g.Config().Plugin

	if !cfg.Enabled {
		if debug {
			log.Println("plugin not enabled, not updating")
		}
		return fmt.Errorf("plugin not enabled")
	}

	if updateInflight {
		if debug {
			log.Println("Previous update inflight, do nothing")
		}
		return nil
	}

	// TODO: add to config
	if time.Now().Unix()-lastPluginUpdate < 600 {
		if debug {
			log.Println("Previous update too recent, do nothing")
		}
		return nil
	}

	parentDir := file.Dir(cfg.CheckoutPath)
	file.InsureDir(parentDir)

	if ver == "" {
		ver = "origin/master"
	}

	var buf bytes.Buffer

	if file.IsExist(cfg.CheckoutPath) {
		// git fetch
		log.Println("Begin update plugins by fetch")
		updateInflight = true
		defer func() { updateInflight = false }()
		lastPluginUpdate = time.Now().Unix()

		buf.Reset()
		cmd := exec.Command("timeout", "120s", "git", "fetch")
		cmd.Dir = cfg.CheckoutPath
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err := cmd.Run()
		if err != nil {
			s := fmt.Sprintf("Update plugins by fetch error: %s", err)
			log.Println(s)
			reportFailure("git-fail", s+"\n"+buf.String())
			return fmt.Errorf("git fetch in dir:%s fail. error: %s", cfg.CheckoutPath, err)
		}

		buf.Reset()
		cmd = exec.Command("git", "reset", "--hard", ver)
		cmd.Dir = cfg.CheckoutPath
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err = cmd.Run()
		if err != nil {
			s := fmt.Sprintf("git reset --hard failed: %s", err)
			log.Println(s)
			reportFailure("git-fail", s+"\n"+buf.String())
			return fmt.Errorf("git reset --hard in dir:%s fail. error: %s", cfg.CheckoutPath, err)
		}
		log.Println("Update plugins by fetch complete")
	} else {
		// git clone
		log.Println("Begin update plugins by clone")
		lastPluginUpdate = time.Now().Unix()
		buf.Reset()
		cmd := exec.Command("timeout", "120s", "git", "clone", cfg.Git, file.Basename(cfg.CheckoutPath))
		cmd.Dir = parentDir
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err := cmd.Run()
		if err != nil {
			s := fmt.Sprintf("Update plugins by clone error: %s", err)
			log.Println(s)
			reportFailure("git-fail", s+"\n"+buf.String())
			return fmt.Errorf("git clone in dir:%s fail. error: %s", parentDir, err)
		}

		buf.Reset()
		cmd = exec.Command("git", "reset", "--hard", ver)
		cmd.Dir = cfg.CheckoutPath
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err = cmd.Run()
		if err != nil {
			s := fmt.Sprintf("git reset --hard failed: %s", err)
			log.Println(s)
			reportFailure("git-fail", s+"\n"+buf.String())
			return fmt.Errorf("git reset --hard in dir:%s fail. error: %s", cfg.CheckoutPath, err)
		}
		log.Println("Update plugins by clone complete")
	}
	return nil
}

func ForceResetPlugin() error {
	cfg := g.Config().Plugin
	if !cfg.Enabled {
		return fmt.Errorf("plugin not enabled")
	}

	dir := cfg.CheckoutPath

	if file.IsExist(dir) {
		var buf bytes.Buffer
		cmd := exec.Command("git", "reset", "--hard")
		cmd.Dir = dir
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err := cmd.Run()
		if err != nil {
			s := fmt.Sprintf("git reset --hard failed: %s", err)
			log.Println(s)
			reportFailure("git-fail", s+"\n"+buf.String())
			return fmt.Errorf("git reset --hard in dir:%s fail. error: %s", dir, err)
		}
	}
	return nil
}
