package sender

import (
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"strings"
	"time"

	nsema "github.com/toolkits/concurrent/semaphore"

	"github.com/leancloud/satori/common/cpool"
	cmodel "github.com/leancloud/satori/common/model"
	nproc "github.com/toolkits/proc"
)

type TransferBackend struct {
	BackendCommon
	ccp   *cpool.ClusteredConnPool
	addrs []string
}

func newTransferBackend(cfg *BackendConfig) Backend {
	if cfg.Engine == "transfer" {
		return &TransferBackend{BackendCommon: *newBackendCommon(cfg), ccp: nil}
	} else {
		return nil
	}
}

type TransferClient struct {
	cli  *rpc.Client
	name string
}

func (this *TransferClient) Name() string {
	return this.name
}

func (this *TransferClient) Closed() bool {
	return this.cli == nil
}

func (this *TransferClient) Close() error {
	if this.cli == nil {
		this.cli.Close()
		this.cli = nil
	}
	return nil
}

func (this *TransferClient) Call(metrics interface{}) (interface{}, error) {
	var resp cmodel.TransferResponse
	err := this.cli.Call("Transfer.Update", metrics, &resp)
	return resp, err
}

func (this *TransferBackend) transferConnect(name string, p *cpool.ConnPool) (cpool.PoolClient, error) {
	connTimeout := time.Duration(p.ConnTimeout) * time.Millisecond
	conn, err := net.DialTimeout("tcp", p.Address, connTimeout)
	if err != nil {
		log.Printf("Connect transfer %s fail: %v", p.Address, err)
		return nil, err
	}

	return &TransferClient{
		cli:  jsonrpc.NewClient(conn),
		name: name,
	}, nil
}

func (this *TransferBackend) Start() error {
	cfg := this.config
	u := cfg.Url
	addrs := strings.Split(u.Host, ",")
	this.addrs = addrs
	this.ccp = cpool.CreateClusteredConnPool(func(addr string) *cpool.ConnPool {
		return cpool.NewConnPool(
			addr,
			addr,
			cfg.MaxConn,
			cfg.MaxIdle,
			cfg.ConnTimeout,
			cfg.CallTimeout,
			this.transferConnect,
		)
	}, addrs)
	go this.sendProc()
	return nil
}

func (this *TransferBackend) sendProc() {
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

			addrs := this.addrs

			var err error
			for i := 0; i < 3; i++ {
				for _, i := range rand.Perm(len(addrs)) {
					_, err := this.ccp.Call(addrs[i], itemList)
					if err != nil {
						log.Println("sendMetrics fail", addrs[i], err)
						continue
					}
					break
				}
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
func (this *TransferBackend) Send(items []*cmodel.MetricValue) {
	for _, i := range items {
		if !this.queue.PushFront(i) {
			this.dropCounter.Incr()
		}
	}
}

func (this *TransferBackend) GetStats() *BackendStats {
	ql := nproc.NewSCounterBase("QueueLength")
	ql.SetCnt(int64(this.queue.Len()))
	return &BackendStats{
		Send:          this.sendCounter.Get(),
		Drop:          this.dropCounter.Get(),
		Fail:          this.failCounter.Get(),
		QueueLength:   ql.Get(),
		ConnPoolStats: this.ccp.Stats(),
	}
}
