package g

import (
	"log"
	"sync"

	"github.com/toolkits/file"
	"gopkg.in/yaml.v2"
)

type GlobalConfig struct {
	Debug    bool     `yaml:"debug"`
	Listen   string   `yaml:"listen"`
	Backends []string `yaml:"backends"`
}

var (
	ConfigFile string
	config     *GlobalConfig
	configLock = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
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

	configLock.Lock()
	defer configLock.Unlock()
	config = &c

	log.Println("g.ParseConfig ok, file ", cfg)
}

// CLUSTER NODE
type ClusterNode struct {
	Addrs []string `yaml:"addrs"`
}

func NewClusterNode(addrs []string) *ClusterNode {
	return &ClusterNode{addrs}
}
