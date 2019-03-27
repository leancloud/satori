package funcs

import (
	"log"
	"time"

	"github.com/gaochao1/sw"

	"github.com/leancloud/satori/common/model"
	"github.com/leancloud/satori/swcollector/config"
)

func CpuMemMetrics(ch chan *model.MetricValue) {
	AliveIpLock.RLock()
	defer AliveIpLock.RUnlock()
	for ip := range AliveIp {
		go cpuMetrics(ip, ch)
		go memMetrics(ip, ch)
	}
}

func cpuMetrics(ip string, ch chan *model.MetricValue) {
	cfg := config.Config()
	util, err := sw.CpuUtilization(ip, cfg.SnmpCommunity, cfg.SnmpTimeout, cfg.SnmpRetry)
	if err != nil {
		log.Println("Error collecting cpuMetrics:", err)
		return
	}
	ch <- &model.MetricValue{
		Endpoint:  ReverseLookup(ip),
		Metric:    "switch.CpuUtilization",
		Value:     float64(util),
		Timestamp: time.Now().Unix(),
	}
}

func memMetrics(ip string, ch chan *model.MetricValue) {
	cfg := config.Config()
	util, err := sw.MemUtilization(ip, cfg.SnmpCommunity, cfg.SnmpTimeout, cfg.SnmpRetry)
	if err != nil {
		log.Println("Error collecting memMetrics:", err)
		return
	}
	ch <- &model.MetricValue{
		Endpoint:  ReverseLookup(ip),
		Metric:    "switch.MemUtilization",
		Value:     float64(util),
		Timestamp: time.Now().Unix(),
	}
}
