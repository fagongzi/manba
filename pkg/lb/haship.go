package lb

import (
	"hash/fnv"

	"github.com/valyala/fasthttp"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/util"
)

type HashIPBalance struct {
}

func NewHashIPBalance() LoadBalance {
	lb := HashIPBalance{}
	return lb
}

func (haship HashIPBalance) Select(ctx *fasthttp.RequestCtx, servers []metapb.Server) uint64 {
	size := len(servers)
	if size < 1 {
		return 0
	}
	hash := fnv.New32a()
	//key为客户端ip
	key := util.ClientIP(ctx)
	hash.Write([]byte(key))
	serve := servers[hash.Sum32()%uint32(size)]
	return serve.ID
}

func init() {
	//Register("hash", newHashIPBalancer())
}
