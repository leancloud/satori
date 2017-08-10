package g

import (
	"encoding/json"
	"log"
	"os"
	"regexp"
	"sync"

	"github.com/toolkits/file"
	"github.com/toolkits/net"
)

type PluginConfig struct {
	Enabled            bool     `json:"enabled"`
	SigningKeys        []string `json:"signingKeys"`
	AltSigningKeysFile string   `json:"altSigningKeysFile"`
	Git                string   `json:"git"`
	CheckoutPath       string   `json:"checkoutPath"`
	Subdir             string   `json:"subDir"`
	LogDir             string   `json:"logs"`
}

type MasterConfig struct {
	Enabled  bool   `json:"enabled"`
	Addr     string `json:"addr"`
	Interval int    `json:"interval"`
	Timeout  int    `json:"timeout"`
}

type TransferConfig struct {
	Enabled  bool     `json:"enabled"`
	Addrs    []string `json:"addrs"`
	Interval int      `json:"interval"`
	Timeout  int      `json:"timeout"`
}

type HttpConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
}

type CollectorConfig struct {
	IfacePrefix []string `json:"ifacePrefix"`
}

type GlobalConfig struct {
	Debug      bool              `json:"debug"`
	Hostname   string            `json:"hostname"`
	IP         string            `json:"ip"`
	Plugin     *PluginConfig     `json:"plugin"`
	Master     *MasterConfig     `json:"master"`
	Transfer   *TransferConfig   `json:"transfer"`
	Http       *HttpConfig       `json:"http"`
	Collector  *CollectorConfig  `json:"collector"`
	Ignore     [][3]string       `json:"ignore"`
	AddTags    map[string]string `json:"addTags"`
	NoBuiltin  bool              `json:"noBuiltin"`
	SelfUpdate bool              `json:"selfUpdate"`
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
		log.Fatalln("config file:", cfg, "is not existent. maybe you need `mv cfg.example.json cfg.json`")
	}

	ConfigFile = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, "fail:", err)
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail:", err)
	}

	for _, item := range c.Ignore {
		for _, v := range item {
			regexp.MustCompile(v)
		}
	}

	lock.Lock()
	defer lock.Unlock()

	config = &c

	log.Println("read config file:", cfg, "successfully")
}
