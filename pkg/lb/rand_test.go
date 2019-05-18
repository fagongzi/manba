package lb

import (
	"github.com/valyala/fasthttp"
	"testing"
)

func Test_RandBalance(t *testing.T) {
	lb := NewRandBalance()
	reqCtx := &fasthttp.RequestCtx{}
	for i := 0; i < 66; i++ {
		id := lb.Select(reqCtx, Servers)
		if id < 1 {
			t.Errorf("Test_HashIPBalance is error=%d", id)
		}
		t.Logf("id=%d", id)
	}
}

func Benchmark_RandBalance(b *testing.B) {
	lb := NewRandBalance()
	reqCtx := &fasthttp.RequestCtx{}
	for i := 0; i < b.N; i++ {
		id := lb.Select(reqCtx, Servers)
		if id < 1 {
			b.Errorf("Test_HashIPBalance is error=%d", id)
		}
	}
}
