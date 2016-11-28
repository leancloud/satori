package funcs

import (
	"github.com/leancloud/satori/common/model"
	"github.com/toolkits/nux"
	"log"
)

func SocketStatSummaryMetrics() (L []*model.MetricValue) {
	ssMap, err := nux.SocketStatSummary()
	if err != nil {
		log.Println(err)
		return
	}

	for k, v := range ssMap {
		L = append(L, V("ss."+k, float64(v)))
	}

	return
}
