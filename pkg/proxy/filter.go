package proxy

import (
	"net/http"
	"time"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/gateway/pkg/model"
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

func (c *proxyContext) GetStartAt() time.Time {
	return c.startAt
}

func (c *proxyContext) GetEndAt() time.Time {
	return c.endAt
}
func (c *proxyContext) SetEndAt(endAt time.Time) {
	c.endAt = endAt
}

func (c *proxyContext) GetProxyServerAddr() string {
	return c.result.dest.meta.Addr
}

func (c *proxyContext) GetProxyOuterRequest() *fasthttp.Request {
	return c.forwardReq
}

func (c *proxyContext) GetProxyResponse() *fasthttp.Response {
	return c.result.res
}

func (c *proxyContext) NeedMerge() bool {
	return c.result.merge
}

func (c *proxyContext) GetOriginRequestCtx() *fasthttp.RequestCtx {
	return c.originCtx
}

func (c *proxyContext) GetMaxQPS() int {
	return c.result.dest.meta.MaxQPS
}

func (c *proxyContext) ValidateProxyOuterRequest() bool {
	return c.result.node.Validate(c.GetProxyOuterRequest())
}

func (c *proxyContext) InBlacklist(ip string) bool {
	return c.result.api.AccessCheckBlacklist(ip)
}

func (c *proxyContext) InWhitelist(ip string) bool {
	return c.result.api.AccessCheckWhitelist(ip)
}

func (c *proxyContext) GetCircuitBreaker() *model.CircuitBreaker {
	return c.result.dest.meta.CircuitBreaker
}

func (c *proxyContext) IsCircuitOpen() bool {
	return c.result.dest.isCircuitStatus(model.CircuitOpen)
}

func (c *proxyContext) IsCircuitHalf() bool {
	return c.result.dest.isCircuitStatus(model.CircuitHalf)
}

func (c *proxyContext) ChangeCircuitStatusToClose() {
	c.result.dest.circuitToClose()
}

func (c *proxyContext) ChangeCircuitStatusToOpen() {
	c.result.dest.circuitToOpen()
}

func (c *proxyContext) GetAnalysis() *model.Analysis {
	return c.rt.analysiser
}
