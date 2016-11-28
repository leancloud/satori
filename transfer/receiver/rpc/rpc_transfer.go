package rpc

import (
	"fmt"
	cmodel "github.com/leancloud/satori/common/model"
	"github.com/leancloud/satori/transfer/receiver/core"
)

type Transfer int

type TransferResp struct {
	Msg        string
	Total      int
	ErrInvalid int
	Latency    int64
}

func (t *TransferResp) String() string {
	s := fmt.Sprintf("TransferResp total=%d, err_invalid=%d, latency=%dms",
		t.Total, t.ErrInvalid, t.Latency)
	if t.Msg != "" {
		s = fmt.Sprintf("%s, msg=%s", s, t.Msg)
	}
	return s
}

func (this *Transfer) Ping(req cmodel.NullRpcRequest, resp *cmodel.SimpleRpcResponse) error {
	return nil
}

func (t *Transfer) Update(args []*cmodel.MetricValue, reply *cmodel.TransferResponse) error {
	return core.RecvMetricValues(args, reply)
}
