package proxy

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/fagongzi/log"
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

const (
	// TimerPrefix timer prefix
	TimerPrefix = "Circuit-"
)

type proxyContext struct {
	startAt   int64
	endAt     int64
	result    *model.RouteResult
	outerReq  *fasthttp.Request
	originCtx *fasthttp.RequestCtx
	rt        *model.RouteTable
}

func newContext(rt *model.RouteTable, originCtx *fasthttp.RequestCtx, outerReq *fasthttp.Request, result *model.RouteResult) filter.Context {
	return &proxyContext{
		result:    result,
		originCtx: originCtx,
		outerReq:  outerReq,
		rt:        rt,
	}
}

func (c *proxyContext) GetStartAt() int64 {
	return c.startAt
}

func (c *proxyContext) SetStartAt(startAt int64) {
	c.startAt = startAt
}

func (c *proxyContext) GetEndAt() int64 {
	return c.endAt
}

func (c *proxyContext) SetEndAt(endAt int64) {
	c.endAt = endAt
}

func (c *proxyContext) GetProxyServerAddr() string {
	return c.result.Svr.Addr
}

func (c *proxyContext) GetProxyOuterRequest() *fasthttp.Request {
	return c.outerReq
}

func (c *proxyContext) GetProxyResponse() *fasthttp.Response {
	return c.result.Res
}

func (c *proxyContext) NeedMerge() bool {
	return c.result.Merge
}

func (c *proxyContext) GetOriginRequestCtx() *fasthttp.RequestCtx {
	return c.originCtx
}

func (c *proxyContext) GetMaxQPS() int {
	return c.result.Svr.MaxQPS
}

func (c *proxyContext) ValidateProxyOuterRequest() bool {
	return c.result.Node.Validate(c.GetProxyOuterRequest())
}

func (c *proxyContext) InBlacklist(ip string) bool {
	return c.result.API.AccessCheckBlacklist(ip)
}

func (c *proxyContext) InWhitelist(ip string) bool {
	return c.result.API.AccessCheckWhitelist(ip)
}

func (c *proxyContext) IsCircuitOpen() bool {
	return c.result.Svr.GetCircuit() == model.CircuitOpen
}

func (c *proxyContext) IsCircuitHalf() bool {
	return c.result.Svr.GetCircuit() == model.CircuitHalf
}

func (c *proxyContext) GetOpenToCloseFailureRate() int {
	return c.result.Svr.OpenToCloseFailureRate
}
func (c *proxyContext) GetOpenToCloseCollectSeconds() int {
	return c.result.Svr.OpenToCloseCollectSeconds
}

func (c *proxyContext) GetHalfTrafficRate() int {
	return c.result.Svr.HalfTrafficRate
}

func (c *proxyContext) GetHalfToOpenSucceedRate() int {
	return c.result.Svr.HalfToOpenSucceedRate
}

func (c *proxyContext) ChangeCircuitStatusToClose() {
	server := c.result.Svr

	server.Lock()

	if server.GetCircuit() == model.CircuitClose {
		server.UnLock()
		return
	}

	server.CloseCircuit()

	log.Warnf("filter: circuit server <%s> change to close", server.Addr)

	c.rt.GetTimeWheel().Schedule(time.Second*time.Duration(server.CloseToHalfSeconds), c.changeCircuitStatusToHalf, getKey(server.Addr))

	server.UnLock()
}

func (c *proxyContext) ChangeCircuitStatusToOpen() {
	server := c.result.Svr

	server.Lock()

	if server.GetCircuit() == model.CircuitOpen || server.GetCircuit() != model.CircuitHalf {
		server.UnLock()
		return
	}

	server.OpenCircuit()

	log.Warnf("filter: circuit server <%s> change to open", server.Addr)

	server.UnLock()
}

func (c *proxyContext) changeCircuitStatusToHalf(key interface{}) {
	addr := getAddr(key.(string))
	server := c.rt.GetServer(addr)

	if nil != server {
		server.Lock()
		server.HalfCircuit()
		server.UnLock()

		log.Warnf("filter: circuit server <%s> change to half", server.Addr)
	}
}

func (c *proxyContext) RecordMetricsForRequest() {
	c.rt.GetAnalysis().Request(c.GetProxyServerAddr())
}

func (c *proxyContext) RecordMetricsForResponse() {
	c.rt.GetAnalysis().Response(c.GetProxyServerAddr(), c.endAt-c.startAt)
}

func (c *proxyContext) RecordMetricsForFailure() {
	c.rt.GetAnalysis().Failure(c.GetProxyServerAddr())
}

func (c *proxyContext) RecordMetricsForReject() {
	c.rt.GetAnalysis().Reject(c.GetProxyServerAddr())
}

func (c *proxyContext) GetRecentlyRequestCount(sec int) int {
	return c.rt.GetAnalysis().GetRecentlyRequestCount(c.GetProxyServerAddr(), sec)
}

func (c *proxyContext) GetRecentlyRequestSuccessedCount(sec int) int {
	return c.rt.GetAnalysis().GetRecentlyRequestSuccessedCount(c.GetProxyServerAddr(), sec)
}

func (c *proxyContext) GetRecentlyRequestFailureCount(sec int) int {
	return c.rt.GetAnalysis().GetRecentlyRequestFailureCount(c.GetProxyServerAddr(), sec)
}

func getKey(addr string) string {
	return fmt.Sprintf("%s%s", TimerPrefix, addr)
}

func getAddr(key string) string {
	info := strings.Split(key, "-")
	return info[1]
}
