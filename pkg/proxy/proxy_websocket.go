package proxy

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

	"github.com/fagongzi/log"
	"github.com/fagongzi/util/hack"
	"github.com/gorilla/websocket"
	"github.com/koding/websocketproxy"
	"github.com/valyala/fasthttp"
)

const (
	websocketRspKey = "__ws_rsp"
)

// ServeHTTP  http reverse handler by http
func (p *Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if p.isStopped() {
		rw.WriteHeader(fasthttp.StatusServiceUnavailable)
		return
	}

	var buf bytes.Buffer
	buf.WriteByte(charLeft)
	buf.Write(hack.StringToSlice(req.Method))
	buf.WriteByte(charRight)
	buf.Write(hack.StringToSlice(req.RequestURI))
	requestTag := hack.SliceToString(buf.Bytes())

	if req.Method != "GET" {
		rw.WriteHeader(fasthttp.StatusMethodNotAllowed)
		return
	}

	ctx := &fasthttp.RequestCtx{}
	for k, vs := range req.Header {
		for _, v := range vs {
			ctx.Request.Header.Add(k, v)
		}
	}
	ctx.Request.SetRequestURI(req.RequestURI)

	api, dispatches := p.dispatcher.dispatch(&ctx.Request, requestTag)
	if len(dispatches) <= 0 &&
		(nil == api || api.meta.DefaultValue == nil) {
		rw.WriteHeader(fasthttp.StatusNotFound)
		p.dispatcher.dispatchCompleted()
		return
	}

	if len(dispatches) != 1 {
		log.Fatalf("websocket not support dispatch to multi backend server")
	}

	if !api.isWebSocket() {
		log.Fatalf("normal http request must use fasthttp")
	}

	dispatches[0].ctx = ctx
	p.doProxy(dispatches[0], func(c *proxyContext) {
		c.SetAttr(websocketRspKey, rw)
	})
	dispatches[0].release()
	p.dispatcher.dispatchCompleted()
}

func (p *Proxy) onWebsocket(c *proxyContext, addr string) (*fasthttp.Response, error) {
	resp := fasthttp.AcquireResponse()

	var r http.Request
	r.Method = "GET"
	r.Proto = "HTTP/1.1"
	r.ProtoMajor = 1
	r.ProtoMinor = 1
	r.RequestURI = string(c.forwardReq.RequestURI())
	r.Host = string(c.forwardReq.Host())

	hdr := make(http.Header)
	c.forwardReq.Header.VisitAll(func(k, v []byte) {
		sk := string(k)
		sv := string(v)
		switch sk {
		case "Transfer-Encoding":
			r.TransferEncoding = append(r.TransferEncoding, sv)
		default:
			hdr.Set(sk, sv)
		}
	})
	r.Header = hdr
	r.URL, _ = url.ParseRequestURI(r.RequestURI)

	wp := &websocketproxy.WebsocketProxy{
		Upgrader: &websocket.Upgrader{
			ReadBufferSize:  c.result.httpOption().ReadBufferSize,
			WriteBufferSize: c.result.httpOption().WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Director: func(incoming *http.Request, out http.Header) {
			out.Set("Origin", fmt.Sprintf("http://%s", addr))
			out.Set("Host", addr)
		},
		Backend: func(r *http.Request) *url.URL {
			u, _ := url.Parse(fmt.Sprintf("ws://%s%s", addr, r.RequestURI))
			return u
		},
	}

	wp.ServeHTTP(c.GetAttr(websocketRspKey).(http.ResponseWriter), &r)
	return resp, nil
}
