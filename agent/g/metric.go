package g

import (
	"log"
	"regexp"

	"github.com/leancloud/satori/common/model"
)

var cachedRegexp map[string]*regexp.Regexp = make(map[string]*regexp.Regexp)

func cachedMatch(re string, v string) bool {
	r, ok := cachedRegexp[re]
	if !ok {
		r = regexp.MustCompile(re)
		cachedRegexp[re] = r
	}
	return r.MatchString(v)
}

func filterMetrics(metrics []*model.MetricValue) []*model.MetricValue {
	cfg := Config()
	addTags := cfg.AddTags
	ignore := cfg.Ignore
	debug := cfg.Debug
	hostname, _ := Hostname()

	filtered := make([]*model.MetricValue, 0)

metricsLoop:
	for _, mv := range metrics {
		for _, item := range ignore {
			metricRe, tagKeyRe, tagRe := item[0], item[1], item[2]
			if !cachedMatch(metricRe, mv.Metric) {
				continue
			}

			for k, v := range mv.Tags {
				if cachedMatch(tagKeyRe, k) && cachedMatch(tagRe, v) {
					if debug {
						log.Println("=> Filtered metric", mv.Metric, "/", mv.Tags, "by rule ", metricRe, " :: ", tagKeyRe, " :: ", tagRe)
					}
					continue metricsLoop
				}
			}
		}

		if mv.Tags == nil {
			mv.Tags = map[string]string{}
		}

		if addTags != nil {
			for k, v := range addTags {
				if _, ok := mv.Tags[k]; !ok {
					mv.Tags[k] = v
				}
			}
		}

		if mv.Endpoint == "" {
			mv.Endpoint = hostname
		}

		filtered = append(filtered, mv)
	}
	return filtered
}
