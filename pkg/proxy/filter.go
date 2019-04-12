package proxy

import (
	"net/http"
	"time"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/fagongzi/log"
	"github.com/valyala/fasthttp"
	"golang.org/x/time/rate"
)

func (f *Proxy) doPreFilters(requestTag string, c filter.Context, filters ...filter.Filter) (filterName string, statusCode int, err error) {
	for _, f := range filters {
		filterName = f.Name()

		statusCode, err = f.Pre(c)
		if nil != err {
			return filterName, statusCode, err
		}

		if statusCode == filter.BreakFilterChainCode {
			log.Debugf("%s: break pre filter chain by filter %s",
				requestTag,
				filterName)
			return filterName, statusCode, err
		}
	}

	return "", http.StatusOK, nil
}

func (f *Proxy) doPostFilters(requestTag string, c filter.Context, filters ...filter.Filter) (filterName string, statusCode int, err error) {
	l := len(filters)
	for i := l - 1; i >= 0; i-- {
		f := filters[i]
		statusCode, err = f.Post(c)
		if nil != err {
			return filterName, statusCode, err
		}

		if statusCode == filter.BreakFilterChainCode {
			log.Debugf("%s: break post filter chain by filter %s",
				requestTag,
				filterName)
			return filterName, statusCode, err
		}
	}

	return "", http.StatusOK, nil
}

func (f *Proxy) doPostErrFilters(c filter.Context, filters ...filter.Filter) {
	l := len(filters)
	for i := l - 1; i >= 0; i-- {
		f := filters[i]
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
	if c.result.api.cb != nil {
		return c.result.api.id
	}

	return c.result.dest.id
}

func (c *proxyContext) rateLimiter() *rate.Limiter {
	if c.result.api.limiter != nil {
		return c.result.api.limiter
	}

	return c.result.dest.limiter
}

func (c *proxyContext) circuitBreaker() (*metapb.CircuitBreaker, *util.RateBarrier) {
	if c.result.api.cb != nil {
		return c.result.api.cb, c.result.api.barrier
	}

	return c.result.dest.cb, c.result.dest.barrier
}

func (c *proxyContext) circuitStatus() metapb.CircuitStatus {
	if c.result.api.cb != nil {
		return c.result.api.getCircuitStatus()
	}

	return c.result.dest.getCircuitStatus()
}

func (c *proxyContext) changeCircuitStatusToClose() {
	if c.result.api.cb != nil {
		c.result.api.circuitToClose()
	}

	c.result.dest.circuitToClose()
}

func (c *proxyContext) changeCircuitStatusToOpen() {
	if c.result.api.cb != nil {
		c.result.api.circuitToOpen()
	}

	c.result.dest.circuitToOpen()
}
