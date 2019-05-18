package lb

import (
	"math/rand"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

type RandBalance struct {
	Random *rand.Rand
}

func NewRandBalance() LoadBalance {
	s := rand.NewSource(time.Now().UnixNano())
	lb := RandBalance{
		Random: rand.New(s),
	}
	return lb
}

func (this RandBalance) Select(ctx *fasthttp.RequestCtx, servers []metapb.Server) uint64 {
	size := len(servers)
	if size < 1 {
		return 0
	}
	server := servers[this.Random.Intn(size)]
	return server.ID
}

func init() {
	//Register("hash", newHashIPBalancer())
}
