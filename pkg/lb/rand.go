package lb

import (
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastrand"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

type RandBalance struct {
}

func NewRandBalance() LoadBalance {
	lb := RandBalance{}
	return lb
}

func (this RandBalance) Select(ctx *fasthttp.RequestCtx, servers []metapb.Server) uint64 {
	size := len(servers)
	if size < 1 {
		return 0
	}
	server := servers[fastrand.Uint32n(uint32(size))]
	return server.ID
}
