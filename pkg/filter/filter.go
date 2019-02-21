package filter

import (
	"time"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/valyala/fasthttp"
)

const (
	// UsingCachingValue using cached value to response
	UsingCachingValue = "__using_cache_value__"
	// UsingResponse using response to response
	UsingResponse = "__using_response__"

	// BreakFilterChainCode break filter chain code
	BreakFilterChainCode = -1
)

// NewCachedValue returns a cached value
func NewCachedValue(body, contentType []byte) []byte {
	size := len(contentType) + 4 + len(body)
	data := make([]byte, size, size)
	idx := 0
	int2BytesTo(len(contentType), data[0:4])
	idx += 4
	copy(data[idx:idx+len(contentType)], contentType)
	idx += len(contentType)
	copy(data[idx:], body)
	return data
}

// ParseCachedValue returns cached value as content-type and body value
func ParseCachedValue(data []byte) ([]byte, []byte) {
	size := byte2Int(data[0:4])
	return data[4 : 4+size], data[4+size:]
}

// Context filter context
type Context interface {
	StartAt() time.Time
	EndAt() time.Time

	OriginRequest() *fasthttp.RequestCtx
	ForwardRequest() *fasthttp.Request
	Response() *fasthttp.Response

	API() *metapb.API
	DispatchNode() *metapb.DispatchNode
	Server() *metapb.Server
	Analysis() *util.Analysis

	SetAttr(key string, value interface{})
	GetAttr(key string) interface{}
}

// Filter filter interface
type Filter interface {
	Name() string
	Init(cfg string) error

	Pre(c Context) (statusCode int, err error)
	Post(c Context) (statusCode int, err error)
	PostErr(c Context)
}

// BaseFilter base filter support default implemention
type BaseFilter struct{}

// Init init filter
func (f BaseFilter) Init(cfg string) error {
	return nil
}

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

func int2BytesTo(v int, ret []byte) {
	ret[0] = byte(v >> 24)
	ret[1] = byte(v >> 16)
	ret[2] = byte(v >> 8)
	ret[3] = byte(v)
}

func byte2Int(data []byte) int {
	return int((int(data[0])&0xff)<<24 | (int(data[1])&0xff)<<16 | (int(data[2])&0xff)<<8 | (int(data[3]) & 0xff))
}
