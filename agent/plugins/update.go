package plugins

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/leancloud/satori/agent/g"
	"github.com/toolkits/file"
)

func GetCurrentPluginVersion() (string, error) {
	cfg := g.Config().Plugin
	if !cfg.Enabled {
		return "", fmt.Errorf("plugin-not-enabled")
	}

	pluginDir := cfg.CheckoutPath
	if !file.IsExist(pluginDir) {
		return "", fmt.Errorf("plugin-dir-does-not-exist")
	}

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = pluginDir

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(out.String()), nil
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

	if file.IsExist(cfg.CheckoutPath) {
		// git fetch
		log.Println("Begin update plugins by fetch")
		updateInflight = true
		defer func() { updateInflight = false }()
		lastPluginUpdate = time.Now().Unix()
		cmd := exec.Command("git", "fetch")
		cmd.Dir = cfg.CheckoutPath
		err := cmd.Run()
		if err != nil {
			log.Println("Update plugins by fetch error: %s", err)
			return fmt.Errorf("git fetch in dir:%s fail. error: %s", cfg.CheckoutPath, err)
		}

		cmd = exec.Command("git", "reset", "--hard", ver)
		cmd.Dir = cfg.CheckoutPath
		err = cmd.Run()
		if err != nil {
			log.Println("git reset --hard failed: %s", err)
			return fmt.Errorf("git reset --hard in dir:%s fail. error: %s", cfg.CheckoutPath, err)
		}
		log.Println("Update plugins by fetch complete")
	} else {
		// git clone
		log.Println("Begin update plugins by clone")
		lastPluginUpdate = time.Now().Unix()
		cmd := exec.Command("git", "clone", cfg.Git, file.Basename(cfg.CheckoutPath))
		cmd.Dir = parentDir
		err := cmd.Run()
		if err != nil {
			log.Println("Update plugins by clone error: %s", err)
			return fmt.Errorf("git clone in dir:%s fail. error: %s", parentDir, err)
		}

		cmd = exec.Command("git", "reset", "--hard", ver)
		cmd.Dir = cfg.CheckoutPath
		err = cmd.Run()
		if err != nil {
			log.Println("git reset --hard failed: %s", err)
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
		cmd := exec.Command("git", "reset", "--hard")
		cmd.Dir = dir
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("git reset --hard in dir:%s fail. error: %s", dir, err)
		}
	}
	return nil
}
