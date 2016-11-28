package funcs

import (
	"github.com/leancloud/satori/common/model"
)

func VT(metric string, val float64, tags map[string]string) *model.MetricValue {
	if tags == nil {
		tags = map[string]string{}
	}

	mv := model.MetricValue{
		Metric: metric,
		Value:  val,
		Tags:   tags,
	}
	return &mv
}

func V(metric string, val float64) *model.MetricValue {
	return VT(metric, val, nil)
}
