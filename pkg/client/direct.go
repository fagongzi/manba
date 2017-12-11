package client

import (
	"sync"

	"golang.org/x/net/context"
	"google.golang.org/grpc/naming"
)

type localResolver struct {
	sync.RWMutex

	firstCalled bool
	index       int
	addrs       []string
	ctx         context.Context
	cancel      context.CancelFunc
}

func newLocalResolver(addrs ...string) naming.Resolver {
	ctx, cancel := context.WithCancel(context.Background())
	return &localResolver{
		addrs:  addrs,
		ctx:    ctx,
		cancel: cancel,
	}
}

func (lr *localResolver) Resolve(target string) (naming.Watcher, error) {
	return lr, nil
}

func (lr *localResolver) Next() ([]*naming.Update, error) {
	lr.Lock()
	defer lr.Unlock()

	if !lr.firstCalled {
		return lr.firstNext()
	}

	// block
	<-lr.ctx.Done()
	return nil, lr.ctx.Err()
}

func (lr *localResolver) Close() {
	lr.cancel()
}

func (lr *localResolver) firstNext() ([]*naming.Update, error) {
	var values []*naming.Update
	for _, addr := range lr.addrs {
		values = append(values, &naming.Update{
			Op:   naming.Add,
			Addr: addr,
		})
	}
	return values, nil
}
