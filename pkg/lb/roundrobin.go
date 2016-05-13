package lb

import (
	"container/list"
	"net/http"
	"sync/atomic"
)

type RoundRobin struct {
	ops *uint64
}

func NewRoundRobin() LoadBalance {
	var ops uint64
	ops = 0

	return RoundRobin{
		ops: &ops,
	}
}

func (self RoundRobin) Select(req *http.Request, servers *list.List) int {
	l := uint64(servers.Len())

	if 0 >= l {
		return -1
	}

	return int(atomic.AddUint64(self.ops, 1) % l)
}
