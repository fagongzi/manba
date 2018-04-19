package grpcx

import (
	"time"

	"github.com/coreos/etcd/clientv3"
	etcdnaming "github.com/coreos/etcd/clientv3/naming"
	"github.com/fagongzi/log"
	"github.com/labstack/echo"
	"google.golang.org/grpc/naming"
)

// ServerOption service side option
type ServerOption func(*serverOptions)

type serverOptions struct {
	publisher Publisher
	httpAddr  string
	httpSetup func(*echo.Echo)
}

// WithEtcdPublisher use etcd to publish service
func WithEtcdPublisher(client *clientv3.Client, prefix string, ttl int64, timeout time.Duration) ServerOption {
	return func(opts *serverOptions) {
		p, err := newEtcdPublisher(client, prefix, ttl, timeout)
		if err != nil {
			log.Fatalf("rpc: use etcd service publish failed, errors:\n%+v",
				err)
		}
		opts.publisher = p
	}
}

// WithHTTPServer with http server
func WithHTTPServer(addr string, httpSetup func(*echo.Echo)) ServerOption {
	return func(opts *serverOptions) {
		opts.httpAddr = addr
		opts.httpSetup = httpSetup
	}
}

// ClientOption is client create option
type ClientOption func(*clientOptions)

type clientOptions struct {
	prefix   string
	resolver naming.Resolver
	timeout  time.Duration
}

// WithEtcdServiceDiscovery returns a etcd discovery option
func WithEtcdServiceDiscovery(prefix string, client *clientv3.Client) ClientOption {
	return func(opts *clientOptions) {
		opts.prefix = prefix
		opts.resolver = &etcdnaming.GRPCResolver{
			Client: client,
		}
	}
}

// WithDirectAddresses returns a direct addresses option
func WithDirectAddresses(addrs ...string) ClientOption {
	return func(opts *clientOptions) {
		opts.resolver = newLocalResolver(addrs...)
	}
}

// WithTimeout returns a timeout option
func WithTimeout(timeout time.Duration) ClientOption {
	return func(opts *clientOptions) {
		opts.timeout = timeout
	}
}
