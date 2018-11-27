package client

import (
	"github.com/fagongzi/gateway/pkg/pb"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/pb/rpcpb"
)

// RoutingBuilder routing builder
type RoutingBuilder struct {
	c     *client
	value metapb.Routing
}

// NewRoutingBuilder return a routing build
func (c *client) NewRoutingBuilder() *RoutingBuilder {
	return &RoutingBuilder{
		c:     c,
		value: metapb.Routing{},
	}
}

// Use use a cluster
func (rb *RoutingBuilder) Use(value metapb.Routing) *RoutingBuilder {
	rb.value = value
	return rb
}

// To routing to
func (rb *RoutingBuilder) To(clusterID uint64) *RoutingBuilder {
	rb.value.ClusterID = clusterID
	return rb
}

// AddCondition add condition
func (rb *RoutingBuilder) AddCondition(param metapb.Parameter, op metapb.CMP, expect string) *RoutingBuilder {
	rb.value.Conditions = append(rb.value.Conditions, metapb.Condition{
		Parameter: param,
		Cmp:       op,
		Expect:    expect,
	})
	return rb
}

// TrafficRate set traffic rate for this routing
func (rb *RoutingBuilder) TrafficRate(rate int) *RoutingBuilder {
	rb.value.TrafficRate = int32(rate)
	return rb
}

// Strategy set strategy  for this routing
func (rb *RoutingBuilder) Strategy(strategy metapb.RoutingStrategy) *RoutingBuilder {
	rb.value.Strategy = strategy
	return rb
}

// Up up this routing
func (rb *RoutingBuilder) Up() *RoutingBuilder {
	rb.value.Status = metapb.Up
	return rb
}

// Down down this routing
func (rb *RoutingBuilder) Down() *RoutingBuilder {
	rb.value.Status = metapb.Down
	return rb
}

// Name routing name
func (rb *RoutingBuilder) Name(name string) *RoutingBuilder {
	rb.value.Name = name
	return rb
}

// API set routing API
func (rb *RoutingBuilder) API(api uint64) *RoutingBuilder {
	rb.value.API = api
	return rb
}

// Commit commit
func (rb *RoutingBuilder) Commit() (uint64, error) {
	err := pb.ValidateRouting(&rb.value)
	if err != nil {
		return 0, err
	}

	return rb.c.putRouting(rb.value)
}

// Build build
func (rb *RoutingBuilder) Build() (*rpcpb.PutRoutingReq, error) {
	err := pb.ValidateRouting(&rb.value)
	if err != nil {
		return nil, err
	}

	return &rpcpb.PutRoutingReq{
		Routing: rb.value,
	}, nil
}
