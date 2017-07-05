package sender

import (
	"log"
	"time"

	"github.com/amir/raidman"
	nsema "github.com/toolkits/concurrent/semaphore"

	"github.com/leancloud/satori/common/cpool"
	cmodel "github.com/leancloud/satori/common/model"
)

type RiemannBackend struct {
	BackendCommon
}

func newRiemannBackend(cfg *BackendConfig) Backend {
	if cfg.Engine == "riemann" {
		return &RiemannBackend{*newBackendCommon(cfg)}
	} else {
		return nil
	}
}

type RiemannClient struct {
	cli  *raidman.Client
	name string
}

func (this RiemannClient) Name() string {
	return this.name
}

func (this RiemannClient) Closed() bool {
	return this.cli == nil
}

func (this RiemannClient) Close() error {
	if this.cli != nil {
		this.cli.Close()
		this.cli = nil
	}
	return nil
}

func (this RiemannClient) Call(items interface{}) (interface{}, error) {
	err := this.cli.SendMulti(items.([]*raidman.Event))
	return nil, err
}

func (this *RiemannBackend) riemannConnect(name string, p *cpool.ConnPool) (cpool.NConn, error) {
	cfg := this.config
	u := cfg.Url

	proto := cfg.Protocol
	if proto == "" {
		proto = "tcp"
	}

	conn, err := raidman.DialWithTimeout(
		proto,
		u.Host,
		time.Duration(p.ConnTimeout)*time.Millisecond,
	)

	if err != nil {
		return nil, err
	}

	return RiemannClient{conn, name}, nil
}

func (this *RiemannBackend) createConnPool() *cpool.ConnPool {
	cfg := this.config
	u := cfg.Url

	p := cpool.NewConnPool(
		cfg.Name,
		u.Host,
		cfg.MaxConn,
		cfg.MaxIdle,
		cfg.ConnTimeout,
		cfg.CallTimeout,
		this.riemannConnect,
	)

	return p
}

func (this *RiemannBackend) Start() error {
	this.pool = this.createConnPool()
	go this.sendProc()
	return nil
}

func (this *RiemannBackend) sendProc() {
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

			riemannItems := make([]*raidman.Event, len(itemList))
			for i := range itemList {
				riemannItems[i] = itemList[i].(*raidman.Event)
			}

			var err error
			for i := 0; i < 3; i++ {
				_, err = this.pool.Call(riemannItems)
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

// 将原始数据入到riemann发送缓存队列
func (this *RiemannBackend) Send(items []*cmodel.MetricValue) {
	for _, item := range items {
		riemannItem := raidman.Event{
			Service:     item.Metric,
			Host:        item.Endpoint,
			State:       "",
			Metric:      item.Value,
			Time:        item.Timestamp,
			Description: item.Desc,
			Tags:        nil,
			Ttl:         float32(item.Step),
			Attributes:  make(map[string]string),
		}

		for k, v := range item.Tags {
			riemannItem.Attributes[k] = v
		}

		isSuccess := this.queue.PushFront(&riemannItem)

		if !isSuccess {
			this.dropCounter.Incr()
		}
	}
}
