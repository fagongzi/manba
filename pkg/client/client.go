package client

import (
	"fmt"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/pb/rpcpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Client gateway client
type Client interface {
	PutCluster(cluster metapb.Cluster) (uint64, error)
}

// NewClient returns a gateway proxy
func NewClient(opts ...Option) (Client, error) {
	clientOptions := &options{}
	for _, opt := range opts {
		opt(clientOptions)
	}

	return newDiscoveryClient(clientOptions)
}

type client struct {
	opts  *options
	metaC rpcpb.MetaServiceClient
}

func newDiscoveryClient(opts *options) (*client, error) {
	conn, err := createConn(fmt.Sprintf("%s/%s", opts.prefix, rpcpb.ServiceMeta), opts)
	if err != nil {
		return nil, err
	}

	return &client{
		opts:  opts,
		metaC: rpcpb.NewMetaServiceClient(conn),
	}, nil
}

func (c *client) PutCluster(cluster metapb.Cluster) (uint64, error) {
	rsp, err := c.metaC.PutCluster(context.Background(), &rpcpb.PutClusterReq{
		Cluster: cluster,
	}, grpc.FailFast(true))
	if err != nil {
		return 0, err
	}

	return rsp.ID, nil
}

func createConn(target string, opts *options) (*grpc.ClientConn, error) {
	var grpcOptions []grpc.DialOption

	grpcOptions = append(grpcOptions, grpc.WithInsecure())
	grpcOptions = append(grpcOptions, grpc.WithTimeout(opts.timeout))
	grpcOptions = append(grpcOptions, grpc.WithBlock())
	grpcOptions = append(grpcOptions, grpc.WithBalancer(grpc.RoundRobin(opts.resolver)))

	return grpc.Dial(target, grpcOptions...)
}
