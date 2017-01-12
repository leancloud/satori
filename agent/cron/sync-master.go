package cron

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"

	"github.com/leancloud/satori/agent/g"
	"github.com/leancloud/satori/agent/plugins"
	"github.com/leancloud/satori/common/cpool"
	"github.com/leancloud/satori/common/model"
)

type HeartbeatClient struct {
	cli  *rpc.Client
	name string
}

func (this HeartbeatClient) Name() string {
	return this.name
}

func (this HeartbeatClient) Closed() bool {
	return this.cli == nil
}

func (this HeartbeatClient) Close() error {
	if this.cli == nil {
		this.cli.Close()
		this.cli = nil
	}
	return nil
}

func (this HeartbeatClient) Call(req interface{}) (interface{}, error) {
	var resp model.AgentHeartbeatResponse
	err := this.cli.Call("Agent.Heartbeat", req, &resp)
	if err != nil {
		log.Println("call Agent.Heartbeat fail:", err, "Request:", req, "Response:", resp)
	}
	return resp, nil
}

func heartbeatConnect(name string, p *cpool.ConnPool) (cpool.NConn, error) {
	connTimeout := time.Duration(p.ConnTimeout) * time.Millisecond
	conn, err := net.DialTimeout("tcp", p.Address, connTimeout)

	if err != nil {
		log.Printf("Connect master %s fail: %v", p.Address, err)
		return nil, err
	}

	return HeartbeatClient{
		cli:  jsonrpc.NewClient(conn),
		name: name,
	}, nil
}

func SyncWithMaster() {
	cfg := g.Config().Master
	debug := g.Config().Debug

	if !cfg.Enabled || cfg.Addr == "" {
		log.Println("Heartbeat not configured, plugins and certain metrics are not usable.")
		return
	}

	cli := cpool.NewConnPool(
		"master", cfg.Addr, 5, 3, cfg.Timeout, cfg.Timeout, heartbeatConnect,
	)

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
		r, err := cli.Call(req)
		if err != nil {
			log.Println("call Agent.Heartbeat fail:", err, "Request:", req, "Response:", resp)
			time.Sleep(interval * 3)
			continue
		}

		resp = r.(model.AgentHeartbeatResponse)
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
