package proxy

import (
	"github.com/fagongzi/gateway/pkg/filter"
)

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailers",
	"Transfer-Encoding",
}

// HeadersFilter HeadersFilter
type HeadersFilter struct {
	filter.BaseFilter
}

func newHeadersFilter() filter.Filter {
	return &HeadersFilter{}
}

// Init init filter
func (f *HeadersFilter) Init(cfg string) error {
	return nil
}

// Name return name of this filter
func (f *HeadersFilter) Name() string {
	return FilterHeader
}

// Pre execute before proxy
func (f *HeadersFilter) Pre(c filter.Context) (statusCode int, err error) {
	for _, h := range hopHeaders {
		c.ForwardRequest().Header.Del(h)
	}

	c.ForwardRequest().Header.SetHost(c.Server().Addr)
	return f.BaseFilter.Pre(c)
}

// Post execute after proxy
func (f *HeadersFilter) Post(c filter.Context) (statusCode int, err error) {
	for _, h := range hopHeaders {
		c.Response().Header.Del(h)
	}

	// 需要合并处理的，不做header的复制，由proxy做合并
	if len(c.API().Nodes) == 1 {
		c.OriginRequest().Response.Header.Reset()
		c.Response().Header.CopyTo(&c.OriginRequest().Response.Header)
	}

	return f.BaseFilter.Post(c)
}
