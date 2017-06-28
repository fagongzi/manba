package filter

import (
	"net/http"

	"github.com/valyala/fasthttp"
)

// Context filter context
type Context interface {
	SetStartAt(startAt int64)
	SetEndAt(endAt int64)
	GetStartAt() int64
	GetEndAt() int64

	GetProxyServerAddr() string
	GetProxyOuterRequest() *fasthttp.Request
	GetProxyResponse() *fasthttp.Response
	NeedMerge() bool

	GetOriginRequestCtx() *fasthttp.RequestCtx

	GetMaxQPS() int

	ValidateProxyOuterRequest() bool

	InBlacklist(ip string) bool
	InWhitelist(ip string) bool

	IsCircuitOpen() bool
	IsCircuitHalf() bool

	GetOpenToCloseFailureRate() int
	GetHalfTrafficRate() int
	GetHalfToOpenSucceedRate() int
	GetOpenToCloseCollectSeconds() int

	ChangeCircuitStatusToClose()
	ChangeCircuitStatusToOpen()

	RecordMetricsForRequest()
	RecordMetricsForResponse()
	RecordMetricsForFailure()
	RecordMetricsForReject()

	GetRecentlyRequestSuccessedCount(sec int) int
	GetRecentlyRequestCount(sec int) int
	GetRecentlyRequestFailureCount(sec int) int
}

// Filter filter interface
type Filter interface {
	Name() string

	Pre(c Context) (statusCode int, err error)
	Post(c Context) (statusCode int, err error)
	PostErr(c Context)
}

// BaseFilter base filter support default implemention
type BaseFilter struct{}

// Pre execute before proxy
func (f BaseFilter) Pre(c Context) (statusCode int, err error) {
	return http.StatusOK, nil
}

// Post execute after proxy
func (f BaseFilter) Post(c Context) (statusCode int, err error) {
	return http.StatusOK, nil
}

// PostErr execute proxy has errors
func (f BaseFilter) PostErr(c Context) {

}
