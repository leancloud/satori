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

var Backends = make([]Backend, 0, 5)

var BackendConstructors = []func(cfg *BackendConfig) Backend{
	newRiemannBackend,
	newInfluxdbBackend,
	newTsdbBackend,
	newTransferBackend,
}

// 连接池
// 初始化数据发送服务, 在main函数中调用
func Start() {
	// 初始化默认参数
	MinStep = g.Config().MinStep
	if MinStep < 1 {
		MinStep = 30 //默认30s
	}

	for _, s := range g.Config().Backends {
		cfg := parseBackendUrl(s)
		for _, f := range BackendConstructors {
			b := f(cfg)
			if b != nil {
				Backends = append(Backends, b)
				break
			}
		}
	}

	for _, b := range Backends {
		err := b.Start()
		if err == nil {
			log.Printf("Started backend %s\n", b.GetConfig().Name)
		} else {
			log.Printf("Backend %s not started: %s\n", b.GetConfig().Name, err)
		}
	}

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
		time.Sleep(120 * time.Second)
		log.Println(">>>>>------------------------------------")
		for _, b := range Backends {
			stats := b.GetStats()
			log.Printf("Backend %s: Send: %d, Drop: %d, Fail: %d, QueueLength: %d\n",
				b.GetConfig().Name,
				stats.Send.Qps,
				stats.Drop.Qps,
				stats.Fail.Qps,
				stats.QueueLength.Cnt,
			)
			for _, s := range stats.ConnPoolStats {
				log.Printf("    - ConnPool: %s\n", s)
			}
			log.Printf("\n")
		}
	}
}
