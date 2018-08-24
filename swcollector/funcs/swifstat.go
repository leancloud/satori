package funcs

import (
	"log"
	"strconv"
	"time"

	"github.com/gaochao1/sw"

	"github.com/leancloud/satori/common/model"
	"github.com/leancloud/satori/swcollector/config"
	"github.com/leancloud/satori/swcollector/logging"
)

var (
	initialized = false
	ignores     = map[string]bool{}
	allIps      = []string{}
)

func SwIfMetrics(ch chan *model.MetricValue) {
	cfg := config.Config()
	if len(cfg.IpRange) <= 0 {
		logging.Fatalln("No ipRange configured, aborting")
	}

	if !initialized {
		initialized = true
		for _, v := range cfg.Ignore {
			ignores[v] = true
		}
		allIps = ExpandIpRanges(cfg.IpRange)
	}

	sem := make(chan bool, cfg.ConcurrentCollectors)
	for _, ip := range allIps {
		sem <- true
		go coreSwIfMetrics(ip, ch, sem)
		time.Sleep(5 * time.Millisecond)
	}
}

func coreSwIfMetrics(ip string, ch chan *model.MetricValue, sem chan bool) {
	log.Println("Collect coreSwIfMetrics for", ip)
	startTime := time.Now().UnixNano()
	host := ReverseLookup(ip)
	cfg := config.Config()

	NM := func(m string, v float64) {
		ch <- &model.MetricValue{
			Endpoint:  host,
			Metric:    m,
			Value:     v,
			Timestamp: time.Now().Unix(),
		}
	}

	pingResult := false

	for i := 0; i < cfg.PingRetry; i++ {
		pingResult = sw.Ping(ip, cfg.PingTimeout, cfg.FastPingMode)
		if pingResult == true {
			break
		}
	}

	if !pingResult {
		<-sem
		return
	}

	AliveIpLock.Lock()
	if !AliveIp[ip] {
		AliveIp[ip] = true
		log.Println("Found alive IP:", ip)
	}
	AliveIpLock.Unlock()

	var ifList []sw.IfStats
	var err error

	if cfg.Gosnmp {
		ifList, err = sw.ListIfStats(
			ip,
			cfg.SnmpCommunity, cfg.SnmpTimeout, cfg.IgnoreIface, cfg.SnmpRetry,
			cfg.ConcurrentQueriesPerHost,
			ignores["packets"],
			ignores["operstatus"],
			ignores["broadcasts"],
			ignores["multicasts"],
			ignores["discards"],
			ignores["errors"],
			ignores["unknownprotos"],
			ignores["qlen"],
		)
	} else {
		ifList, err = sw.ListIfStatsSnmpWalk(
			ip,
			cfg.SnmpCommunity, cfg.SnmpTimeout*5, cfg.IgnoreIface, cfg.SnmpRetry,
			ignores["packets"],
			ignores["operstatus"],
			ignores["broadcasts"],
			ignores["multicasts"],
			ignores["discards"],
			ignores["errors"],
			ignores["unknownprotos"],
			ignores["qlen"],
		)
	}

	<-sem

	if err != nil {
		log.Printf(ip, err)
		NM("switch.CollectTime", -1.0)
	} else {
		NM("switch.CollectTime", float64(time.Now().UnixNano()-startTime)/float64(time.Millisecond))
	}

	for _, ifStat := range ifList {
		ts := ifStat.TS
		tags := map[string]string{
			"ifName":  ifStat.IfName,
			"ifIndex": strconv.Itoa(ifStat.IfIndex),
		}
		M := func(m string, v float64, ignore string) {
			if !ignores[ignore] {
				ch <- &model.MetricValue{
					Endpoint:  ip,
					Metric:    m,
					Value:     v,
					Timestamp: ts,
					Tags:      tags,
				}
			}
		}
		M("switch.if.OperStatus", float64(ifStat.IfOperStatus), "operstatus")
		M("switch.if.Speed", float64(ifStat.IfSpeed), "speed")
		M("switch.if.InBroadcastPkt", float64(ifStat.IfHCInBroadcastPkts), "broadcasts")
		M("switch.if.OutBroadcastPkt", float64(ifStat.IfHCOutBroadcastPkts), "broadcasts")
		M("switch.if.InMulticastPkt", float64(ifStat.IfHCInMulticastPkts), "multicasts")
		M("switch.if.OutMulticastPkt", float64(ifStat.IfHCOutMulticastPkts), "multicasts")
		M("switch.if.InDiscards", float64(ifStat.IfInDiscards), "discards")
		M("switch.if.OutDiscards", float64(ifStat.IfOutDiscards), "discards")
		M("switch.if.InErrors", float64(ifStat.IfInErrors), "errors")
		M("switch.if.OutErrors", float64(ifStat.IfOutErrors), "errors")
		M("switch.if.InUnknownProtos", float64(ifStat.IfInUnknownProtos), "unknownprotos")
		M("switch.if.OutQLen", float64(ifStat.IfOutQLen), "qlen")
		M("switch.if.InPkts", float64(ifStat.IfHCInUcastPkts), "packets")
		M("switch.if.OutPkts", float64(ifStat.IfHCOutUcastPkts), "packets")
		M("switch.if.In", float64(ifStat.IfHCInOctets), "octets")
		M("switch.if.Out", float64(ifStat.IfHCOutOctets), "octets")
	}
}
