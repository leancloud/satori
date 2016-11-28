package sender

import (
	cmodel "github.com/leancloud/satori/common/model"
	"github.com/leancloud/satori/transfer/g"
	"log"
	"time"
)

// 默认参数
var (
	MinStep int //最小上报周期,单位sec
)

var Backends = []*Backend{
	&TsdbBackend,
	&InfluxdbBackend,
	&RiemannBackend,
}

// 连接池
// 初始化数据发送服务, 在main函数中调用
func Start() {
	// 初始化默认参数
	MinStep = g.Config().MinStep
	if MinStep < 1 {
		MinStep = 30 //默认30s
	}

	avail := make([]*Backend, 0, 5)
	for _, b := range Backends {
		err := b.Start()
		if err == nil {
			log.Printf("Started backend %s\n", b.Name)
			avail = append(avail, b)
		} else {
			log.Printf("Backend %s not started: %s\n", b.Name, err)
		}
	}

	Backends = avail

	go periodicallyPrintBackendStats()
	log.Println("send.Start, ok")
}

func Send(items []*cmodel.MetricValue) {
	for _, b := range Backends {
		b.Send(items)
	}
}

func periodicallyPrintBackendStats() {
	for {
		time.Sleep(DefaultLogCronPeriod)
		log.Println(">>>>>------------------------------------")
		for _, b := range Backends {
			stats := b.GetStats()
			log.Printf("Backend %s: Send: %d, Drop: %d, Fail: %d, QueueLength: %d\n",
				b.Name,
				stats.SendCounter.Qps,
				stats.DropCounter.Qps,
				stats.FailCounter.Qps,
				stats.QueueLength.Cnt,
			)
			for _, s := range stats.ConnPoolStats {
				log.Printf("    - ConnPool: %s\n", s)
			}
			log.Printf("\n")

		}
	}
}
