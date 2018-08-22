package funcs

import (
	"github.com/leancloud/satori/common/model"
)

func AgentMetrics() []*model.MetricValue {
	return []*model.MetricValue{V("agent.alive", 1)}
}

func ContainerAliveMetrics() []*model.MetricValue {
	return []*model.MetricValue{V("agent.container-alive", 1)}
}
