package g

import (
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/toolkits/file"
	tknet "github.com/toolkits/net"
	"gopkg.in/yaml.v2"
)

type GlobalConfig struct {
	Debug     bool     `yaml:"debug"`
	Hostname  string   `yaml:"hostname"`
	FQDN      bool     `yaml:"fqdn"`
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

	hostnameCache string
)

func Config() *GlobalConfig {
	lock.RLock()
	defer lock.RUnlock()
	return config
}

func toFQDN(ip string) string {
	var err error
	hosts, err := net.LookupAddr(ip)
	if err != nil || len(hosts) == 0 {
		return ""
	}
	fqdn := hosts[0]
	return strings.TrimSuffix(fqdn, ".") // return fqdn without trailing dot
}

func Hostname() string {
	if hostnameCache != "" {
		return hostnameCache
	}

	hostname := Config().Hostname
	if hostname != "" {
		hostnameCache = hostname
		return hostname
	}

	hostname, err := os.Hostname()
	if err != nil {
		panic("ERROR: os.Hostname() fail")
	}

	hostnameCache = hostname

	if Config().FQDN {
		if fqdn := toFQDN(IP()); fqdn != "" {
			hostnameCache = fqdn
			return fqdn
		}
	}

	return hostname
}

func IP() string {
	ip := Config().IP
	if ip != "" {
		// use ip in configuration
		return ip
	}

	ips, err := tknet.IntranetIP()
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
