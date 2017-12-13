package filter

import (
	"time"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/valyala/fasthttp"
)

// Context filter context
type Context interface {
	GetStartAt() time.Time
	GetEndAt() time.Time
	SetEndAt(time.Time)

	OriginRequest() *fasthttp.RequestCtx
	ForwardRequest() *fasthttp.Request
	Response() *fasthttp.Response

	API() *metapb.API
	Server() *metapb.Server
	Analysis() *util.Analysis
	CircuitStatus() metapb.CircuitStatus

	ChangeCircuitStatusToClose()
	ChangeCircuitStatusToOpen()

	AllowWithBlacklist(ip string) bool
	AllowWithWhitelist(ip string) bool
	ValidateRequest() bool
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
