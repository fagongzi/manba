package lb

import (
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/valyala/fasthttp"
)

// WeightRobin weight robin loadBalance impl
type WeightRobin struct {
	opts map[uint64]*weightRobin
}

// weightRobin used to save the weight info of server
type weightRobin struct {
	effectiveWeight int64
	currentWeight   int64
}

// NewWeightRobin create a WeightRobin
func NewWeightRobin() LoadBalance {
	return &WeightRobin{
		opts: make(map[uint64]*weightRobin, 1024),
	}
}

// Select select a server from servers using WeightRobin
func (w *WeightRobin) Select(req *fasthttp.RequestCtx, servers []metapb.Server) (best uint64) {
	var total int64
	l := len(servers)

	for i := l - 1; i >= 0; i-- {
		svr := servers[i]

		id := svr.ID
		if _, ok := w.opts[id]; !ok {
			w.opts[id] = &weightRobin{
				effectiveWeight: svr.Weight,
			}
		}

		wt := w.opts[id]
		wt.currentWeight += wt.effectiveWeight
		total += wt.effectiveWeight

		if wt.effectiveWeight < svr.Weight {
			wt.effectiveWeight++
		}

		if best == 0 || w.opts[uint64(best)] == nil || wt.currentWeight > w.opts[best].currentWeight {
			best = id
		}
	}

	if best == 0 {
		return 0
	}

	w.opts[best].currentWeight -= total

	return best
}
