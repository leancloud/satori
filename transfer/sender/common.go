package sender

import (
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/leancloud/satori/common/cpool"
	cmodel "github.com/leancloud/satori/common/model"
	nlist "github.com/toolkits/container/list"
	nproc "github.com/toolkits/proc"
)

const (
	DefaultSendQueueMaxSize      = 10240
	DefaultSendTaskSleepInterval = time.Millisecond * 50 //默认睡眠间隔为50ms
)

type BackendStats struct {
	Send          *nproc.SCounterQps
	Drop          *nproc.SCounterQps
	Fail          *nproc.SCounterQps
	QueueLength   *nproc.SCounterBase
	ConnPoolStats []*cpool.ConnPoolStats
}

type BackendConfig struct {
	Name         string
	Engine       string
	Protocol     string
	ConnTimeout  int
	CallTimeout  int
	SendInterval int
	Batch        int
	MaxConn      int
	MaxIdle      int
	Retry        int
	Url          url.URL
}

func parseBackendUrl(connString string) *BackendConfig {
	u, err := url.Parse(connString)
	if err != nil {
		log.Printf("Can't parse backend string %s: %s\n", connString, err.Error())
		return nil
	}
	q := u.Query()
	getInt := func(f string, def int) int {
		v := q.Get(f)
		if v == "" {
			return def
		}
		intv, err := strconv.ParseInt(v, 10, 32)
		if err == nil {
			return int(intv)
		} else {
			return def
		}
	}

	name := q.Get("name")
	if name == "" {
		name = u.String()
	}

	v := strings.SplitN(u.Scheme, "+", 2)
	var engine, protocol string
	if len(v) > 1 {
		engine = v[0]
		protocol = v[1]
	} else {
		engine = v[0]
		protocol = ""
	}

	return &BackendConfig{
		Name:         name,
		Engine:       engine,
		Protocol:     protocol,
		ConnTimeout:  getInt("connTimeout", 1000),
		CallTimeout:  getInt("callTimeout", 5000),
		SendInterval: getInt("sendInterval", 1000),
		Batch:        getInt("batch", 20),
		MaxConn:      getInt("maxConn", 32),
		MaxIdle:      getInt("maxIdle", 32),
		Retry:        getInt("retry", 3),
		Url:          *u,
	}
}

type Backend interface {
	GetConfig() *BackendConfig
	Start() error
	Send(items []*cmodel.MetricValue)
	GetStats() *BackendStats
}

type BackendCommon struct {
	config *BackendConfig
	pool   *cpool.ConnPool
	queue  *nlist.SafeListLimited

	sendCounter *nproc.SCounterQps
	dropCounter *nproc.SCounterQps
	failCounter *nproc.SCounterQps
}

func (this *BackendCommon) GetConfig() *BackendConfig {
	return this.config
}

func (this *BackendCommon) GetStats() *BackendStats {
	ql := nproc.NewSCounterBase("QueueLength")
	ql.SetCnt(int64(this.queue.Len()))
	return &BackendStats{
		Send:          this.sendCounter.Get(),
		Drop:          this.dropCounter.Get(),
		Fail:          this.failCounter.Get(),
		QueueLength:   ql.Get(),
		ConnPoolStats: []*cpool.ConnPoolStats{this.pool.Stats()},
	}
}

func newBackendCommon(cfg *BackendConfig) *BackendCommon {
	return &BackendCommon{
		config:      cfg,
		queue:       nlist.NewSafeListLimited(DefaultSendQueueMaxSize),
		sendCounter: nproc.NewSCounterQps("Send"),
		dropCounter: nproc.NewSCounterQps("Drop"),
		failCounter: nproc.NewSCounterQps("Fail"),
	}
}
