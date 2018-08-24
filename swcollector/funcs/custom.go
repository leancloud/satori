package funcs

import (
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	go_snmp "github.com/gaochao1/gosnmp"
	"github.com/gaochao1/sw"

	"github.com/leancloud/satori/common/model"
	"github.com/leancloud/satori/swcollector/config"
)

var (
	customMetricIp = map[string]map[string]bool{}
)

func CustomMetrics(ch chan *model.MetricValue) {
	customMetrics := config.Config().CustomMetrics
	if len(customMetrics) <= 0 {
		return
	}
	AliveIpLock.RLock()
	defer AliveIpLock.RUnlock()
	for _, metric := range customMetrics {
		key := strings.Join(metric.IpRange, "|")
		ips := customMetricIp[key]
		if ips == nil {
			ipsSlice := ExpandIpRanges(metric.IpRange)
			ips = map[string]bool{}
			for _, v := range ipsSlice {
				ips[v] = true
			}
			customMetricIp[key] = ips
		}
		for ip := range AliveIp {
			if ips[ip] {
				go collectMetric(ip, &metric, ch)
			}
		}
	}
}

func collectMetric(ip string, metric *config.CustomMetric, ch chan *model.MetricValue) {
	cfg := config.Config()
	log.Println("Collect custom metric", metric.Metric, "for", ip)
	value, err := snmpGet(ip, cfg.SnmpCommunity, metric.Oid, cfg.SnmpTimeout, cfg.SnmpRetry)

	if err != nil {
		log.Println("Error collectiing custom metric", metric.Metric, ip, metric.Oid, err)
		return
	}

	ch <- &model.MetricValue{
		Endpoint:  ReverseLookup(ip),
		Metric:    metric.Metric,
		Value:     value,
		Timestamp: time.Now().Unix(),
		Tags:      metric.Tags,
	}
}

func snmpGet(ip, community, oid string, timeout, retry int) (float64, error) {
	defer func() {
		if r := recover(); r != nil {
			log.Println(ip, "recovered in CustomMetric, Oid is", oid, r)
		}
	}()
	method := "get"
	var value float64
	var err error
	var snmpPDUs []go_snmp.SnmpPDU
	for i := 0; i < retry; i++ {
		snmpPDUs, err = sw.RunSnmp(ip, community, oid, method, timeout)
		if len(snmpPDUs) > 0 && err == nil {
			value, err = interfaceTofloat64(snmpPDUs[0].Value)
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	return value, err
}

func interfaceTofloat64(v interface{}) (float64, error) {
	var err error
	switch value := v.(type) {
	case int:
		return float64(value), nil
	case int8:
		return float64(value), nil
	case int16:
		return float64(value), nil
	case int32:
		return float64(value), nil
	case int64:
		return float64(value), nil
	case uint:
		return float64(value), nil
	case uint8:
		return float64(value), nil
	case uint16:
		return float64(value), nil
	case uint32:
		return float64(value), nil
	case uint64:
		return float64(value), nil
	case float32:
		return float64(value), nil
	case float64:
		return value, nil
	case string:
		value_parsed, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, err
		} else {
			return value_parsed, nil
		}
	default:
		err = errors.New("value cannot not Parse to digital")
		return 0, err
	}
}
