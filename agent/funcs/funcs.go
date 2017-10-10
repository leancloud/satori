package funcs

import (
	"log"

	"github.com/leancloud/satori/agent/g"
	"github.com/leancloud/satori/common/model"
)

type FuncsAndInterval struct {
	Fs       []func() []*model.MetricValue
	Interval int
}

var Mappers []FuncsAndInterval

func BuildMappers() {
	if g.Config().NoBuiltin {
		log.Println("No builtin specified, enable only `agent.alive`")
		Mappers = []FuncsAndInterval{
			FuncsAndInterval{
				Fs: []func() []*model.MetricValue{
					AgentMetrics,
				},
				Interval: 60,
			},
		}
	} else {
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
				Interval: 60,
			},
		}
	}
}
