package proxy

import (
	"sync"

	"github.com/fagongzi/goetty"
)

var (
	renderPool       sync.Pool
	contextPool      sync.Pool
	dispathNodePool  sync.Pool
	multiContextPool sync.Pool
	bytesPool        = goetty.NewSyncPool(2, 1024*1024*5, 2)

	emptyRender      = render{}
	emptyContext     = proxyContext{}
	emptyDispathNode = dispathNode{}
)

func acquireMultiContext() *multiContext {
	v := multiContextPool.Get()
	if v == nil {
		return &multiContext{}
	}

	return v.(*multiContext)
}

func releaseMultiContext(value *multiContext) {
	if value != nil {
		value.reset()
		multiContextPool.Put(value)
	}
}

func acquireDispathNode() *dispathNode {
	v := dispathNodePool.Get()
	if v == nil {
		return &dispathNode{}
	}

	return v.(*dispathNode)
}

func releaseDispathNode(value *dispathNode) {
	if value != nil {
		value.reset()
		dispathNodePool.Put(value)
	}
}

func acquireContext() *proxyContext {
	v := contextPool.Get()
	if v == nil {
		return &proxyContext{}
	}

	return v.(*proxyContext)
}

func releaseContext(value *proxyContext) {
	if value != nil {
		value.reset()
		contextPool.Put(value)
	}
}

func acquireRender() *render {
	v := renderPool.Get()
	if v == nil {
		return &render{}
	}

	return v.(*render)
}

func releaseRender(value *render) {
	if value != nil {
		value.reset()
		renderPool.Put(value)
	}
}
