package lb

import (
	"container/list"

	"github.com/valyala/fasthttp"
)

const (
	// ROUNDROBIN round robin
	ROUNDROBIN = "ROUNDROBIN"
)

var (
	supportLbs = []string{ROUNDROBIN}
)

var (
	// LBS map loadBalance name and process function
	LBS = map[string]func() LoadBalance{
		ROUNDROBIN: NewRoundRobin,
	}
)

// LoadBalance loadBalance interface
type LoadBalance interface {
	Select(req *fasthttp.Request, servers *list.List) int
}

// GetSupportLBS return supported loadBalances
func GetSupportLBS() []string {
	return supportLbs
}

// NewLoadBalance create a LoadBalance
func NewLoadBalance(name string) LoadBalance {
	return LBS[name]()
}
