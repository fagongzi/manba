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
	for _, f := range f.filters {
		filterName = f.Name()

		statusCode, err = f.Pre(c)
		if nil != err {
			return filterName, statusCode, err
		}
	}

	return "", http.StatusOK, nil
}

func (f *Proxy) doPostFilters(c filter.Context) (filterName string, statusCode int, err error) {
	l := len(f.filters)
	for i := l - 1; i >= 0; i-- {
		f := f.filters[i]
		statusCode, err = f.Post(c)
		if nil != err {
			return filterName, statusCode, err
		}
	}

	return "", http.StatusOK, nil
}

func (f *Proxy) doPostErrFilters(c filter.Context) {
	l := len(f.filters)
	for i := l - 1; i >= 0; i-- {
		f := f.filters[i]
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

	attrs map[string]interface{}
}

func (c *proxyContext) init(rt *dispatcher, originCtx *fasthttp.RequestCtx, forwardReq *fasthttp.Request, result *dispathNode) {
	c.result = result
	c.originCtx = originCtx
	c.forwardReq = forwardReq
	c.rt = rt
	c.startAt = time.Now()
	c.attrs = make(map[string]interface{})
}

func (c *proxyContext) reset() {
	if c.forwardReq != nil {
		fasthttp.ReleaseRequest(c.forwardReq)
	}
	*c = emptyContext
}

func (c *proxyContext) SetAttr(key string, value interface{}) {
	c.attrs[key] = value
}

func (c *proxyContext) GetAttr(key string) interface{} {
	return c.attrs[key]
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

func (c *proxyContext) circuitResourceID() uint64 {
	if c.result.dest != nil {
		return c.result.dest.id
	}

	return c.result.api.id
}

func (c *proxyContext) circuitBreaker() (*metapb.CircuitBreaker, *util.RateBarrier) {
	if c.result.dest != nil {
		return c.result.dest.cb, c.result.dest.barrier
	}

	return c.result.api.cb, c.result.api.barrier
}

func (c *proxyContext) circuitStatus() metapb.CircuitStatus {
	if c.result.dest != nil {
		return c.result.dest.getCircuitStatus()
	}

	return c.result.api.getCircuitStatus()
}

func (c *proxyContext) changeCircuitStatusToClose() {
	if c.result.dest != nil {
		c.result.dest.circuitToClose()
	}

	c.result.api.circuitToClose()
}

func (c *proxyContext) changeCircuitStatusToOpen() {
	if c.result.dest != nil {
		c.result.dest.circuitToOpen()
	}

	c.result.api.circuitToOpen()
}
