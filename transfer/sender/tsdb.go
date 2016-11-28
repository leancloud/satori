package sender

import (
	"bytes"
	"fmt"
	"github.com/leancloud/satori/transfer/g"
	"log"
	"net"
	"strings"
	"time"

	nsema "github.com/toolkits/concurrent/semaphore"
	nlist "github.com/toolkits/container/list"
	nproc "github.com/toolkits/proc"

	cmodel "github.com/leancloud/satori/common/model"
	cpool "github.com/leancloud/satori/transfer/sender/conn_pool"
)

var (
	tsdbConnPool    *cpool.ConnPool
	tsdbQueue       = nlist.NewSafeListLimited(DefaultSendQueueMaxSize)
	tsdbSendCounter = nproc.NewSCounterQps("tsdbSend")
	tsdbDropCounter = nproc.NewSCounterQps("tsdbDrop")
	tsdbFailCounter = nproc.NewSCounterQps("tsdbFail")
	tsdbQueueLength = nproc.NewSCounterBase("tsdbQueueLength")
)

var TsdbBackend = Backend{
	Name:     "tsdb",
	Start:    startTsdbTransfer,
	Send:     pushToTsdb,
	GetStats: tsdbStats,
}

type TsdbItem struct {
	Metric    string            `json:"metric"`
	Tags      map[string]string `json:"tags"`
	Value     float64           `json:"value"`
	Timestamp int64             `json:"timestamp"`
}

func (this *TsdbItem) String() string {
	return fmt.Sprintf(
		"<Metric:%s, Tags:%v, Value:%v, TS:%d>",
		this.Metric,
		this.Tags,
		this.Value,
		this.Timestamp,
	)
}

func (this *TsdbItem) TsdbString() (s string) {
	s = fmt.Sprintf("put %s %d %.3f ", this.Metric, this.Timestamp, this.Value)

	for k, v := range this.Tags {
		key := strings.ToLower(strings.Replace(k, " ", "_", -1))
		value := strings.Replace(v, " ", "_", -1)
		s += key + "=" + value + " "
	}

	return s
}

type TsdbClient struct {
	cli  net.Conn
	name string
}

func (this TsdbClient) Name() string {
	return this.name
}

func (this TsdbClient) Closed() bool {
	return this.cli == nil
}

func (this TsdbClient) Close() error {
	if this.cli != nil {
		err := this.cli.Close()
		this.cli = nil
		return err
	}
	return nil
}

func (this TsdbClient) Call(items interface{}) error {
	var err error
	_, err = this.cli.Write(items.([]byte))
	return err
}

func tsdbConnect(name string, p *cpool.ConnPool) (cpool.NConn, error) {
	addr := p.Address
	_, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}

	connTimeout := time.Duration(p.ConnTimeout) * time.Millisecond
	conn, err := net.DialTimeout("tcp", addr, connTimeout)
	if err != nil {
		return nil, err
	}

	return TsdbClient{conn, name}, nil
}

func tsdbConnPoolFactory() *cpool.ConnPool {
	cfg := g.Config().Tsdb
	addr := cfg.Address
	p := cpool.NewConnPool(
		"tsdb",
		addr,
		cfg.MaxConns,
		cfg.MaxIdle,
		cfg.ConnTimeout,
		cfg.CallTimeout,
		tsdbConnect,
	)

	return p
}

func startTsdbTransfer() error {
	cfg := g.Config().Tsdb
	if cfg == nil {
		return fmt.Errorf("TSDB not configured")
	}

	if !cfg.Enabled {
		return fmt.Errorf("TSDB not enabled")
	}

	tsdbConnPool = tsdbConnPoolFactory()

	go tsdbTransfer()
	return nil
}

func tsdbTransfer() {
	cfg := g.Config().Tsdb

	batch := cfg.Batch // 一次发送,最多batch条数据
	retry := cfg.MaxRetry
	sema := nsema.NewSemaphore(cfg.MaxConns)

	for {
		items := tsdbQueue.PopBackBy(batch)
		if len(items) == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}
		//  同步Call + 有限并发 进行发送
		sema.Acquire()
		go func(itemList []interface{}) {
			defer sema.Release()

			var tsdbBuffer bytes.Buffer
			for i := 0; i < len(itemList); i++ {
				tsdbItem := itemList[i].(*TsdbItem)
				tsdbBuffer.WriteString(tsdbItem.TsdbString())
				tsdbBuffer.WriteString("\n")
			}

			var err error
			for i := 0; i < retry; i++ {
				err = tsdbConnPool.Call(tsdbBuffer.Bytes())
				if err == nil {
					tsdbSendCounter.IncrBy(int64(len(itemList)))
					break
				}
				time.Sleep(100 * time.Millisecond)
			}

			if err != nil {
				tsdbFailCounter.IncrBy(int64(len(itemList)))
				log.Println(err)
				return
			}
		}(items)
	}
}

// 将原始数据入到tsdb发送缓存队列
func pushToTsdb(items []*cmodel.MetricValue) {
	for _, item := range items {
		tsdbItem := convert2TsdbItem(item)
		isSuccess := tsdbQueue.PushFront(tsdbItem)

		if !isSuccess {
			tsdbDropCounter.Incr()
		}
	}
}

// 转化为tsdb格式
func convert2TsdbItem(d *cmodel.MetricValue) *TsdbItem {
	t := TsdbItem{Tags: make(map[string]string)}

	for k, v := range d.Tags {
		t.Tags[k] = v
	}
	t.Tags["endpoint"] = d.Endpoint
	t.Metric = d.Metric
	t.Timestamp = d.Timestamp
	t.Value = d.Value
	return &t
}

func alignTs(ts int64, period int64) int64 {
	return ts - ts%period
}

func tsdbStats() *BackendStats {
	tsdbQueueLength.SetCnt(int64(tsdbQueue.Len()))
	return &BackendStats{
		SendCounter:   tsdbSendCounter.Get(),
		DropCounter:   tsdbDropCounter.Get(),
		FailCounter:   tsdbFailCounter.Get(),
		QueueLength:   tsdbQueueLength.Get(),
		ConnPoolStats: []*cpool.ConnPoolStats{tsdbConnPool.Stats()},
	}
}
