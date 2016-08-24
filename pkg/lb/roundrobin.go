package lb

import (
	"container/list"
	"net/http"
	"sync/atomic"
)

// RoundRobin round robin loadBalance impl
type RoundRobin struct {
	ops *uint64
}

// NewRoundRobin create a RoundRobin
func NewRoundRobin() LoadBalance {
	var ops uint64
	ops = 0

	return RoundRobin{
		ops: &ops,
	}
}

// Select select a server from servers using RoundRobin
func (rr RoundRobin) Select(req *http.Request, servers *list.List) int {
	l := uint64(servers.Len())

	if 0 >= l {
		return -1
	}

	return int(atomic.AddUint64(rr.ops, 1) % l)
}
