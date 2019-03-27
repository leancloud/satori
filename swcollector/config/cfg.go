package config

import (
	"encoding/json"
	"sync"

	"github.com/leancloud/satori/swcollector/logging"
)

type CustomMetric struct {
	IpRange []string          `json:"ipRange"`
	Metric  string            `json:"metric"`
	Tags    map[string]string `json:"tags"`
	Oid     string            `json:"oid"`
}

type GlobalConfig struct {
	Interval int64  `json:"interval" default:30`
	LogFile  string `json:"logFile"`

	IpRange       []string `json:"ipRange"`
	Gosnmp        bool     `json:"gosnmp" default:"true"`
	PingTimeout   int      `json:"pingTimeout" default:"300"`
	PingRetry     int      `json:"pingRetry" default:"4"`
	SnmpCommunity string   `json:"snmpCommunity" default:"\"public\""`
	SnmpTimeout   int      `json:"snmpTimeout" default:"1000"`
	SnmpRetry     int      `json:"snmpRetry" default:"5"`
	IgnoreIface   []string `json:"ignoreIface" default:"[\"Nu\", \"NU\", \"Vlan\", \"Vl\"]"`
	Ignore        []string `json:"ignore"`
	FastPingMode  bool     `json:"fastPingMode"`

	ConcurrentQueriesPerHost int `json:"concurrentQueriesPerHost" default:"4"`
	ConcurrentCollectors     int `json:"concurrentCollectors" default:"1000"`

	CustomMetrics []CustomMetric    `json:"customMetrics"`
	ReverseLookup bool              `json:"reverseLookup" default:"false"`
	CustomHosts   map[string]string `json:"customHosts"`
}

var (
	config  *GlobalConfig
	cfgLock = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	cfgLock.RLock()
	defer cfgLock.RUnlock()
	return config
}

func ParseConfig(cfg []byte) {
	var c []GlobalConfig
	err := json.Unmarshal(cfg, &c)
	if err != nil {
		logging.Fatalln("parse config:", err)
	}
	if len(c) > 1 {
		logging.Fatalln("swcollector cannot handle multiple configs this time, please schedule each swcollector to a different node")
	} else if len(c) < 1 {
		logging.Fatalln("Missing config")
	}
	FillDefault(&c[0])
	cfgLock.Lock()
	defer cfgLock.Unlock()
	config = &c[0]
}
