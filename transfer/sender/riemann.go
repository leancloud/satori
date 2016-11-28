package sender

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/amir/raidman"
	nsema "github.com/toolkits/concurrent/semaphore"
	nlist "github.com/toolkits/container/list"
	nproc "github.com/toolkits/proc"

	cmodel "github.com/leancloud/satori/common/model"
	"github.com/leancloud/satori/transfer/g"
	cpool "github.com/leancloud/satori/transfer/sender/conn_pool"
)

var (
	riemannConnPool    *cpool.ConnPool
	riemannQueue       = nlist.NewSafeListLimited(DefaultSendQueueMaxSize)
	riemannSendCounter = nproc.NewSCounterQps("riemannSend")
	riemannDropCounter = nproc.NewSCounterQps("riemannDrop")
	riemannFailCounter = nproc.NewSCounterQps("riemannFail")
	riemannQueueLength = nproc.NewSCounterBase("riemannQueueLength")
)

var RiemannBackend = Backend{
	Name:     "riemann",
	Start:    startRiemannTransfer,
	Send:     pushToRiemann,
	GetStats: riemannStats,
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

func (this RiemannClient) Call(items interface{}) error {
	err := this.cli.SendMulti(items.([]*raidman.Event))
	return err
}

func riemannConnect(name string, p *cpool.ConnPool) (cpool.NConn, error) {
	u, err := url.Parse(p.Address)
	if err != nil {
		return nil, err
	}

	conn, err := raidman.DialWithTimeout(
		u.Scheme,
		u.Host,
		time.Duration(p.ConnTimeout)*time.Millisecond,
	)

	if err != nil {
		return nil, err
	}

	return RiemannClient{conn, name}, nil
}

func riemannConnPoolFactory() *cpool.ConnPool {
	cfg := g.Config().Riemann
	addr := cfg.Address
	p := cpool.NewConnPool(
		"riemann",
		addr,
		cfg.MaxConns,
		cfg.MaxIdle,
		cfg.ConnTimeout,
		cfg.CallTimeout,
		riemannConnect,
	)

	return p
}

func startRiemannTransfer() error {
	cfg := g.Config().Riemann
	if cfg == nil {
		return fmt.Errorf("Riemann not configured")
	}

	if !cfg.Enabled {
		return fmt.Errorf("Riemann not enabled")
	}

	riemannConnPool = riemannConnPoolFactory()

	go riemannTransfer()
	return nil
}

func riemannTransfer() {
	cfg := g.Config().Riemann
	batch := cfg.Batch // 一次发送,最多batch条数据
	sema := nsema.NewSemaphore(cfg.MaxConns)

	for {
		items := riemannQueue.PopBackBy(batch)
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
				err = riemannConnPool.Call(riemannItems)
				if err == nil {
					riemannSendCounter.IncrBy(int64(len(itemList)))
					break
				}
				time.Sleep(100 * time.Millisecond)
			}

			if err != nil {
				riemannFailCounter.IncrBy(int64(len(itemList)))
				log.Println(err)
				return
			}
		}(items)
	}
}

// 将原始数据入到riemann发送缓存队列
func pushToRiemann(items []*cmodel.MetricValue) {
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

		isSuccess := riemannQueue.PushFront(&riemannItem)

		if !isSuccess {
			riemannDropCounter.Incr()
		}
	}
}

func riemannStats() *BackendStats {
	riemannQueueLength.SetCnt(int64(riemannQueue.Len()))
	return &BackendStats{
		SendCounter:   riemannSendCounter.Get(),
		DropCounter:   riemannDropCounter.Get(),
		FailCounter:   riemannFailCounter.Get(),
		QueueLength:   riemannQueueLength.Get(),
		ConnPoolStats: []*cpool.ConnPoolStats{riemannConnPool.Stats()},
	}
}
