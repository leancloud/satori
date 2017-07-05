package sender

import (
	"log"
	"net"
	"net/url"
	"strings"
	"time"

	influxdb "github.com/influxdata/influxdb/client/v2"
	nsema "github.com/toolkits/concurrent/semaphore"

	"github.com/leancloud/satori/common/cpool"
	cmodel "github.com/leancloud/satori/common/model"
)

type InfluxdbBackend struct {
	BackendCommon
}

func newInfluxdbBackend(cfg *BackendConfig) Backend {
	if cfg.Engine == "influxdb" {
		return &InfluxdbBackend{*newBackendCommon(cfg)}
	} else {
		return nil
	}
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

func (this InfluxdbClient) Call(arg interface{}) (interface{}, error) {
	bp, err := influxdb.NewBatchPoints(influxdb.BatchPointsConfig{
		Database:  this.dbName,
		Precision: "s",
	})
	if err != nil {
		return nil, err
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
			return nil, err
		}
		bp.AddPoint(pt)
	}

	// Write the batch
	return nil, this.cli.Write(bp)
}

func (this *InfluxdbBackend) influxdbConnect(name string, p *cpool.ConnPool) (cpool.PoolClient, error) {
	cfg := this.config
	u := cfg.Url

	connTimeout := time.Duration(p.ConnTimeout) * time.Millisecond
	_, err := net.DialTimeout("tcp", u.Host, connTimeout)
	if err != nil {
		// log.Printf("new conn fail, addr %s, err %v", p.Address, err)
		return nil, err
	}

	pwd, _ := u.User.Password()

	proto := cfg.Protocol
	if proto == "" {
		proto = "http"
	}

	c, err := influxdb.NewHTTPClient(
		influxdb.HTTPConfig{
			Addr:     (&url.URL{Scheme: proto, Host: u.Host}).String(),
			Username: u.User.Username(),
			Password: pwd,
		},
	)

	if err != nil {
		return nil, err
	}

	return InfluxdbClient{
		cli:    c,
		name:   name,
		dbName: u.Path[1:],
	}, nil
}

func (this *InfluxdbBackend) createConnPool() *cpool.ConnPool {
	cfg := this.config
	return cpool.NewConnPool(
		cfg.Name,
		cfg.Url.Host,
		cfg.MaxConn,
		cfg.MaxIdle,
		cfg.ConnTimeout,
		cfg.CallTimeout,
		this.influxdbConnect,
	)
}

func (this *InfluxdbBackend) Start() error {
	this.pool = this.createConnPool()
	go this.sendProc()
	return nil
}

func (this *InfluxdbBackend) sendProc() {
	cfg := this.config
	batch := cfg.Batch // 一次发送,最多batch条数据

	sema := nsema.NewSemaphore(cfg.MaxConn)
	for {
		items := this.queue.PopBackBy(batch)
		count := len(items)
		if count == 0 {
			time.Sleep(DefaultSendTaskSleepInterval)
			continue
		}

		influxdbItems := make([]*cmodel.MetricValue, 0, count)
		for i := 0; i < count; i++ {
			m := items[i].(*cmodel.MetricValue)
			if m.Metric[0:7] == ".satori" {
				// skip internal events
				continue
			}
			influxdbItems = append(influxdbItems, m)
		}

		//	同步Call + 有限并发 进行发送
		sema.Acquire()
		go func(influxdbItems []*cmodel.MetricValue, count int) {
			defer sema.Release()

			var err error
			sendOk := false
			for i := 0; i < cfg.Retry; i++ {
				_, err = this.pool.Call(influxdbItems)
				if err == nil {
					sendOk = true
					break
				}
				time.Sleep(time.Millisecond * 10)
			}

			// statistics
			if !sendOk {
				log.Printf("send influxdb %s fail: %v", cfg.Name, err)
				this.failCounter.IncrBy(int64(count))
			} else {
				this.sendCounter.IncrBy(int64(count))
			}
		}(influxdbItems, count)
	}
}

// Push data to 3rd-party database
func (this *InfluxdbBackend) Send(items []*cmodel.MetricValue) {
	for _, item := range items {
		myItem := item
		myItem.Timestamp = item.Timestamp

		isSuccess := this.queue.PushFront(myItem)

		// statistics
		if !isSuccess {
			this.dropCounter.Incr()
		}
	}
}
