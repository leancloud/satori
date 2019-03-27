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

func pingRedis(c redis.Conn, t time.Time) error {
	_, err := c.Do("ping")
	if err != nil {
		log.Println("[ERROR] ping redis fail", err)
	}
	return err
}

func receiveAgentStates() {
	cfg := g.Config()

	for {
		time.Sleep(time.Second)

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
			continue
		}
		c := redis.NewConn(tc, time.Hour*24, time.Second*10)
		pubsub := redis.PubSubConn{c}
		pubsub.Subscribe("satori:master-state")
		for {
			switch n := pubsub.Receive().(type) {
			case redis.Message:
				doRecvState(n.Data)
			case error:
				log.Printf("pubsub error: %v\n", n)
				goto duh
			}
		}
	duh:
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
		json.Unmarshal(raw, &dir)
		if debug {
			log.Printf("New PluginDir: %s", dir)
		}
		State.PluginDirs[dir.Hostname] = dir.Dirs
		State.ConfigVersions[dir.Hostname] = time.Now().Unix()

	case "plugin":
		m := PluginInfo{}
		json.Unmarshal(raw, &m)
		if debug {
			log.Printf("New PluginMetric: %s", m)
		}
		State.Plugin[m.Hostname] = m.Params
		State.ConfigVersions[m.Hostname] = time.Now().Unix()

	case "plugin-version":
		v := PluginVersionInfo{}
		json.Unmarshal(raw, &v)
		if v.Version != "" {
			if debug {
				log.Printf("New PluginVersion: %s\n", v.Version)
			}
			State.PluginVersion = v.Version
		}
	}
}
