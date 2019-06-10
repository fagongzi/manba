package filter

import (
	"sync"
	"time"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/valyala/fasthttp"
)

// TestContext the context for test
type TestContext struct {
	Attr                     sync.Map
	StartAtValue, EndAtValue time.Time
	OriginValue              *fasthttp.RequestCtx
	ForwardValue             *fasthttp.Request
	ResponseValue            *fasthttp.Response
	APIValue                 *metapb.API
	NodeValue                *metapb.DispatchNode
	ServerValue              *metapb.Server
	AnalysisValue            *util.Analysis
}

// StartAt returns StartAt value
func (ctx *TestContext) StartAt() time.Time {
	return ctx.StartAtValue
}

// EndAt returns EndAt value
func (ctx *TestContext) EndAt() time.Time {
	return ctx.EndAtValue
}

// OriginRequest returns OriginRequest value
func (ctx *TestContext) OriginRequest() *fasthttp.RequestCtx {
	return ctx.OriginValue
}

// ForwardRequest returns ForwardRequest value
func (ctx *TestContext) ForwardRequest() *fasthttp.Request {
	return ctx.ForwardValue
}

// Response returns Response value
func (ctx *TestContext) Response() *fasthttp.Response {
	return ctx.ResponseValue
}

// API returns API value
func (ctx *TestContext) API() *metapb.API {
	return ctx.APIValue
}

// DispatchNode returns DispatchNode value
func (ctx *TestContext) DispatchNode() *metapb.DispatchNode {
	return ctx.NodeValue
}

// Server returns Server value
func (ctx *TestContext) Server() *metapb.Server {
	return ctx.ServerValue
}

// Analysis returns Analysis value
func (ctx *TestContext) Analysis() *util.Analysis {
	return ctx.AnalysisValue
}

// SetAttr set attr
func (ctx *TestContext) SetAttr(key string, value interface{}) {
	ctx.Attr.Store(key, value)
}

// GetAttr get attr value
func (ctx *TestContext) GetAttr(key string) interface{} {
	value, _ := ctx.Attr.Load(key)
	return value
}
