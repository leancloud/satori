package cron

import (
	"fmt"
	"github.com/leancloud/satori/agent/g"
	"github.com/leancloud/satori/agent/plugins"
	"github.com/leancloud/satori/agent/rpc"
	"github.com/leancloud/satori/common/model"
	"log"
	"time"
)

func SyncWithMaster() {
	cfg := g.Config().Master
	debug := g.Config().Debug

	if !cfg.Enabled || cfg.Addr == "" {
		log.Println("Heartbeat not configured, plugins and certain metrics are not usable.")
		return
	}

	cli := &rpc.RpcClient{
		RpcServer: cfg.Addr,
		Timeout:   time.Duration(cfg.Timeout) * time.Millisecond,
	}

	interval := time.Duration(cfg.Interval) * time.Second
	for {
		hostname, err := g.Hostname()
		if err != nil {
			hostname = fmt.Sprintf("error:%s", err.Error())
		}

		ver, err := plugins.GetCurrentPluginVersion()
		if err != nil {
			ver = err.Error()
		}

		req := model.AgentHeartbeatRequest{
			Hostname:      hostname,
			IP:            g.IP(),
			AgentVersion:  g.VERSION,
			PluginVersion: ver,
			ConfigVersion: g.ConfigVersion,
		}

		var resp model.AgentHeartbeatResponse
		err = cli.Call("Agent.Heartbeat", req, &resp)
		if err != nil {
			log.Println("call Agent.Heartbeat fail:", err, "Request:", req, "Response:", resp)
			time.Sleep(interval * 3)
			continue
		}

		if debug {
			log.Printf("Response from master: %s", resp)
		}

		if resp.ConfigModified {
			g.ConfigVersion = resp.ConfigVersion
			// and PluginDirs & PluginMetrics non-null
		}

		go plugins.SyncConfig(resp.PluginVersion, resp.PluginDirs, resp.PluginMetrics)

		time.Sleep(interval)
	}
}
