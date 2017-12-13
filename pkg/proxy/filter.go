package proxy

import (
	"net/http"
	"time"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/valyala/fasthttp"
)

func (f *Proxy) doPreFilters(c filter.Context) (filterName string, statusCode int, err error) {
	for iter := f.filters.Front(); iter != nil; iter = iter.Next() {
		f, _ := iter.Value.(filter.Filter)
		filterName = f.Name()

		statusCode, err = f.Pre(c)
		if nil != err {
			return filterName, statusCode, err
		}
	}

	return "", http.StatusOK, nil
}

func (f *Proxy) doPostFilters(c filter.Context) (filterName string, statusCode int, err error) {
	for iter := f.filters.Back(); iter != nil; iter = iter.Prev() {
		f, _ := iter.Value.(filter.Filter)

		statusCode, err = f.Post(c)
		if nil != err {
			return filterName, statusCode, err
		}
	}

	return "", http.StatusOK, nil
}

func (f *Proxy) doPostErrFilters(c filter.Context) {
	for iter := f.filters.Back(); iter != nil; iter = iter.Prev() {
		f, _ := iter.Value.(filter.Filter)

		f.PostErr(c)
	}
}

type proxyContext struct {
	startAt    time.Time
	endAt      time.Time
	result     *dispathNode
	forwardReq *fasthttp.Request
	originCtx  *fasthttp.RequestCtx
	rt         *dispatcher
}

func newContext(rt *dispatcher, originCtx *fasthttp.RequestCtx, forwardReq *fasthttp.Request, result *dispathNode) filter.Context {
	return &proxyContext{
		result:     result,
		originCtx:  originCtx,
		forwardReq: forwardReq,
		rt:         rt,
		startAt:    time.Now(),
	}
}

func (c *proxyContext) StartAt() time.Time {
	return c.startAt
}

func (c *proxyContext) EndAt() time.Time {
	return c.endAt
}

func (c *proxyContext) DispatchNode() *metapb.DispatchNode {
	return c.result.node.meta
}

func (c *proxyContext) API() *metapb.API {
	return c.result.api.meta
}

func (c *proxyContext) Server() *metapb.Server {
	return c.result.dest.meta
}

func (c *proxyContext) ForwardRequest() *fasthttp.Request {
	return c.forwardReq
}

func (c *proxyContext) Response() *fasthttp.Response {
	return c.result.res
}

func (c *proxyContext) OriginRequest() *fasthttp.RequestCtx {
	return c.originCtx
}

func (c *proxyContext) Analysis() *util.Analysis {
	return c.rt.analysiser
}

func (c *proxyContext) setEndAt(endAt time.Time) {
	c.endAt = endAt
}

func (c *proxyContext) validateRequest() bool {
	return c.result.node.validate(c.ForwardRequest())
}

func (c *proxyContext) allowWithBlacklist(ip string) bool {
	return c.result.api.allowWithBlacklist(ip)
}

func (c *proxyContext) allowWithWhitelist(ip string) bool {
	return c.result.api.allowWithWhitelist(ip)
}

func (c *proxyContext) circuitStatus() metapb.CircuitStatus {
	return c.result.dest.circuit
}

func (c *proxyContext) changeCircuitStatusToClose() {
	c.result.dest.circuitToClose()
}

func (c *proxyContext) changeCircuitStatusToOpen() {
	c.result.dest.circuitToOpen()
}
