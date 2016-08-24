package proxy

import (
	"github.com/fagongzi/gateway/conf"
)

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
	baseFilter
	config *conf.Conf
	proxy  *Proxy
}

func newHeadersFilter(config *conf.Conf, proxy *Proxy) Filter {
	return HeadersFilter{
		config: config,
		proxy:  proxy,
	}
}

// Name return name of this filter
func (f HeadersFilter) Name() string {
	return FilterHeader
}

// Pre execute before proxy
func (f HeadersFilter) Pre(c *filterContext) (statusCode int, err error) {
	c.outreq.Proto = "HTTP/1.1"
	c.outreq.ProtoMajor = 1
	c.outreq.ProtoMinor = 1
	c.outreq.Close = false

	copyHeader(c.outreq.Header, c.req.Header)

	for _, h := range hopHeaders {
		c.outreq.Header.Del(h)
	}

	c.outreq.Host = c.req.Host

	return f.baseFilter.Pre(c)
}

// Post execute after proxy
func (f HeadersFilter) Post(c *filterContext) (statusCode int, err error) {
	for _, h := range hopHeaders {
		c.result.Res.Header.Del(h)
	}

	// 需要合并处理的，不做header的复制，由proxy做合并
	if !c.result.Merge {
		copyHeader(c.rw.Header(), c.result.Res.Header)
	}

	return f.baseFilter.Post(c)
}
