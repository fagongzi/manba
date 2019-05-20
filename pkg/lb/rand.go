package lb

import (
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastrand"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

// RandBalance is rand loadBalance impl
type RandBalance struct {
}

// NewRandBalance create a RandBalance
func NewRandBalance() LoadBalance {
	lb := RandBalance{}
	return lb
}

// Select select a server from servers using rand
func (this RandBalance) Select(ctx *fasthttp.RequestCtx, servers []metapb.Server) uint64 {
	l := len(servers)
	if 0 >= l {
		return 0
	}
	server := servers[fastrand.Uint32n(uint32(l))]
	return server.ID
}
