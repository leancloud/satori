package funcs

import (
	"github.com/leancloud/satori/agent/g"
	"github.com/leancloud/satori/common/model"
)

type FuncsAndInterval struct {
	Fs       []func() []*model.MetricValue
	Interval int
}

var Mappers []FuncsAndInterval

func BuildMappers() {
	interval := g.Config().Transfer.Interval
	Mappers = []FuncsAndInterval{
		FuncsAndInterval{
			Fs: []func() []*model.MetricValue{
				AgentMetrics,
				CpuMetrics,
				NetMetrics,
				KernelMetrics,
				LoadAvgMetrics,
				MemMetrics,
				DiskIOMetrics,
				IOStatsMetrics,
				NetstatMetrics,
				UdpMetrics,

				DeviceMetrics,
				SocketStatSummaryMetrics,
			},
			Interval: interval,
		},
	}
}
