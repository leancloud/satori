package funcs

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gaochao1/sw"
	"github.com/miekg/dns"

	"github.com/leancloud/satori/swcollector/config"
	"github.com/leancloud/satori/swcollector/logging"
)

var (
	reverseLookups    = map[string]string{}
	reverseLookupLock = sync.RWMutex{}
	dnsServer         string

	AliveIp     = map[string]bool{}
	AliveIpLock = sync.RWMutex{}
)

func init() {
	dnsConfig, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	if err != nil {
		logging.Fatalln("Failed to parse /etc/resolv.conf:", err)
	}
	dnsServer = dnsConfig.Servers[0] + ":" + dnsConfig.Port
}

func ExpandIpRanges(ipRange []string) []string {
	allIp := []string{}
	if len(ipRange) > 0 {
		for _, sip := range ipRange {
			aip := sw.ParseIp(sip)
			for _, ip := range aip {
				allIp = append(allIp, ip)
			}
		}
	}
	return allIp
}

func ReverseLookup(ip string) string {
	cfg := config.Config()

	if h := cfg.CustomHosts[ip]; h != "" {
		return h
	}

	if !cfg.ReverseLookup {
		return ip
	}

	reverseLookupLock.RLock()
	h := reverseLookups[ip]
	reverseLookupLock.RUnlock()

	if h != "" {
		return h
	}

	c := dns.Client{Timeout: 1 * time.Second}

	arpa, err := dns.ReverseAddr(ip)
	if err != nil {
		log.Println("ReverseLookup got invalid ip:", ip, ":", err)
		return ip
	}

	var host string
	for i := 0; i < 3; i++ {
		m := dns.Msg{}
		m.RecursionDesired = true
		m.SetQuestion(arpa, dns.TypePTR)
		r, _, err := c.Exchange(&m, dnsServer)
		if err != nil {
			log.Println("Failed to perform DNS query for", ip, ":", err)
			host = ip
		} else {
			host = r.Answer[0].(*dns.PTR).Ptr
			host = strings.TrimSuffix(host, ".")
		}
	}

	reverseLookupLock.Lock()
	reverseLookups[ip] = host
	reverseLookupLock.Unlock()

	return host
}
