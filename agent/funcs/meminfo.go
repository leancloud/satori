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

	memFree := m.MemFree + m.Buffers + m.Cached
	memUsed := m.MemTotal - memFree

	pmemFree := 0.0
	pmemUsed := 0.0
	if m.MemTotal != 0 {
		pmemFree = float64(memFree) * 100.0 / float64(m.MemTotal)
		pmemUsed = float64(memUsed) * 100.0 / float64(m.MemTotal)
	}

	pswapFree := 0.0
	pswapUsed := 0.0
	if m.SwapTotal != 0 {
		pswapFree = float64(m.SwapFree) * 100.0 / float64(m.SwapTotal)
		pswapUsed = float64(m.SwapUsed) * 100.0 / float64(m.SwapTotal)
	}

	return []*model.MetricValue{
		V("mem.memtotal", float64(m.MemTotal)),
		V("mem.memused", float64(memUsed)),
		V("mem.memfree", float64(memFree)),
		V("mem.swaptotal", float64(m.SwapTotal)),
		V("mem.swapused", float64(m.SwapUsed)),
		V("mem.swapfree", float64(m.SwapFree)),
		V("mem.memfree.percent", float64(pmemFree)),
		V("mem.memused.percent", float64(pmemUsed)),
		V("mem.swapfree.percent", float64(pswapFree)),
		V("mem.swapused.percent", float64(pswapUsed)),
	}

}
