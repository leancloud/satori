package g

import (
	"log"
	"os"
	"regexp"
	"sync"

	"github.com/toolkits/file"
	"github.com/toolkits/net"
	"gopkg.in/yaml.v2"
)

type GlobalConfig struct {
	Debug     bool     `yaml:"debug"`
	Hostname  string   `yaml:"hostname"`
	IP        string   `yaml:"ip"`
	Master    string   `yaml:"master"`
	Transfer  []string `yaml:"transfer"`
	HTTP      string   `yaml:"http"`
	NoBuiltin bool     `yaml:"noBuiltin"`
	Cgroups   *struct {
		Memory  int64   `yaml:"mem"`
		CPU     float32 `yaml:"cpu"`
		Enforce bool    `yaml:"enforce"`
	} `yaml:"cgroups"`
	Plugin struct {
		Enabled     bool `yaml:"enabled"`
		SigningKeys []struct {
			Key string `yaml:"key"`
		} `yaml:"signingKeys"`
		AuthorizedKeys string `yaml:"authorizedKeys"`
		Update         string `yaml:"update"`
		Git            string `yaml:"git"`
		CheckoutPath   string `yaml:"checkoutPath"`
		Subdir         string `yaml:"subDir"`
		LogDir         string `yaml:"logs"`
	} `yaml:"plugin"`
	Ignore []struct {
		Metric   string `yaml:"metric"`
		Tag      string `yaml:"tag"`
		TagValue string `yaml:"tagValue"`
	} `yaml:"ignore"`
	Collector *struct {
		IfacePrefix []string `yaml:"ifacePrefix"`
	} `yaml:"collector"`
	AddTags map[string]string `yaml:"addTags"`
}

var (
	ConfigFile string
	config     *GlobalConfig
	lock       = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	lock.RLock()
	defer lock.RUnlock()
	return config
}

func Hostname() (string, error) {
	hostname := Config().Hostname
	if hostname != "" {
		return hostname, nil
	}

	hostname, err := os.Hostname()
	if err != nil {
		log.Println("ERROR: os.Hostname() fail", err)
	}
	return hostname, err
}

func IP() string {
	ip := Config().IP
	if ip != "" {
		// use ip in configuration
		return ip
	}

	ips, err := net.IntranetIP()
	if err != nil {
		log.Fatalln("get intranet ip fail:", err)
	}

	if len(ips) > 0 {
		ip = ips[0]
	} else {
		ip = "127.0.0.1"
	}

	return ip
}

func ParseConfig(cfg string) {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent. maybe you need `mv cfg.example.yaml cfg.yaml`")
	}

	ConfigFile = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, "fail:", err)
	}

	var c GlobalConfig
	err = yaml.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail:", err)
	}

	for _, item := range c.Ignore {
		regexp.MustCompile(item.Metric)
		regexp.MustCompile(item.Tag)
		regexp.MustCompile(item.TagValue)
	}

	lock.Lock()
	defer lock.Unlock()

	config = &c

	log.Println("read config file:", cfg, "successfully")
}
