package model

type PluginParam map[string]interface{}

type AgentHeartbeatRequest struct {
	Hostname      string
	IP            string
	AgentVersion  string
	PluginVersion string
	ConfigVersion int64
}

type AgentHeartbeatResponse struct {
	ConfigModified bool
	ConfigVersion  int64
	PluginVersion  string
	PluginDirs     []string
	PluginMetrics  []PluginParam // [{"_metric": "net.port.listen", "_step": 60, "port": 6379}, ...]
}
