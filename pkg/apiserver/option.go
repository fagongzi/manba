package apiserver

import (
	"time"

	"github.com/coreos/etcd/clientv3"
)

// Option option
type Option func(*options)

type options struct {
	publisher ServicePublisher
}

// WithEtcdServiceDiscovery use etcd to publish service
func WithEtcdServiceDiscovery(client interface{}, prefix string, ttl int64, timeout time.Duration) Option {
	return func(opts *options) {
		opts.publisher = newEtcdServicePublisher(client.(*clientv3.Client), prefix, ttl, timeout)
	}
}
