package conn_pool

import (
	"fmt"
	"sync"
)

// ConnPools Manager
type ClusteredConnPool struct {
	sync.RWMutex
	M map[string]*ConnPool
}

func CreateClusteredConnPool(connPoolFactory func(addr string) *ConnPool, cluster []string) *ClusteredConnPool {
	cp := &ClusteredConnPool{M: make(map[string]*ConnPool)}

	for _, addr := range cluster {
		cp.M[addr] = connPoolFactory(addr)
	}
	return cp
}

// 同步发送, 完成发送或超时后 才能返回
func (this *ClusteredConnPool) Call(addr string, arg interface{}) error {
	connPool, exists := this.Get(addr)
	if !exists {
		return fmt.Errorf("%s has no connection pool", addr)
	}
	return connPool.Call(arg)
}

func (this *ClusteredConnPool) Get(address string) (*ConnPool, bool) {
	this.RLock()
	defer this.RUnlock()
	p, exists := this.M[address]
	return p, exists
}

func (this *ClusteredConnPool) Destroy() {
	this.Lock()
	defer this.Unlock()
	addresses := make([]string, 0, len(this.M))
	for address := range this.M {
		addresses = append(addresses, address)
	}

	for _, address := range addresses {
		this.M[address].Destroy()
		delete(this.M, address)
	}
}

func (this *ClusteredConnPool) Stats() []*ConnPoolStats {
	rst := make([]*ConnPoolStats, 0, 5)
	for _, p := range this.M {
		rst = append(rst, p.Stats())
	}
	return rst
}
