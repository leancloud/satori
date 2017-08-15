package cron

import (
	"fmt"
	"log"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"
	"strconv"
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
		return nil, err
	}
	return resp, nil
}

func heartbeatConnect(name string, p *cpool.ConnPool) (cpool.PoolClient, error) {
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
	debug := g.Config().Debug
	s := g.Config().Master
	if s == "" {
		log.Println("Master not configured, plugins and certain metrics are not usable.")
	}
	u, err := url.Parse(s)
	if err != nil {
		log.Fatalln("Error parsing master url:", err.Error())
	}
	q := u.Query()

	getInt := func(f string, def int) int {
		var v string
		if v = q.Get(f); v == "" {
			return def
		}
		if intv, err := strconv.ParseInt(v, 10, 32); err == nil {
			return int(intv)
		} else {
			return def
		}
	}
	timeout := getInt("timeout", 5000)
	interval := time.Duration(getInt("interval", 60)) * time.Second

	cli := cpool.NewConnPool(
		"master", u.Host, 5, 3, timeout, timeout, heartbeatConnect,
	)

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
