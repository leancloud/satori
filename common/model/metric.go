package model

import (
	"fmt"
	MUtils "github.com/leancloud/satori/common/utils"
)

type MetricValue struct {
	Endpoint  string            `json:"endpoint"`
	Metric    string            `json:"metric"`
	Value     float64           `json:"value"`
	Step      int64             `json:"step"`
	Tags      map[string]string `json:"tags"`
	Desc      string            `json:"description"`
	Timestamp int64             `json:"timestamp"`
}

func (this *MetricValue) String() string {
	return fmt.Sprintf(
		"<Endpoint:%s, Metric:%s, Tags:%s, Desc: %s, Step:%d, Time:%d, Value:%v>",
		this.Endpoint,
		this.Metric,
		this.Tags,
		this.Desc,
		this.Step,
		this.Timestamp,
		this.Value,
	)
}

func (t *MetricValue) PK() string {
	return MUtils.PK(t.Endpoint, t.Metric, t.Tags)
}
