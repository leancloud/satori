package cpool

import "io"

type PoolClient interface {
	io.Closer
	Name() string
	Call(arg interface{}) (interface{}, error)
	Closed() bool
}
