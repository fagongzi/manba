package lb

import (
	"container/list"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/valyala/fasthttp"
)

var (
	supportLbs = []metapb.LoadBalance{metapb.RoundRobin}
)

var (
	// LBS map loadBalance name and process function
	LBS = map[metapb.LoadBalance]func() LoadBalance{
		metapb.RoundRobin: NewRoundRobin,
	}
)

// LoadBalance loadBalance interface
type LoadBalance interface {
	Select(req *fasthttp.Request, servers *list.List) int
}

// GetSupportLBS return supported loadBalances
func GetSupportLBS() []metapb.LoadBalance {
	return supportLbs
}

// NewLoadBalance create a LoadBalance,if LoadBalance function is not supported
// it will return NewRoundRobin
func NewLoadBalance(name metapb.LoadBalance) LoadBalance {
	if l, ok := LBS[name]; ok {
		return l()
	}
	return NewRoundRobin()
}
