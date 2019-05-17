package plugins

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/agl/ed25519"
	"github.com/cavaliercoder/grab"
	"gopkg.in/yaml.v2"

	"github.com/leancloud/satori/agent/g"
	"github.com/leancloud/satori/common/model"
	"github.com/toolkits/file"
)

func reportFailure(subject string, desc string) {
	hostname := g.Hostname()
	now := time.Now().Unix()
	m := []*model.MetricValue{
		&model.MetricValue{
			Endpoint:  hostname,
			Metric:    ".satori.agent.plugin." + subject,
			Value:     1,
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

var updateInflight bool
var lastPluginUpdate int64

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
		s := "Previous update inflight, do nothing"
		if debug {
			log.Println(s)
		}
		return fmt.Errorf(s)
	}

	// TODO: add to config
	if time.Now().Unix()-lastPluginUpdate < 300 {
		s := "Previous update too recent, do nothing"
		if debug {
			log.Println(s)
		}
		return fmt.Errorf(s)
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
		keys := []string{}
		for _, k := range cfg.SigningKeys {
			keys = append(keys, k.Key)
		}
		if cfg.AuthorizedKeys != "" {
			altKeys, err := getAuthorizedKeys(cfg.CheckoutPath, ver, cfg.AuthorizedKeys, keys)
			if err != nil {
				log.Println("Failed to get alternative signing keys: " + err.Error())
				reportFailure("alt-key-fail", err.Error())
			} else {
				if debug {
					for _, k := range altKeys {
						log.Printf("Got alt key: [%s]\n", k)
					}
				}
				keys = append(altKeys, keys...)
			}
		}
		if err := verifySignature(cfg.CheckoutPath, ver, keys); err != nil {
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
			return fmt.Errorf("Can't add repo remote, aborting: %s\n%s", err, buf.String())
		}
	} else {
		buf.Reset()
		cmd := exec.Command("git", "remote", "set-url", "origin", remote)
		cmd.Dir = path
		cmd.Stdout = &buf
		cmd.Stderr = &buf
		err := cmd.Run()
		if err != nil {
			os.RemoveAll(path)
			return fmt.Errorf("Can't set repo remote, aborting: %s\n%s", err, buf.String())
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

func verifySignature(checkoutPath string, head string, validKeys []string) error {
	var buf bytes.Buffer
	var err error

	cmd := exec.Command("git", "cat-file", "-p", head)
	cmd.Dir = checkoutPath
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
		if tree == "" && strings.HasPrefix(l, "tree ") {
			tree = strings.TrimSpace(l[len("tree "):])
			continue
		}
		if strings.HasPrefix(l, "satori-sign ") {
			s := strings.TrimSpace(l[len("satori-sign "):])
			a := strings.Split(s, ":")
			keyid := a[0]
			for _, k := range validKeys {
				if strings.HasPrefix(k, keyid) {
					key = strings.Fields(k)[0]
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

func getAuthorizedKeys(checkoutPath string, head string, keyFile string, validKeys []string) ([]string, error) {
	var buf bytes.Buffer
	var err error

	cmd := exec.Command("git", "rev-list", "-1", head, "--", keyFile)
	cmd.Dir = checkoutPath
	cmd.Stdout = &buf
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Can't get most recent commit hash of key file: %s\n%s", err, buf.String())
	}
	mostRecentHash := strings.TrimSpace(buf.String())

	if err = verifySignature(checkoutPath, mostRecentHash, validKeys); err != nil {
		return nil, err
	}

	buf.Reset()

	cmd = exec.Command("git", "show", fmt.Sprintf("%s:%s", head, keyFile))
	cmd.Dir = checkoutPath
	cmd.Stdout = &buf
	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("Can't get content of key file: %s\n%s", err, buf.String())
	}
	content := buf.Bytes()

	parsed := make([]struct {
		Key string `yaml:"key"`
	}, 0, 5)

	if err = yaml.Unmarshal(content, &parsed); err != nil {
		return nil, err
	}

	lst := make([]string, 0, 5)
	for _, l := range parsed {
		lst = append(lst, l.Key)
	}

	return lst, nil
}

func checkoutCommit(checkoutPath string, head string) error {
	var buf bytes.Buffer

	cmd := exec.Command("git", "reset", "--hard", head)
	cmd.Dir = checkoutPath
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

func TryUpdateAgent() error {
	cfg := g.Config().Plugin
	if cfg.Update == "" {
		return nil
	}

	uconfPath := path.Join(cfg.CheckoutPath, cfg.Update)
	uconfString, err := ioutil.ReadFile(uconfPath)
	if err != nil {
		return fmt.Errorf("Can't read update conf file %s: %s", uconfPath, err.Error())
	}

	var uconf struct {
		Sha256 string `yaml:"sha256"`
		Url    string `yaml:"url"`
	}

	if err = yaml.UnmarshalStrict(uconfString, &uconf); err != nil {
		return fmt.Errorf("Can't parse update conf: %s", err.Error())
	}

	if uconf.Sha256 == "" || uconf.Url == "" {
		return fmt.Errorf("Update conf incomplete, do nothing")
	}

	newHash, err := hex.DecodeString(uconf.Sha256)
	if err != nil {
		return fmt.Errorf("Can't parse sha256: %s", err.Error())
	}

	if bytes.Equal(g.BinaryHash, newHash) {
		return nil
	}

	newPath := g.BinaryPath + "." + strconv.FormatInt(rand.Int63(), 10)
	if file.IsExist(newPath) {
		os.Remove(newPath)
	}

	req, err := grab.NewRequest(newPath, uconf.Url)
	if err != nil {
		return fmt.Errorf("Can't create download request: %s", err.Error())
	}
	req.SetChecksum(sha256.New(), newHash, true)
	cli := grab.NewClient()
	resp := cli.Do(req)
	if err = resp.Err(); err != nil {
		return fmt.Errorf("Can't download satori-agent: %s", err.Error())
	}

	script := fmt.Sprintf(
		"SELF=\"%s\"\nRENAME=\"%s\"\nNEW=\"%s\"\n",
		g.BinaryPath, g.BinaryPath+"."+hex.EncodeToString(g.BinaryHash), newPath,
	)

	script += `
	set -e
	if [ ! -f "$NEW" ]; then
		exit 1
	fi
	if [ -f "$RENAME" ]; then
		rm -f $RENAME
	fi
	mv $SELF $RENAME
	mv $NEW $SELF
	chmod 0755 $SELF
	`

	cmd := exec.Command("bash", "-c", script)
	if err := cmd.Run(); err != nil {
		return err
	}

	log.Println("Update triggered, restarting")
	syscall.Exec(g.BinaryPath, os.Args, os.Environ())

	return fmt.Errorf("Can't do exec")
}
