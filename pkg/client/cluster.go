package client

import (
	"github.com/fagongzi/gateway/pkg/pb"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/pb/rpcpb"
)

// ClusterBuilder cluster builder
type ClusterBuilder struct {
	c     *client
	value metapb.Cluster
}

// NewClusterBuilder return a cluster build
func (c *client) NewClusterBuilder() *ClusterBuilder {
	return &ClusterBuilder{
		c:     c,
		value: metapb.Cluster{},
	}
}

// Use use a cluster
func (cb *ClusterBuilder) Use(value metapb.Cluster) *ClusterBuilder {
	cb.value = value
	return cb
}

// Name set a name
func (cb *ClusterBuilder) Name(name string) *ClusterBuilder {
	cb.value.Name = name
	return cb
}

// Loadbalance set a loadbalance
func (cb *ClusterBuilder) Loadbalance(lb metapb.LoadBalance) *ClusterBuilder {
	cb.value.LoadBalance = lb
	return cb
}

// Commit commit
func (cb *ClusterBuilder) Commit() (uint64, error) {
	err := pb.ValidateCluster(&cb.value)
	if err != nil {
		return 0, err
	}

	return cb.c.putCluster(cb.value)
}

// Build build
func (cb *ClusterBuilder) Build() (*rpcpb.PutClusterReq, error) {
	err := pb.ValidateCluster(&cb.value)
	if err != nil {
		return nil, err
	}

	return &rpcpb.PutClusterReq{
		Cluster: cb.value,
	}, nil
}
