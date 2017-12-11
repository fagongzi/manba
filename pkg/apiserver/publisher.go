package apiserver

import (
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	etcdnaming "github.com/coreos/etcd/clientv3/naming"
	"golang.org/x/net/context"
	"google.golang.org/grpc/naming"
)

// ServicePublisher a service publish
type ServicePublisher interface {
	Publish(service string, meta naming.Update) error
}

type etcdServicePublisher struct {
	prefix   string
	ttl      int64
	timeout  time.Duration
	client   *clientv3.Client
	resolver *etcdnaming.GRPCResolver
}

func newEtcdServicePublisher(client *clientv3.Client, prefix string, ttl int64, timeout time.Duration) ServicePublisher {
	return &etcdServicePublisher{
		prefix:  prefix,
		ttl:     ttl,
		timeout: timeout,
		client:  client,
		resolver: &etcdnaming.GRPCResolver{
			Client: client,
		},
	}
}

func (sp *etcdServicePublisher) Publish(service string, meta naming.Update) error {
	lessor := clientv3.NewLease(sp.client)
	defer lessor.Close()

	ctx, cancel := context.WithTimeout(sp.client.Ctx(), sp.timeout)
	leaseResp, err := lessor.Grant(ctx, sp.ttl)
	cancel()
	if err != nil {
		return err
	}

	ctx, cancel = context.WithTimeout(sp.client.Ctx(), sp.timeout)
	defer cancel()

	return sp.resolver.Update(ctx, fmt.Sprintf("%s/%s", sp.prefix, service), meta, clientv3.WithLease(clientv3.LeaseID(leaseResp.ID)))
}
