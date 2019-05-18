package lb

import (
	"testing"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/valyala/fasthttp"
)

func TestWeightRobin_Select(t *testing.T) {
	var values []metapb.Server
	values = append(values, metapb.Server{
		ID:     1,
		Weight: 20,
	})

	values = append(values, metapb.Server{
		ID:     2,
		Weight: 10,
	})

	values = append(values, metapb.Server{
		ID:     3,
		Weight: 35,
	})

	values = append(values, metapb.Server{
		ID:     4,
		Weight: 5,
	})

	type fields struct {
		opts map[uint64]*weightRobin
	}
	type args struct {
		req     *fasthttp.Request
		servers []metapb.Server
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantBest []int
	}{
		{
			name:   "test_case_1",
			fields: struct{ opts map[uint64]*weightRobin }{opts: make(map[uint64]*weightRobin, 50)},
			args: struct {
				req     *fasthttp.Request
				servers []metapb.Server
			}{req: nil, servers: values},
			wantBest: []int{20, 10, 35, 5},
		},
	}
	for _, tt := range tests {
		var res = make(map[uint64]int)
		t.Run(tt.name, func(t *testing.T) {
			w := &WeightRobin{
				opts: tt.fields.opts,
			}
			for i := 0; i < 70; i++ {
				res[w.Select(tt.args.req, tt.args.servers)]++
			}
		})
		for k, v := range res {
			if tt.wantBest[k-1] != v {
				t.Errorf("WeightRobin.Select() = %v, want %v", res, tt.wantBest)
			}
		}
	}
}
