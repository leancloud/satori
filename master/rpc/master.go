package rpc

import (
	"github.com/leancloud/satori/common/model"
	"github.com/leancloud/satori/master/state"
	"time"
)

func (t *Agent) Heartbeat(req *model.AgentHeartbeatRequest, resp *model.AgentHeartbeatResponse) error {
	host := req.Hostname
	if host == "" {
		return nil
	}

	state.StateLock.Lock()
	state.State.Agents[host] = state.AgentInfo{
		Hostname:      host,
		IP:            req.IP,
		AgentVersion:  req.AgentVersion,
		PluginVersion: req.PluginVersion,
		LastSeen:      time.Now().Unix(),
	}
	state.StateLock.Unlock()

	state.StateLock.RLock()
	cfgVer := state.State.ConfigVersions[host]
	cfgModified := cfgVer != 0 && req.ConfigVersion != cfgVer

	resp.ConfigModified = cfgModified
	resp.PluginVersion = state.State.PluginVersion

	if cfgModified {
		resp.ConfigVersion = cfgVer
		resp.PluginDirs = state.State.PluginDirs[host]
		resp.PluginMetrics = state.State.PluginMetrics[host]
	}
	state.StateLock.RUnlock()
	return nil
}
