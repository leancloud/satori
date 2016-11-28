package conn_pool

import (
	"fmt"
	"io"
	"sync"
	"time"
)

//TODO: 保存所有的连接, 而不是只保存连接计数

var ErrMaxConn = fmt.Errorf("maximum connections reached")

//
type NConn interface {
	io.Closer
	Name() string
	Call(arg interface{}) error
	Closed() bool
}

type ConnPool struct {
	sync.RWMutex

	Name        string
	Address     string
	MaxConns    int
	MaxIdle     int
	ConnTimeout int
	CallTimeout int
	Cnt         int64
	New         func(name string, pool *ConnPool) (NConn, error)

	active int
	free   []NConn
	all    map[string]NConn
}

type ConnPoolStats struct {
	Name   string
	Count  int64
	Active int
	All    int
	Free   int
}

func (this *ConnPoolStats) String() string {
	return fmt.Sprintf("%s[Count: %d, Active: %d, All: %d, Free: %d]",
		this.Name,
		this.Count,
		this.Active,
		this.All,
		this.Free,
	)
}

func NewConnPool(name string, address string, maxConns int, maxIdle int, connTimeout int, callTimeout int, new func(string, *ConnPool) (NConn, error)) *ConnPool {
	return &ConnPool{
		Name:        name,
		Address:     address,
		MaxConns:    maxConns,
		MaxIdle:     maxIdle,
		CallTimeout: callTimeout,
		ConnTimeout: connTimeout,
		Cnt:         0,
		New:         new,
		all:         make(map[string]NConn),
	}
}

func (this *ConnPool) Stats() *ConnPoolStats {
	this.RLock()
	defer this.RUnlock()

	return &ConnPoolStats{
		Name:   this.Name,
		Count:  this.Cnt,
		Active: this.active,
		All:    len(this.all),
		Free:   len(this.free),
	}

}

func (this *ConnPool) Fetch() (NConn, error) {
	this.Lock()
	defer this.Unlock()

	// get from free
	conn := this.fetchFree()
	if conn != nil {
		return conn, nil
	}

	if this.overMax() {
		return nil, ErrMaxConn
	}

	// create new conn
	conn, err := this.newConn()
	if err != nil {
		return nil, err
	}

	this.increActive()
	return conn, nil
}

func (this *ConnPool) Call(arg interface{}) error {
	conn, err := this.Fetch()
	if err != nil {
		return fmt.Errorf("%s get connection fail: conn %v, err %v. stats: %s", this.Name, conn, err, this.Stats())
	}

	callTimeout := time.Duration(this.CallTimeout) * time.Millisecond

	done := make(chan error)
	go func() {
		done <- conn.Call(arg)
	}()

	select {
	case <-time.After(callTimeout):
		this.ForceClose(conn)
		return fmt.Errorf("%s, call timeout", conn.Name())
	case err = <-done:
		if err != nil {
			this.ForceClose(conn)
			err = fmt.Errorf("%s, call failed, err %v. stats: %s", this.Name, err, this.Stats())
		} else {
			this.Release(conn)
		}
		return err
	}
}

func (this *ConnPool) Release(conn NConn) {
	this.Lock()
	defer this.Unlock()

	if this.overMaxIdle() {
		this.deleteConn(conn)
		this.decreActive()
	} else {
		this.addFree(conn)
	}
}

func (this *ConnPool) ForceClose(conn NConn) {
	this.Lock()
	defer this.Unlock()

	this.deleteConn(conn)
	this.decreActive()
}

func (this *ConnPool) Destroy() {
	this.Lock()
	defer this.Unlock()

	for _, conn := range this.free {
		if conn != nil && !conn.Closed() {
			conn.Close()
		}
	}

	for _, conn := range this.all {
		if conn != nil && !conn.Closed() {
			conn.Close()
		}
	}

	this.active = 0
	this.free = []NConn{}
	this.all = map[string]NConn{}
}

// internal, concurrently unsafe
func (this *ConnPool) newConn() (NConn, error) {
	name := fmt.Sprintf("%s_%d_%d", this.Name, this.Cnt, time.Now().Unix())
	conn, err := this.New(name, this)
	if err != nil {
		if conn != nil {
			conn.Close()
		}
		return nil, err
	}

	this.Cnt++
	this.all[conn.Name()] = conn
	return conn, nil
}

func (this *ConnPool) deleteConn(conn NConn) {
	if conn != nil {
		conn.Close()
	}
	delete(this.all, conn.Name())
}

func (this *ConnPool) addFree(conn NConn) {
	this.free = append(this.free, conn)
}

func (this *ConnPool) fetchFree() NConn {
	if len(this.free) == 0 {
		return nil
	}

	conn := this.free[0]
	this.free = this.free[1:]
	return conn
}

func (this *ConnPool) increActive() {
	this.active += 1
}

func (this *ConnPool) decreActive() {
	this.active -= 1
}

func (this *ConnPool) overMax() bool {
	return this.active >= this.MaxConns
}

func (this *ConnPool) overMaxIdle() bool {
	return len(this.free) >= this.MaxIdle
}
