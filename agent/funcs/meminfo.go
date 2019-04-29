package funcs

import (
	"github.com/leancloud/satori/common/model"
	"github.com/toolkits/nux"
	"log"
)

func MemMetrics() []*model.MetricValue {
	m, err := nux.MemInfo()
	if err != nil {
		log.Println(err)
		return nil
	}

	memUsable := m.MemFree + m.Buffers + m.Cached
	memUsed := m.MemTotal - memUsable

	pmemFree := 0.0
	pmemUsed := 0.0
	if m.MemTotal != 0 {
		pmemFree = float64(memUsable) * 100.0 / float64(m.MemTotal)
		pmemUsed = float64(memUsed) * 100.0 / float64(m.MemTotal)
	}

	pswapFree := 0.0
	pswapUsed := 0.0
	if m.SwapTotal != 0 {
		pswapFree = float64(m.SwapFree) * 100.0 / float64(m.SwapTotal)
		pswapUsed = float64(m.SwapUsed) * 100.0 / float64(m.SwapTotal)
	}

	return []*model.MetricValue{
		V("mem.free", float64(m.MemFree)),
		V("mem.buffers", float64(m.Buffers)),
		V("mem.cached", float64(m.Cached)),
		V("mem.memtotal", float64(m.MemTotal)),
		V("mem.memused", float64(memUsed)),
		V("mem.memusable", float64(memUsable)),
		V("mem.swaptotal", float64(m.SwapTotal)),
		V("mem.swapused", float64(m.SwapUsed)),
		V("mem.swapfree", float64(m.SwapFree)),
		V("mem.memusable.percent", float64(pmemFree)),
		V("mem.memused.percent", float64(pmemUsed)),
		V("mem.swapfree.percent", float64(pswapFree)),
		V("mem.swapused.percent", float64(pswapUsed)),
	}

}
