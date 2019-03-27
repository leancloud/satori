package funcs

import (
	"log"
	"time"

	"github.com/gaochao1/sw"

	"github.com/leancloud/satori/common/model"
	"github.com/leancloud/satori/swcollector/config"
)

func PingMetrics(ch chan *model.MetricValue) {
	AliveIpLock.RLock()
	defer AliveIpLock.RUnlock()
	for ip := range AliveIp {
		go pingMetrics(ip, ch)
	}
}

func pingMetrics(ip string, ch chan *model.MetricValue) {
	cfg := config.Config()
	timeout := cfg.PingTimeout * cfg.PingRetry
	fastPingMode := cfg.FastPingMode
	rtt, err := sw.PingRtt(ip, timeout, fastPingMode)
	if err != nil {
		log.Println(ip, err)
		return
	}
	ch <- &model.MetricValue{
		Endpoint:  ReverseLookup(ip),
		Metric:    "switch.Ping",
		Value:     float64(rtt),
		Timestamp: time.Now().Unix(),
	}
}
