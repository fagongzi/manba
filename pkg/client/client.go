package client

import (
	"fmt"
	"io"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/pb/rpcpb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// Client gateway client
type Client interface {
	PutCluster(cluster metapb.Cluster) (uint64, error)
	RemoveCluster(id uint64) error
	GetCluster(id uint64) (*metapb.Cluster, error)
	GetClusterList(fn func(*metapb.Cluster) bool) error

	PutServer(server metapb.Server) (uint64, error)
	RemoveServer(id uint64) error
	GetServer(id uint64) (*metapb.Server, error)
	GetServerList(fn func(*metapb.Server) bool) error
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

func (c *client) RemoveCluster(id uint64) error {
	_, err := c.metaC.RemoveCluster(context.Background(), &rpcpb.RemoveClusterReq{
		ID: id,
	}, grpc.FailFast(true))
	if err != nil {
		return err
	}

	return nil
}

func (c *client) GetCluster(id uint64) (*metapb.Cluster, error) {
	rsp, err := c.metaC.GetCluster(context.Background(), &rpcpb.GetClusterReq{
		ID: id,
	}, grpc.FailFast(true))
	if err != nil {
		return nil, err
	}

	return rsp.Cluster, nil
}

func (c *client) GetClusterList(fn func(*metapb.Cluster) bool) error {
	stream, err := c.metaC.GetClusterList(context.Background(), &rpcpb.GetClusterListReq{}, grpc.FailFast(true))
	if err != nil {
		return err
	}

	for {
		c, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		next := fn(c)
		if !next {
			err = stream.(grpc.ClientStream).CloseSend()
			if err != nil {
				return err
			}
		}
	}
}

func (c *client) PutServer(server metapb.Server) (uint64, error) {
	rsp, err := c.metaC.PutServer(context.Background(), &rpcpb.PutServerReq{
		Server: server,
	}, grpc.FailFast(true))
	if err != nil {
		return 0, err
	}

	return rsp.ID, nil
}

func (c *client) RemoveServer(id uint64) error {
	_, err := c.metaC.RemoveServer(context.Background(), &rpcpb.RemoveServerReq{
		ID: id,
	}, grpc.FailFast(true))
	if err != nil {
		return err
	}

	return nil
}

func (c *client) GetServer(id uint64) (*metapb.Server, error) {
	rsp, err := c.metaC.GetServer(context.Background(), &rpcpb.GetServerReq{
		ID: id,
	}, grpc.FailFast(true))
	if err != nil {
		return nil, err
	}

	return rsp.Server, nil
}

func (c *client) GetServerList(fn func(*metapb.Server) bool) error {
	stream, err := c.metaC.GetServerList(context.Background(), &rpcpb.GetServerListReq{}, grpc.FailFast(true))
	if err != nil {
		return err
	}

	for {
		c, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		next := fn(c)
		if !next {
			err = stream.(grpc.ClientStream).CloseSend()
			if err != nil {
				return err
			}
		}
	}
}

func createConn(target string, opts *options) (*grpc.ClientConn, error) {
	var grpcOptions []grpc.DialOption

	grpcOptions = append(grpcOptions, grpc.WithInsecure())
	grpcOptions = append(grpcOptions, grpc.WithTimeout(opts.timeout))
	grpcOptions = append(grpcOptions, grpc.WithBlock())
	grpcOptions = append(grpcOptions, grpc.WithBalancer(grpc.RoundRobin(opts.resolver)))

	return grpc.Dial(target, grpcOptions...)
}
