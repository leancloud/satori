package cron

import (
	"log"
	"time"

	"github.com/leancloud/satori/agent/funcs"
	"github.com/leancloud/satori/agent/g"
	"github.com/leancloud/satori/common/model"
)

func StartCollect() {
	if len(g.Config().Transfer.Addrs) == 0 {
		log.Fatalln("Transfer not configured!")
	}

	go collectCpuDisks()
	for _, v := range funcs.Mappers {
		go collect(int64(v.Interval), v.Fs)
	}
}

func collectCpuDisks() {
	for {
		funcs.UpdateCpuStat()
		funcs.UpdateDiskStats()
		time.Sleep(time.Second)
	}
}

func collect(sec int64, fns []func() []*model.MetricValue) {
	t := time.NewTicker(time.Second * time.Duration(sec)).C
	for {
		<-t

		hostname, err := g.Hostname()
		if err != nil {
			continue
		}

		mvs := []*model.MetricValue{}
		debug := g.Config().Debug

		for _, fn := range fns {
			items := fn()
			if items == nil {
				continue
			}

			if len(items) == 0 {
				continue
			}

			if debug {
				log.Println(" -> collect ", len(items), " metrics\n")
			}
			mvs = append(mvs, items...)
		}

		now := time.Now().Unix()
		for j := 0; j < len(mvs); j++ {
			mvs[j].Step = sec
			mvs[j].Endpoint = hostname
			mvs[j].Timestamp = now
		}

		g.SendToTransfer(mvs)
	}
}
