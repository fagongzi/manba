package client

import (
	"time"

	"github.com/coreos/etcd/clientv3"
	etcdnaming "github.com/coreos/etcd/clientv3/naming"
	"google.golang.org/grpc/naming"
)

type options struct {
	prefix   string
	resolver naming.Resolver

	timeout time.Duration
}

// Option is client create option
type Option func(*options)

// WithEtcdServiceDiscovery returns a etcd discovery option
func WithEtcdServiceDiscovery(prefix string, client interface{}) Option {
	return func(opts *options) {
		opts.prefix = prefix
		opts.resolver = &etcdnaming.GRPCResolver{
			Client: client.(*clientv3.Client),
		}
	}
}

// WithDirectAddresses returns a direct addresses option
func WithDirectAddresses(addrs ...string) Option {
	return func(opts *options) {
		opts.resolver = newLocalResolver(addrs...)
	}
}
