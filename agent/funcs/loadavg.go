package funcs

import (
	"github.com/leancloud/satori/common/model"
	"github.com/toolkits/nux"
	"log"
)

func LoadAvgMetrics() []*model.MetricValue {
	load, err := nux.LoadAvg()
	if err != nil {
		log.Println(err)
		return nil
	}

	return []*model.MetricValue{
		V("load.1min", load.Avg1min),
		V("load.5min", load.Avg5min),
		V("load.15min", load.Avg15min),
	}

}
