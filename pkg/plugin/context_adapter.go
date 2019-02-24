package plugin

import (
	"sync"
	"time"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/util/hack"
	"github.com/valyala/fasthttp"
)

var (
	contextPool sync.Pool
)

func acquireContext() *Ctx {
	v := contextPool.Get()
	if v == nil {
		return &Ctx{}
	}

	return v.(*Ctx)
}

func releaseContext(value *Ctx) {
	if value != nil {
		value.reset()
		contextPool.Put(value)
	}
}

// Ctx plugin ctx
type Ctx struct {
	delegate filter.Context
	origin   *FastHTTPRequestAdapter
	forward  *FastHTTPRequestAdapter
	response *FastHTTPResponseAdapter
}

func (c *Ctx) reset() {
	*c = Ctx{}
}

// OriginRequest return origin request
func (c *Ctx) OriginRequest() *FastHTTPRequestAdapter {
	if c.origin == nil {
		c.origin = newFastHTTPRequestAdapter(&c.delegate.OriginRequest().Request)
	}

	return c.origin
}

// ForwardRequest return forward request
func (c *Ctx) ForwardRequest() *FastHTTPRequestAdapter {
	if c.forward == nil {
		c.forward = newFastHTTPRequestAdapter(c.delegate.ForwardRequest())
	}

	return c.forward
}

// Response return response
func (c *Ctx) Response() *FastHTTPResponseAdapter {
	if c.response == nil {
		c.response = newFastHTTPResponseAdapter(c.delegate.Response())
	}

	return c.response
}

// SetAttr set attr to context
func (c *Ctx) SetAttr(key string, value interface{}) {
	c.delegate.SetAttr(key, value)
}

// HasAttr is attr in context
func (c *Ctx) HasAttr(key string) bool {
	return c.delegate.GetAttr(key) != nil
}

// GetAttr return attr value in the context
func (c *Ctx) GetAttr(key string) interface{} {
	return c.delegate.GetAttr(key)
}

// FastHTTPRequestAdapter fasthttp request adapter
type FastHTTPRequestAdapter struct {
	delegate *fasthttp.Request
}

func newFastHTTPRequestAdapter(req *fasthttp.Request) *FastHTTPRequestAdapter {
	return &FastHTTPRequestAdapter{
		delegate: req,
	}
}

// Header returns header value
func (req *FastHTTPRequestAdapter) Header(name string) string {
	return hack.SliceToString(req.delegate.Header.Peek(name))
}

// RemoveHeader remove header
func (req *FastHTTPRequestAdapter) RemoveHeader(name string) {
	req.delegate.Header.Del(name)
}

// Cookie returns cookie value
func (req *FastHTTPRequestAdapter) Cookie(name string) string {
	return hack.SliceToString(req.delegate.Header.Cookie(name))
}

// RemoveCookie remove cookie
func (req *FastHTTPRequestAdapter) RemoveCookie(name string) {
	req.delegate.Header.DelCookie(name)
}

// Query returns query string value
func (req *FastHTTPRequestAdapter) Query(name string) string {
	return hack.SliceToString(req.delegate.URI().QueryArgs().Peek(name))
}

// Body returns request body
func (req *FastHTTPRequestAdapter) Body() string {
	return hack.SliceToString(req.delegate.Body())
}

// SetBody set body
func (req *FastHTTPRequestAdapter) SetBody(body string) {
	req.delegate.SetBodyString(body)
}

// SetHeader set header value
func (req *FastHTTPRequestAdapter) SetHeader(name, value string) {
	req.delegate.Header.Add(name, value)
}

// SetCookie set cookie value
func (req *FastHTTPRequestAdapter) SetCookie(name, value string) {
	req.delegate.Header.SetCookie(name, value)
}

// FastHTTPResponseAdapter fasthttp response adapter
type FastHTTPResponseAdapter struct {
	delegate *fasthttp.Response
}

func newFastHTTPResponseAdapter(rsp *fasthttp.Response) *FastHTTPResponseAdapter {
	return &FastHTTPResponseAdapter{
		delegate: rsp,
	}
}

// Delegate returns delegate fasthttp response
func (rsp *FastHTTPResponseAdapter) Delegate() *fasthttp.Response {
	return rsp.delegate
}

// Header returns header value
func (rsp *FastHTTPResponseAdapter) Header(name string) string {
	return hack.SliceToString(rsp.delegate.Header.Peek(name))
}

// RemoveHeader remove header
func (rsp *FastHTTPResponseAdapter) RemoveHeader(name string) {
	rsp.delegate.Header.Del(name)
}

// Cookie returns cookie value
func (rsp *FastHTTPResponseAdapter) Cookie(name string) string {
	value := ""
	rsp.delegate.Header.VisitAllCookie(func(key, v []byte) {
		if hack.SliceToString(key) == name {
			value = hack.SliceToString(v)
		}
	})
	return value
}

// RemoveCookie remove cookie
func (rsp *FastHTTPResponseAdapter) RemoveCookie(name string) {
	rsp.delegate.Header.DelCookie(name)
}

// Body returns request body
func (rsp *FastHTTPResponseAdapter) Body() string {
	return hack.SliceToString(rsp.delegate.Body())
}

// SetBody set body
func (rsp *FastHTTPResponseAdapter) SetBody(body string) {
	rsp.delegate.SetBodyString(body)
}

// SetHeader set header value
func (rsp *FastHTTPResponseAdapter) SetHeader(name, value string) {
	rsp.delegate.Header.Add(name, value)
}

// SetCookie set cookie value
func (rsp *FastHTTPResponseAdapter) SetCookie(domain, path, name, value string, expire int64, httpOnly, secure bool) {
	ck := &fasthttp.Cookie{}
	ck.SetKey(name)
	ck.SetValue(value)
	ck.SetDomain(domain)
	ck.SetExpire(time.Now().Add(time.Second * time.Duration(expire)))
	ck.SetPath(path)
	ck.SetHTTPOnly(httpOnly)
	ck.SetSecure(secure)
	rsp.delegate.Header.SetCookie(ck)
}

// SetStatusCode set status code
func (rsp *FastHTTPResponseAdapter) SetStatusCode(code int) {
	rsp.delegate.SetStatusCode(code)
}
