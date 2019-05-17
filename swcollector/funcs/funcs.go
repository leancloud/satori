package funcs

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/leancloud/satori/common/model"
	"github.com/leancloud/satori/swcollector/config"
)

var Collectors = []func(chan *model.MetricValue){
	SwIfMetrics,
	CpuMemMetrics,
	PingMetrics,
	CustomMetrics,
}

func StartCollect() {
	ch := make(chan *model.MetricValue, 30)
	go doCollect(ch)
	go doOutput(ch)
}

func doCollect(ch chan *model.MetricValue) {
	step := config.Config().Interval

	for {
		log.Println("doCollect round started")
		for _, f := range Collectors {
			go f(ch)
		}
		time.Sleep(time.Duration(step) * time.Second)
	}
}

func doOutput(ch chan *model.MetricValue) {
	emptyTags := map[string]string{}

	cook := func(m *model.MetricValue) *model.MetricValue {
		if m.Tags == nil {
			m.Tags = emptyTags
		}
		return m
	}

	for {
		L := []*model.MetricValue{}
		v := cook(<-ch)
		L = append(L, v)
		for i := 0; i < len(ch); i++ {
			L = append(L, cook(<-ch))
		}
		s, _ := json.Marshal(L)
		os.Stdout.Write(s)
		os.Stdout.WriteString("\n")
	}
}
