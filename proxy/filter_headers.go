package proxy

import "github.com/fagongzi/gateway/pkg/filter"

// Hop-by-hop headers. These are removed when sent to the backend.
// http://www.w3.org/Protocols/rfc2616/rfc2616-sec13.html
var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

// HeadersFilter HeadersFilter
type HeadersFilter struct {
	filter.BaseFilter
}

func newHeadersFilter() filter.Filter {
	return &HeadersFilter{}
}

// Name return name of this filter
func (f HeadersFilter) Name() string {
	return FilterHeader
}

// Pre execute before proxy
func (f HeadersFilter) Pre(c filter.Context) (statusCode int, err error) {
	for _, h := range hopHeaders {
		c.GetProxyOuterRequest().Header.Del(h)
	}

	return f.BaseFilter.Pre(c)
}

// Post execute after proxy
func (f HeadersFilter) Post(c filter.Context) (statusCode int, err error) {
	for _, h := range hopHeaders {
		c.GetProxyResponse().Header.Del(h)
	}

	// 需要合并处理的，不做header的复制，由proxy做合并
	if !c.NeedMerge() {
		c.GetOriginRequestCtx().Response.Header.Reset()
		c.GetProxyResponse().Header.CopyTo(&c.GetOriginRequestCtx().Response.Header)
	}

	return f.BaseFilter.Post(c)
}
