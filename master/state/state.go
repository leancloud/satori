package state

import (
	"github.com/leancloud/satori/common/model"

	"sync"
)

type AgentInfo struct {
	Hostname      string `json:"hostname"`
	IP            string `json:"ip"`
	AgentVersion  string `json:"agent-version"`
	PluginVersion string `json:"plugin-version"`
	LastSeen      int64  `json:"lastseen"`
}

type MsgType struct {
	Type string `json:"type"`
}

// Hear these from Riemann
type PluginDirInfo struct {
	// type = plugin-dir
	Hostname string   `json:"hostname"`
	Dirs     []string `json:"dirs"`
}

type PluginInfo struct {
	// type = plugin
	Hostname string              `json:"hostname"`
	Params   []model.PluginParam `json:"plugins"`
}

type PluginVersionInfo struct {
	// type = plugin-version
	Version string `json:"version"`
}

type MasterState struct {
	PluginVersion  string                         `json:"plugin-version"`
	ConfigVersions map[string]int64               `json:"config-version"`
	Agents         map[string]AgentInfo           `json:"agents"`
	PluginDirs     map[string][]string            `json:"plugin-dirs"`
	Plugin         map[string][]model.PluginParam `json:"plugins"`
}

var StateLock = new(sync.RWMutex)

var State = MasterState{
	PluginVersion:  "",
	ConfigVersions: make(map[string]int64),
	Agents:         make(map[string]AgentInfo),
	PluginDirs:     make(map[string][]string),
	Plugin:         make(map[string][]model.PluginParam),
}

func Start() {
	go receiveAgentStates()
	go purgeStaleNodes()
}
