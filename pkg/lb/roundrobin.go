package lb

import (
	"sync/atomic"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
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
func (rr RoundRobin) Select(req *fasthttp.Request, servers []metapb.Server) uint64 {
	l := uint64(len(servers))

	if 0 >= l {
		return 0
	}

	target := servers[int(atomic.AddUint64(rr.ops, 1)%l)]
	return target.ID
}
