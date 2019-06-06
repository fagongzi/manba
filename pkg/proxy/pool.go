package proxy

import (
	"sync"

	"github.com/fagongzi/gateway/pkg/expr"
	"github.com/fagongzi/goetty"
)

var (
	renderPool       sync.Pool
	contextPool      sync.Pool
	dispatchNodePool  sync.Pool
	multiContextPool sync.Pool
	wgPool           sync.Pool
	exprCtxPool      sync.Pool
	bytesPool        = goetty.NewSyncPool(2, 1024*1024*5, 2)

	emptyRender      = render{}
	emptyContext     = proxyContext{}
	emptyDispathNode = dispatchNode{}
)

func acquireWG() *sync.WaitGroup {
	v := wgPool.Get()
	if v == nil {
		return &sync.WaitGroup{}
	}

	return v.(*sync.WaitGroup)
}

func releaseWG(value *sync.WaitGroup) {
	if value != nil {
		wgPool.Put(value)
	}
}

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

func acquireDispathNode() *dispatchNode {
	v := dispatchNodePool.Get()
	if v == nil {
		return &dispatchNode{}
	}

	return v.(*dispatchNode)
}

func releaseDispathNode(value *dispatchNode) {
	if value != nil {
		value.reset()
		dispatchNodePool.Put(value)
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

func acquireExprCtx() *expr.Ctx {
	v := exprCtxPool.Get()
	if v == nil {
		return &expr.Ctx{
			Params: make(map[string][]byte),
		}
	}

	return v.(*expr.Ctx)
}

func releaseExprCtx(value *expr.Ctx) {
	if value != nil {
		value.Reset()
		exprCtxPool.Put(value)
	}
}
