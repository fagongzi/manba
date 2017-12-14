package client

import (
	"io"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/pb/rpcpb"
	"github.com/fagongzi/grpcx"
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

// NewClient returns a gateway client, using direct address
func NewClient(timeout time.Duration, addrs ...string) (Client, error) {
	return newDiscoveryClient(grpcx.WithDirectAddresses(addrs...),
		grpcx.WithTimeout(timeout))
}

// NewClientWithEtcdDiscovery returns a gateway client, using etcd service discovery
func NewClientWithEtcdDiscovery(prefix string, timeout time.Duration, etcdAddrs ...string) (Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdAddrs,
		DialTimeout: time.Second * 10,
	})
	if err != nil {
		return nil, err
	}

	return newDiscoveryClient(grpcx.WithEtcdServiceDiscovery(prefix, cli),
		grpcx.WithTimeout(timeout))
}

type client struct {
	clients *grpcx.GRPCClient
}

func newDiscoveryClient(opts ...grpcx.ClientOption) (*client, error) {
	clients := grpcx.NewGRPCClient(func(name string, raw *grpc.ClientConn) interface{} {
		if name == rpcpb.ServiceMeta {
			return rpcpb.NewMetaServiceClient(raw)
		}

		return nil
	}, opts...)

	return &client{
		clients: clients,
	}, nil
}

func (c *client) getMetaClient() (rpcpb.MetaServiceClient, error) {
	cli, err := c.clients.GetServiceClient(rpcpb.ServiceMeta)
	if err != nil {
		return nil, err
	}

	return cli.(rpcpb.MetaServiceClient), nil
}

func (c *client) PutCluster(cluster metapb.Cluster) (uint64, error) {
	meta, err := c.getMetaClient()
	if err != nil {
		return 0, err
	}

	rsp, err := meta.PutCluster(context.Background(), &rpcpb.PutClusterReq{
		Cluster: cluster,
	}, grpc.FailFast(true))
	if err != nil {
		return 0, err
	}

	return rsp.ID, nil
}

func (c *client) RemoveCluster(id uint64) error {
	meta, err := c.getMetaClient()
	if err != nil {
		return err
	}

	_, err = meta.RemoveCluster(context.Background(), &rpcpb.RemoveClusterReq{
		ID: id,
	}, grpc.FailFast(true))
	if err != nil {
		return err
	}

	return nil
}

func (c *client) GetCluster(id uint64) (*metapb.Cluster, error) {
	meta, err := c.getMetaClient()
	if err != nil {
		return nil, err
	}

	rsp, err := meta.GetCluster(context.Background(), &rpcpb.GetClusterReq{
		ID: id,
	}, grpc.FailFast(true))
	if err != nil {
		return nil, err
	}

	return rsp.Cluster, nil
}

func (c *client) GetClusterList(fn func(*metapb.Cluster) bool) error {
	meta, err := c.getMetaClient()
	if err != nil {
		return err
	}

	stream, err := meta.GetClusterList(context.Background(), &rpcpb.GetClusterListReq{}, grpc.FailFast(true))
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
	meta, err := c.getMetaClient()
	if err != nil {
		return 0, err
	}

	rsp, err := meta.PutServer(context.Background(), &rpcpb.PutServerReq{
		Server: server,
	}, grpc.FailFast(true))
	if err != nil {
		return 0, err
	}

	return rsp.ID, nil
}

func (c *client) RemoveServer(id uint64) error {
	meta, err := c.getMetaClient()
	if err != nil {
		return err
	}

	_, err = meta.RemoveServer(context.Background(), &rpcpb.RemoveServerReq{
		ID: id,
	}, grpc.FailFast(true))
	if err != nil {
		return err
	}

	return nil
}

func (c *client) GetServer(id uint64) (*metapb.Server, error) {
	meta, err := c.getMetaClient()
	if err != nil {
		return nil, err
	}

	rsp, err := meta.GetServer(context.Background(), &rpcpb.GetServerReq{
		ID: id,
	}, grpc.FailFast(true))
	if err != nil {
		return nil, err
	}

	return rsp.Server, nil
}

func (c *client) GetServerList(fn func(*metapb.Server) bool) error {
	meta, err := c.getMetaClient()
	if err != nil {
		return err
	}

	stream, err := meta.GetServerList(context.Background(), &rpcpb.GetServerListReq{}, grpc.FailFast(true))
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
	meta, err := c.getMetaClient()
	if err != nil {
		return 0, err
	}

	rsp, err := meta.PutAPI(context.Background(), &rpcpb.PutAPIReq{
		API: api,
	}, grpc.FailFast(true))
	if err != nil {
		return 0, err
	}

	return rsp.ID, nil
}

func (c *client) RemoveAPI(id uint64) error {
	meta, err := c.getMetaClient()
	if err != nil {
		return err
	}

	_, err = meta.RemoveAPI(context.Background(), &rpcpb.RemoveAPIReq{
		ID: id,
	}, grpc.FailFast(true))
	if err != nil {
		return err
	}

	return nil
}

func (c *client) GetAPI(id uint64) (*metapb.API, error) {
	meta, err := c.getMetaClient()
	if err != nil {
		return nil, err
	}

	rsp, err := meta.GetAPI(context.Background(), &rpcpb.GetAPIReq{
		ID: id,
	}, grpc.FailFast(true))
	if err != nil {
		return nil, err
	}

	return rsp.API, nil
}

func (c *client) GetAPIList(fn func(*metapb.API) bool) error {
	meta, err := c.getMetaClient()
	if err != nil {
		return err
	}

	stream, err := meta.GetAPIList(context.Background(), &rpcpb.GetAPIListReq{}, grpc.FailFast(true))
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
	meta, err := c.getMetaClient()
	if err != nil {
		return 0, err
	}

	rsp, err := meta.PutRouting(context.Background(), &rpcpb.PutRoutingReq{
		Routing: routing,
	}, grpc.FailFast(true))
	if err != nil {
		return 0, err
	}

	return rsp.ID, nil
}

func (c *client) RemoveRouting(id uint64) error {
	meta, err := c.getMetaClient()
	if err != nil {
		return err
	}

	_, err = meta.RemoveRouting(context.Background(), &rpcpb.RemoveRoutingReq{
		ID: id,
	}, grpc.FailFast(true))
	if err != nil {
		return err
	}

	return nil
}

func (c *client) GetRouting(id uint64) (*metapb.Routing, error) {
	meta, err := c.getMetaClient()
	if err != nil {
		return nil, err
	}

	rsp, err := meta.GetRouting(context.Background(), &rpcpb.GetRoutingReq{
		ID: id,
	}, grpc.FailFast(true))
	if err != nil {
		return nil, err
	}

	return rsp.Routing, nil
}

func (c *client) GetRoutingList(fn func(*metapb.Routing) bool) error {
	meta, err := c.getMetaClient()
	if err != nil {
		return err
	}

	stream, err := meta.GetRoutingList(context.Background(), &rpcpb.GetRoutingListReq{}, grpc.FailFast(true))
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
