package sender

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
	nsema "github.com/toolkits/concurrent/semaphore"
	nlist "github.com/toolkits/container/list"
	nproc "github.com/toolkits/proc"

	cmodel "github.com/leancloud/satori/common/model"
	"github.com/leancloud/satori/transfer/g"
	cpool "github.com/leancloud/satori/transfer/sender/conn_pool"
)

var (
	errInvalidDSNUnescaped = errors.New("Invalid DSN: Did you forget to escape a param value?")
	errInvalidDSNAddr      = errors.New("Invalid DSN: Network Address not terminated (missing closing brace)")
	errInvalidDSNNoSlash   = errors.New("Invalid DSN: Missing the slash separating the database name")
)

var (
	influxdbConnPool    *cpool.ConnPool
	influxdbQueue       = nlist.NewSafeListLimited(DefaultSendQueueMaxSize)
	influxdbSendCounter = nproc.NewSCounterQps("influxdbSend")
	influxdbDropCounter = nproc.NewSCounterQps("influxdbDrop")
	influxdbFailCounter = nproc.NewSCounterQps("influxdbFail")
	influxdbQueueLength = nproc.NewSCounterBase("influxdbQueueLength")
)

var InfluxdbBackend = Backend{
	Name:     "influxdb",
	Start:    startInfluxdbTransfer,
	Send:     pushToInfluxdb,
	GetStats: influxdbStats,
}

type InfluxdbClient struct {
	cli    influxdb.Client
	name   string
	dbName string
}

func (this InfluxdbClient) Name() string {
	return this.name
}

func (this InfluxdbClient) Closed() bool {
	return this.cli == nil
}

func (this InfluxdbClient) Close() error {
	if this.cli != nil {
		err := this.cli.Close()
		this.cli = nil
		return err
	}
	return nil
}

func (this InfluxdbClient) Call(arg interface{}) error {
	bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
		Database:  this.dbName,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	items := arg.([]*cmodel.MetricValue)

	for _, item := range items {
		token := strings.SplitN(item.Metric, ".", 2)
		var measurement, field string
		if len(token) == 1 {
			measurement = "_other"
			field = token[0]
		} else if len(token) == 2 {
			measurement = token[0]
			field = token[1]
		}

		// Create a point and add to batch
		tags := map[string]string{
			"host": item.Endpoint,
		}
		fields := map[string]interface{}{
			field: item.Value,
		}
		for k, v := range item.Tags {
			tags[k] = v
		}
		pt, err := influxdb.NewPoint(measurement, tags, fields, time.Unix(item.Timestamp, 0))
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}

	// Write the batch
	return this.cli.Write(bp)
}

func influxdbConnect(name string, p *cpool.ConnPool) (cpool.NConn, error) {
	conf, err := url.Parse(p.Address)
	if err != nil {
		return nil, err
	}

	connTimeout := time.Duration(p.ConnTimeout) * time.Millisecond
	_, err = net.DialTimeout("tcp", conf.Host, connTimeout)
	if err != nil {
		// log.Printf("new conn fail, addr %s, err %v", p.Address, err)
		return nil, err
	}

	pwd, _ := conf.User.Password()

	c, err := influxdb.NewHTTPClient(
		influxdb.HTTPConfig{
			Addr:     (&url.URL{Scheme: conf.Scheme, Host: conf.Host}).String(),
			Username: conf.User.Username(),
			Password: pwd,
		},
	)

	if err != nil {
		return nil, err
	}

	return InfluxdbClient{
		cli:    c,
		name:   name,
		dbName: conf.Path[1:],
	}, nil
}

func influxdbConnPoolFactory() (*cpool.ConnPool, error) {
	cfg := g.Config().Influxdb
	addr := cfg.Address
	_, err := url.Parse(cfg.Address)

	if err != nil {
		return nil, err
	}

	p := cpool.NewConnPool(
		"influxdb",
		addr,
		cfg.MaxConns,
		cfg.MaxIdle,
		cfg.ConnTimeout,
		cfg.CallTimeout,
		influxdbConnect,
	)

	return p, nil
}

func startInfluxdbTransfer() error {
	cfg := g.Config().Influxdb
	if cfg == nil {
		return fmt.Errorf("Influxdb not configured")
	}

	if !cfg.Enabled {
		return fmt.Errorf("Influxdb not enabled")
	}

	var err error
	influxdbConnPool, err = influxdbConnPoolFactory()

	if err != nil {
		log.Print("syntax of influxdb address is wrong")
		return err
	}

	go influxdbTransfer()
	return nil
}

func influxdbTransfer() {
	cfg := g.Config().Influxdb
	batch := cfg.Batch // 一次发送,最多batch条数据
	conn, err := url.Parse(cfg.Address)
	if err != nil {
		log.Print("syntax of influxdb address is wrong")
		return
	}
	addr := conn.Host

	sema := nsema.NewSemaphore(cfg.MaxConns)

	for {
		items := influxdbQueue.PopBackBy(batch)
		count := len(items)
		if count == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}

		influxdbItems := make([]*cmodel.MetricValue, count)
		for i := 0; i < count; i++ {
			influxdbItems[i] = items[i].(*cmodel.MetricValue)
		}

		//	同步Call + 有限并发 进行发送
		sema.Acquire()
		go func(addr string, influxdbItems []*cmodel.MetricValue, count int) {
			defer sema.Release()

			var err error
			sendOk := false
			for i := 0; i < 3; i++ { //最多重试3次
				err = influxdbConnPool.Call(influxdbItems)
				if err == nil {
					sendOk = true
					break
				}
				time.Sleep(time.Millisecond * 10)
			}

			// statistics
			if !sendOk {
				log.Printf("send influxdb %s fail: %v", addr, err)
				influxdbFailCounter.IncrBy(int64(count))
			} else {
				influxdbSendCounter.IncrBy(int64(count))
			}
		}(addr, influxdbItems, count)
	}
}

// Push data to 3rd-party database
func pushToInfluxdb(items []*cmodel.MetricValue) {
	for _, item := range items {
		// align ts
		step := int(item.Step)
		if step < MinStep {
			step = MinStep
		}
		ts := alignTs(item.Timestamp, int64(step))

		myItem := item
		myItem.Timestamp = ts

		isSuccess := influxdbQueue.PushFront(myItem)

		// statistics
		if !isSuccess {
			influxdbDropCounter.Incr()
		}
	}
}

func influxdbStats() *BackendStats {
	influxdbQueueLength.SetCnt(int64(influxdbQueue.Len()))
	return &BackendStats{
		SendCounter:   influxdbSendCounter.Get(),
		DropCounter:   influxdbDropCounter.Get(),
		FailCounter:   influxdbFailCounter.Get(),
		QueueLength:   influxdbQueueLength.Get(),
		ConnPoolStats: []*cpool.ConnPoolStats{influxdbConnPool.Stats()},
	}
}
