package core

import (
	"time"

	nproc "github.com/toolkits/proc"

	cmodel "github.com/leancloud/satori/common/model"
	"github.com/leancloud/satori/transfer/sender"
)

var (
	RecvCnt = nproc.NewSCounterQps("RecvCnt")

	RecvDataTrace  = nproc.NewDataTrace("RecvDataTrace", 3)
	RecvDataFilter = nproc.NewDataFilter("RecvDataFilter", 5)
)

// process new metric values
func RecvMetricValues(args []*cmodel.MetricValue, reply *cmodel.TransferResponse) error {
	start := time.Now()
	reply.Invalid = 0

	items := []*cmodel.MetricValue{}
	for _, v := range args {
		if v == nil {
			reply.Invalid += 1
			continue
		}

		if v.Metric == "" || v.Endpoint == "" {
			reply.Invalid += 1
			continue
		}

		if v.Step <= 0 {
			reply.Invalid += 1
			continue
		}

		if len(v.Metric)+len(v.Tags) > 510 {
			reply.Invalid += 1
			continue
		}

		// TODO 呵呵,这里需要再优雅一点
		now := start.Unix()
		if v.Timestamp <= 0 || v.Timestamp > now*2 {
			v.Timestamp = now
		}

		{
			pk := v.PK()
			RecvDataTrace.Trace(pk, v)
			RecvDataFilter.Filter(pk, v.Value, v)

		}

		items = append(items, v)
	}

	// statistics
	cnt := int64(len(items))
	RecvCnt.IncrBy(cnt)

	sender.Send(items)

	reply.Message = "ok"
	reply.Total = len(args)
	reply.Latency = (time.Now().UnixNano() - start.UnixNano()) / 1000000

	return nil
}
