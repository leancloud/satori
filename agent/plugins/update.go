package plugins

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/agl/ed25519"

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
	if time.Now().Unix()-lastPluginUpdate < 300 {
		if debug {
			log.Println("Previous update too recent, do nothing")
		}
		return nil
	}

	parentDir := path.Dir(cfg.CheckoutPath)

	if !file.IsExist(parentDir) {
		os.MkdirAll(parentDir, os.ModePerm)
	}

	if ver == "" {
		ver = "origin/master"
	}

	if err := ensureGitRepo(cfg.CheckoutPath, cfg.Git); err != nil {
		log.Println(err.Error())
		reportFailure("git-fail", err.Error())
		return err
	}
	if err := updateByFetch(cfg.CheckoutPath); err != nil {
		log.Println(err.Error())
		reportFailure("git-fail", err.Error())
		return err
	}
	if len(cfg.SigningKeys) > 0 {
		if err := verifySignature(cfg.CheckoutPath, ver, cfg.SigningKeys); err != nil {
			log.Println(err.Error())
			reportFailure("signature-fail", err.Error())
			return err
		}
	} else {
		log.Println("Signing keys not configured, signature verification skipped")
	}

	if err := checkoutCommit(cfg.CheckoutPath, ver); err != nil {
		log.Println(err.Error())
		reportFailure("git-fail", err.Error())
		return err
	}
	log.Println("Update plugins complete")
	return nil
}

func ensureGitRepo(path string, remote string) error {
	var buf bytes.Buffer

	if !file.IsExist(path) {
		log.Println("Plugin repo does not exist, creating one")
		buf.Reset()
		cmd := exec.Command("git", "init", path)
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("Can't init plugin repo: %s\n%s", err, buf.String())
		}

		buf.Reset()
		cmd = exec.Command("git", "remote", "add", "origin", remote)
		cmd.Dir = path
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err = cmd.Run()
		if err != nil {
			os.RemoveAll(path)
			return fmt.Errorf("Can't set repo remote, aborting: %s", err)
		}
	}

	return nil
}

func updateByFetch(path string) error {
	var buf bytes.Buffer

	log.Println("Begin update plugins")
	updateInflight = true
	defer func() { updateInflight = false }()
	lastPluginUpdate = time.Now().Unix()

	buf.Reset()
	cmd := exec.Command("timeout", "120s", "git", "fetch")
	cmd.Dir = path
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Update plugins by fetch error: %s\n%s", err, buf.String())
	}
	return nil
}

func verifySignature(path string, ver string, validKeys []string) error {
	var buf bytes.Buffer
	var err error

	cmd := exec.Command("git", "cat-file", "-p", ver)
	cmd.Dir = path
	cmd.Stdout = &buf
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Can't get content of desired commit: %s\n%s", err, buf.String())
	}
	content := buf.String()

	tree := ""
	key := ""
	sign := ""
	for _, l := range strings.Split(content, "\n") {
		if strings.HasPrefix(l, "tree ") {
			tree = strings.TrimSpace(l[len("tree "):])
			continue
		}
		if strings.HasPrefix(l, "satori-sign ") {
			s := strings.TrimSpace(l[len("satori-sign "):])
			a := strings.Split(s, ":")
			keyid := a[0]
			for _, k := range validKeys {
				if strings.HasPrefix(k, keyid) {
					key = k
					break
				}
			}
			sign = a[1]
			continue
		}
	}
	if tree == "" {
		return fmt.Errorf("Can't find tree hash")
	} else if sign == "" {
		return fmt.Errorf("Signature not found")
	} else if key == "" {
		return fmt.Errorf("Signing key untrusted")
	}

	var vkslice []byte
	var vk [32]byte
	if vkslice, err = base64.StdEncoding.DecodeString(key); err != nil {
		return err
	}
	copy(vk[:], vkslice)

	var signslice []byte
	var s [64]byte
	if signslice, err = base64.StdEncoding.DecodeString(sign); err != nil {
		return err
	}
	copy(s[:], signslice)

	if !ed25519.Verify(&vk, []byte(tree), &s) {
		return fmt.Errorf("Signature invalid")
	}

	return nil
}

func checkoutCommit(path string, ver string) error {
	var buf bytes.Buffer

	cmd := exec.Command("git", "reset", "--hard", ver)
	cmd.Dir = path
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("git reset --hard failed: %s\n%s", err, buf.String())
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
			return fmt.Errorf("git reset --hard failed: %s\n%s", err, buf.String())
		}
	}
	return nil
}
