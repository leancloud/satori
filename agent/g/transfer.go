package g

import (
	"log"
	"math/rand"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/leancloud/satori/common/cpool"
	"github.com/leancloud/satori/common/model"
)

//----------------------------
type TransferClient struct {
	cli  *rpc.Client
	name string
}

func (this TransferClient) Name() string {
	return this.name
}

func (this TransferClient) Closed() bool {
	return this.cli == nil
}

func (this TransferClient) Close() error {
	if this.cli == nil {
		this.cli.Close()
		this.cli = nil
	}
	return nil
}

func (this TransferClient) Call(metrics interface{}) (interface{}, error) {
	var resp model.TransferResponse
	err := this.cli.Call("Transfer.Update", metrics, &resp)
	if Config().Debug {
		log.Println("<=", &resp)
	}
	return resp, err
}

func transferConnect(name string, p *cpool.ConnPool) (cpool.PoolClient, error) {
	connTimeout := time.Duration(p.ConnTimeout) * time.Millisecond
	conn, err := net.DialTimeout("tcp", p.Address, connTimeout)
	if err != nil {
		log.Printf("Connect transfer %s fail: %v", p.Address, err)
		return nil, err
	}

	return TransferClient{
		cli:  jsonrpc.NewClient(conn),
		name: name,
	}, nil
}

var (
	transferClients []*cpool.ConnPool = []*cpool.ConnPool{}

	metricsBufferLock *sync.RWMutex        = new(sync.RWMutex)
	metricsBuffer     []*model.MetricValue = make([]*model.MetricValue, 0, 5)
)

// -------------------------
func sendMetrics() {
	metricsBufferLock.Lock()
	if len(metricsBuffer) == 0 {
		metricsBufferLock.Unlock()
		return
	}
	send := metricsBuffer
	metricsBuffer = make([]*model.MetricValue, 0, 5)
	metricsBufferLock.Unlock()

	for c := 0; c < 3; c++ {
		for _, i := range rand.Perm(len(transferClients)) {
			cli := transferClients[i]
			_, err := cli.Call(send)
			if err != nil {
				log.Println("sendMetrics fail", cli.Address, err)
				continue
			}
			return
		}
	}
	log.Printf("%s\n", "No available transfer client to send metrics, metrics dropped!")
}

func SendToTransferProc() {
	rand.Seed(time.Now().UnixNano())
	cfg := Config().Transfer
	for _, s := range cfg {
		u, err := url.Parse(s)
		if err != nil {
			log.Printf("Error parsing %s: %s\n", s, err.Error())
			continue
		}
		args := u.Query()
		var timeout int
		if ts := args.Get("timeout"); ts != "" {
			timeout, _ = strconv.Atoi(ts)
		} else {
			timeout = 3000
		}
		transferClients = append(transferClients, cpool.NewConnPool(
			"transfer", u.Host, 5, 3, timeout, timeout, transferConnect,
		))
	}

	for {
		time.Sleep(5 * time.Second)
		go sendMetrics()
	}
}

func SendToTransfer(metrics []*model.MetricValue) {
	if len(metrics) == 0 {
		return
	}

	metrics = filterMetrics(metrics)

	if len(metrics) == 0 {
		return
	}

	debug := Config().Debug

	if debug {
		log.Printf("=> <Total=%d> %v\n", len(metrics), metrics[0])
	}

	metricsBufferLock.Lock()
	defer metricsBufferLock.Unlock()

	metricsBuffer = append(metricsBuffer, metrics...)
}
