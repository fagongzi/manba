package filter

import (
	"time"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/valyala/fasthttp"
)

// Context filter context
type Context interface {
	GetStartAt() time.Time
	GetEndAt() time.Time
	SetEndAt(time.Time)

	GetProxyServerAddr() string
	GetOriginRequestCtx() *fasthttp.RequestCtx
	GetProxyOuterRequest() *fasthttp.Request
	ValidateProxyOuterRequest() bool
	GetProxyResponse() *fasthttp.Response
	NeedMerge() bool

	GetMaxQPS() int

	IsCircuitOpen() bool
	IsCircuitHalf() bool
	ChangeCircuitStatusToClose()
	ChangeCircuitStatusToOpen()
	GetCircuitBreaker() *model.CircuitBreaker

	InBlacklist(ip string) bool
	InWhitelist(ip string) bool

	GetAnalysis() *model.Analysis
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
	return fasthttp.StatusOK, nil
}

// Post execute after proxy
func (f BaseFilter) Post(c Context) (statusCode int, err error) {
	return fasthttp.StatusOK, nil
}

// PostErr execute proxy has errors
func (f BaseFilter) PostErr(c Context) {

}
