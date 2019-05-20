package lb

import (
	"testing"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/valyala/fasthttp"
)

var (
	Servers = []metapb.Server{
		metapb.Server{
			ID:     1,
			Weight: 10,
		},
		metapb.Server{
			ID:     2,
			Weight: 20,
		},
		metapb.Server{
			ID:     3,
			Weight: 40,
		},
		metapb.Server{
			ID:     5,
			Weight: 50,
		},
		metapb.Server{
			ID:     19,
			Weight: 20,
		},
	}
)

func Test_HashIPBalance(t *testing.T) {
	lb := NewHashIPBalance()
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.Add("X-Forwarded-For", "192.168.0.5")
	for i := 0; i < 66; i++ {
		id := lb.Select(reqCtx, Servers)
		if id < 1 {
			t.Errorf("Test_HashIPBalance is error=%d", id)
		}
		t.Logf("id=%d", id)
	}
}

func Benchmark_HashIPBalance(b *testing.B) {
	lb := NewHashIPBalance()
	reqCtx := &fasthttp.RequestCtx{}
	reqCtx.Request.Header.Add("X-Forwarded-For", "192.168.0.5")
	for i := 0; i < b.N; i++ {
		id := lb.Select(reqCtx, Servers)
		if id < 1 {
			b.Errorf("Test_HashIPBalance is error=%d", id)
		}
	}
}
