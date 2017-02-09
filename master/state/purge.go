package state

import (
	"log"
	"time"

	"github.com/leancloud/satori/master/g"
)

func purgeStaleNodes() {
	cfg := g.Config()
	deadline := cfg.PurgeSeconds
	if deadline <= 0 {
		log.Println("PurgeSeconds <= 0, not purging stale nodes")
		return
	}

	for {
		time.Sleep(time.Second * 10)
		now := time.Now().Unix()
		purging := []string{}

		StateLock.Lock()
		for k, v := range State.Agents {
			if now-v.LastSeen > deadline {
				purging = append(purging, k)
			}
		}

		for _, k := range purging {
			delete(State.Agents, k)
		}
		StateLock.Unlock()
	}
}
