package funcs

import (
	"github.com/leancloud/satori/agent/g"
	"github.com/leancloud/satori/common/model"
	"github.com/toolkits/nux"
	"log"
)

func NetMetrics() []*model.MetricValue {
	return CoreNetMetrics(g.Config().Collector.IfacePrefix)
}

func CoreNetMetrics(ifacePrefix []string) []*model.MetricValue {

	netIfs, err := nux.NetIfs(ifacePrefix)
	if err != nil {
		log.Println(err)
		return []*model.MetricValue{}
	}

	cnt := len(netIfs)
	ret := make([]*model.MetricValue, cnt*20)

	for idx, netIf := range netIfs {
		iface := map[string]string{
			"iface": netIf.Iface,
		}
		ret[idx*20+0] = VT("net.if.in.bytes", float64(netIf.InBytes), iface)
		ret[idx*20+1] = VT("net.if.in.packets", float64(netIf.InPackages), iface)
		ret[idx*20+2] = VT("net.if.in.errors", float64(netIf.InErrors), iface)
		ret[idx*20+3] = VT("net.if.in.dropped", float64(netIf.InDropped), iface)
		ret[idx*20+4] = VT("net.if.in.fifo.errs", float64(netIf.InFifoErrs), iface)
		ret[idx*20+5] = VT("net.if.in.frame.errs", float64(netIf.InFrameErrs), iface)
		ret[idx*20+6] = VT("net.if.in.compressed", float64(netIf.InCompressed), iface)
		ret[idx*20+7] = VT("net.if.in.multicast", float64(netIf.InMulticast), iface)
		ret[idx*20+8] = VT("net.if.out.bytes", float64(netIf.OutBytes), iface)
		ret[idx*20+9] = VT("net.if.out.packets", float64(netIf.OutPackages), iface)
		ret[idx*20+10] = VT("net.if.out.errors", float64(netIf.OutErrors), iface)
		ret[idx*20+11] = VT("net.if.out.dropped", float64(netIf.OutDropped), iface)
		ret[idx*20+12] = VT("net.if.out.fifo.errs", float64(netIf.OutFifoErrs), iface)
		ret[idx*20+13] = VT("net.if.out.collisions", float64(netIf.OutCollisions), iface)
		ret[idx*20+14] = VT("net.if.out.carrier.errs", float64(netIf.OutCarrierErrs), iface)
		ret[idx*20+15] = VT("net.if.out.compressed", float64(netIf.OutCompressed), iface)
		ret[idx*20+16] = VT("net.if.total.bytes", float64(netIf.TotalBytes), iface)
		ret[idx*20+17] = VT("net.if.total.packets", float64(netIf.TotalPackages), iface)
		ret[idx*20+18] = VT("net.if.total.errors", float64(netIf.TotalErrors), iface)
		ret[idx*20+19] = VT("net.if.total.dropped", float64(netIf.TotalDropped), iface)
	}
	return ret
}
