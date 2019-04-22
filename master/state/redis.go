package state

import (
	"encoding/json"
	"log"
	"net"
	"net/url"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/leancloud/satori/master/g"
)

func connect() net.Conn {
	cfg := g.Config()

	for {
		u, err := url.Parse(cfg.Redis)
		if err != nil {
			panic("Malformed redis url")
		}

		dialer := net.Dialer{
			Timeout:   time.Second * 30,
			KeepAlive: time.Second * 120,
		}

		tc, err := dialer.Dial("tcp", u.Host)
		if err != nil {
			log.Println("Can't connect redis:", err)
			time.Sleep(time.Second)
			continue
		}
		return tc
	}
}

func notifyRestart() {
	tc := connect()
	c := redis.NewConn(tc, time.Hour*24, time.Second*10)
	_, _ = c.Do("PUBLISH", "satori:component-started", "master")
	_ = c.Close()
}

func receiveAgentStates() {
	for {
		tc := connect()
		c := redis.NewConn(tc, time.Hour*24, time.Second*10)

		pubsub := redis.PubSubConn{c}
		_ = pubsub.Subscribe("satori:master-state")
		for {
			switch n := pubsub.Receive().(type) {
			case redis.Message:
				doRecvState(n.Data)
			case error:
				log.Printf("pubsub error: %v\n", n)
				time.Sleep(time.Second)
				break
			}
		}
	}
}

func doRecvState(raw []byte) {
	debug := g.Config().Debug

	t := MsgType{}
	if err := json.Unmarshal(raw, &t); err != nil {
		log.Println("Can't decode: %s", raw)
		return
	}

	StateLock.Lock()
	defer StateLock.Unlock()

	switch t.Type {
	case "plugin-dir":
		dir := PluginDirInfo{}
		if err := json.Unmarshal(raw, &dir); err != nil {
			log.Printf("Can't unmarshal plugin-dir info: %s", err)
			break
		}
		if debug {
			log.Printf("New PluginDir: %s", dir)
		}
		State.PluginDirs[dir.Hostname] = dir.Dirs
		State.ConfigVersions[dir.Hostname] = time.Now().Unix()

	case "plugin":
		m := PluginInfo{}
		if err := json.Unmarshal(raw, &m); err != nil {
			log.Printf("Can't unmarshal plugin info: %s", err)
			break
		}
		if debug {
			log.Printf("New PluginMetric: %s", m)
		}
		State.Plugin[m.Hostname] = m.Params
		State.ConfigVersions[m.Hostname] = time.Now().Unix()

	case "plugin-version":
		v := PluginVersionInfo{}
		if err := json.Unmarshal(raw, &v); err != nil {
			log.Printf("Can't unmarshal plugin-version info: %s", err)
			break
		}
		if v.Version != "" {
			if debug {
				log.Printf("New PluginVersion: %s\n", v.Version)
			}
			State.PluginVersion = v.Version
		}
	}
}
