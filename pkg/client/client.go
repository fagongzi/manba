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

	PutAPI(api metapb.API) (uint64, error)
	RemoveAPI(id uint64) error
	GetAPI(id uint64) (*metapb.API, error)
	GetAPIList(fn func(*metapb.API) bool) error

	PutRouting(routing metapb.Routing) (uint64, error)
	RemoveRouting(id uint64) error
	GetRouting(id uint64) (*metapb.Routing, error)
	GetRoutingList(fn func(*metapb.Routing) bool) error
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
			return nil
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
			return nil
		}
	}
}

func (c *client) PutAPI(api metapb.API) (uint64, error) {
	rsp, err := c.metaC.PutAPI(context.Background(), &rpcpb.PutAPIReq{
		API: api,
	}, grpc.FailFast(true))
	if err != nil {
		return 0, err
	}

	return rsp.ID, nil
}

func (c *client) RemoveAPI(id uint64) error {
	_, err := c.metaC.RemoveAPI(context.Background(), &rpcpb.RemoveAPIReq{
		ID: id,
	}, grpc.FailFast(true))
	if err != nil {
		return err
	}

	return nil
}

func (c *client) GetAPI(id uint64) (*metapb.API, error) {
	rsp, err := c.metaC.GetAPI(context.Background(), &rpcpb.GetAPIReq{
		ID: id,
	}, grpc.FailFast(true))
	if err != nil {
		return nil, err
	}

	return rsp.API, nil
}

func (c *client) GetAPIList(fn func(*metapb.API) bool) error {
	stream, err := c.metaC.GetAPIList(context.Background(), &rpcpb.GetAPIListReq{}, grpc.FailFast(true))
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
			return nil
		}
	}
}

func (c *client) PutRouting(routing metapb.Routing) (uint64, error) {
	rsp, err := c.metaC.PutRouting(context.Background(), &rpcpb.PutRoutingReq{
		Routing: routing,
	}, grpc.FailFast(true))
	if err != nil {
		return 0, err
	}

	return rsp.ID, nil
}

func (c *client) RemoveRouting(id uint64) error {
	_, err := c.metaC.RemoveRouting(context.Background(), &rpcpb.RemoveRoutingReq{
		ID: id,
	}, grpc.FailFast(true))
	if err != nil {
		return err
	}

	return nil
}

func (c *client) GetRouting(id uint64) (*metapb.Routing, error) {
	rsp, err := c.metaC.GetRouting(context.Background(), &rpcpb.GetRoutingReq{
		ID: id,
	}, grpc.FailFast(true))
	if err != nil {
		return nil, err
	}

	return rsp.Routing, nil
}

func (c *client) GetRoutingList(fn func(*metapb.Routing) bool) error {
	stream, err := c.metaC.GetRoutingList(context.Background(), &rpcpb.GetRoutingListReq{}, grpc.FailFast(true))
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
			return nil
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
