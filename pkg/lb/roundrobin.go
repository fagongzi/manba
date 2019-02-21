package lb

import (
	"container/list"
	"github.com/domgoer/hiruok/pkg/collection"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"sync/atomic"

	"github.com/valyala/fasthttp"
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
func (rr RoundRobin) Select(req *fasthttp.Request, servers *list.List) uint64 {
	l := uint64(servers.Len())

	if 0 >= l {
		return 0
	}

	idx := int(atomic.AddUint64(rr.ops, 1) % l)

	v := collection.Get(servers, idx).Value
	if v == nil {
		return 0
	}

	return v.(*metapb.Server).ID
}
