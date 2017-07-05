package sender

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	nsema "github.com/toolkits/concurrent/semaphore"

	"github.com/leancloud/satori/common/cpool"
	cmodel "github.com/leancloud/satori/common/model"
)

type TsdbBackend struct {
	BackendCommon
}

func newTsdbBackend(cfg *BackendConfig) Backend {
	if cfg.Engine == "tsdb" {
		return &TsdbBackend{*newBackendCommon(cfg)}
	} else {
		return nil
	}
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

func (this TsdbClient) Call(items interface{}) (interface{}, error) {
	return this.cli.Write(items.([]byte))
}

func (this *TsdbBackend) tsdbConnect(name string, p *cpool.ConnPool) (cpool.NConn, error) {
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

func (this *TsdbBackend) createConnPool() *cpool.ConnPool {
	cfg := this.config
	p := cpool.NewConnPool(
		cfg.Name,
		cfg.Url.Host,
		cfg.MaxConn,
		cfg.MaxIdle,
		cfg.ConnTimeout,
		cfg.CallTimeout,
		this.tsdbConnect,
	)

	return p
}

func (this *TsdbBackend) Start() error {
	this.pool = this.createConnPool()
	go this.sendProc()
	return nil
}

func (this *TsdbBackend) sendProc() {
	cfg := this.config
	sema := nsema.NewSemaphore(cfg.MaxConn)

	for {
		items := this.queue.PopBackBy(cfg.Batch)
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
			for i := 0; i < cfg.Retry; i++ {
				_, err = this.pool.Call(tsdbBuffer.Bytes())
				if err == nil {
					this.sendCounter.IncrBy(int64(len(itemList)))
					break
				}
				time.Sleep(100 * time.Millisecond)
			}

			if err != nil {
				this.failCounter.IncrBy(int64(len(itemList)))
				log.Println(err)
				return
			}
		}(items)
	}
}

// 将原始数据入到tsdb发送缓存队列
func (this *TsdbBackend) Send(items []*cmodel.MetricValue) {
	for _, item := range items {
		tsdbItem := convert2TsdbItem(item)
		isSuccess := this.queue.PushFront(tsdbItem)

		if !isSuccess {
			this.dropCounter.Incr()
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
