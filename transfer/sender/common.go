package sender

import (
	cmodel "github.com/leancloud/satori/common/model"
	"github.com/leancloud/satori/transfer/sender/conn_pool"
	"github.com/toolkits/consistent"
	nproc "github.com/toolkits/proc"
	"time"
)

const (
	DefaultLogCronPeriod         = time.Duration(180) * time.Second //LogCron的周期,默认 180s
	DefaultSendQueueMaxSize      = 10240
	DefaultSendTaskSleepInterval = time.Millisecond * 50 //默认睡眠间隔为50ms
)

type BackendStats struct {
	SendCounter   *nproc.SCounterQps
	DropCounter   *nproc.SCounterQps
	FailCounter   *nproc.SCounterQps
	QueueLength   *nproc.SCounterBase
	ConnPoolStats []*conn_pool.ConnPoolStats
}

type Backend struct {
	Name     string
	Start    func() error
	Send     func(items []*cmodel.MetricValue)
	GetStats func() *BackendStats
}

func newConsistentHashNodesRing(numberOfReplicas int, nodeConf map[string]string) *consistent.Consistent {
	ring := consistent.New()
	ring.NumberOfReplicas = numberOfReplicas
	for k, _ := range nodeConf {
		ring.Add(k)
	}
	return ring
}
