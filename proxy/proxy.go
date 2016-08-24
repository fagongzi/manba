package proxy

import (
	"container/list"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/gateway/conf"
	"github.com/fagongzi/gateway/pkg/model"
)

const (
	// ErrPrefixRequestCancel user cancel request error
	ErrPrefixRequestCancel = "request canceled"
)

var (
	// ErrNoServer no server
	ErrNoServer = errors.New("has no server")
)

var (
	// HeaderContentType content-type header
	HeaderContentType = "Content-Type"
	// MergeContentType merge operation using content-type
	MergeContentType = "application/json; charset=utf-8"
	// MergeRemoveHeaders merge operation need to remove headers
	MergeRemoveHeaders = []string{
		"Content-Length",
		"Content-Type",
		"Date",
	}
)

// Proxy Proxy
type Proxy struct {
	config        *conf.Conf
	routeTable    *model.RouteTable
	flushInterval time.Duration
	transport     http.RoundTripper
	filters       *list.List
}

// NewProxy create a new proxy
func NewProxy(config *conf.Conf, routeTable *model.RouteTable) *Proxy {
	p := &Proxy{
		config:     config,
		routeTable: routeTable,
		transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 30 * time.Second,
			}).Dial,
			ResponseHeaderTimeout: 10 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			DisableKeepAlives:     true,
		},

		filters: list.New(),
	}

	return p
}

// RegistryFilter registry a filter
func (p *Proxy) RegistryFilter(name string) {
	f, err := newFilter(name, p.config, p)
	if nil != err {
		log.Panicf("Proxy unknow filter <%s>.", name)
	}

	p.filters.PushBack(f)
}

// Start start proxy
func (p *Proxy) Start() {
	err := p.startRPCServer()

	if nil != err {
		log.PanicErrorf(err, "Proxy start rpc at <%s> fail.", p.config.MgrAddr)
	}

	log.ErrorErrorf(http.ListenAndServe(p.config.Addr, p), "Proxy exit at %s", p.config.Addr)
}

// ServeHTTP start http serve
func (p *Proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	results := p.routeTable.Select(req)

	if nil == results || len(results) == 0 {
		rw.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	count := len(results)
	merge := count > 1

	if merge {
		wg := &sync.WaitGroup{}
		wg.Add(count)

		for _, result := range results {
			result.Merge = merge

			go func(result *model.RouteResult) {
				p.doProxy(rw, req, wg, result)
			}(result)
		}

		wg.Wait()
	} else {
		p.doProxy(rw, req, nil, results[0])
	}

	for _, result := range results {
		if result.Err != nil {
			rw.WriteHeader(result.Code)
			return
		}

		if !merge {
			p.writeResult(rw, result.Res)
			return
		}
	}

	for _, result := range results {
		for _, h := range MergeRemoveHeaders {
			result.Res.Header.Del(h)
		}
		copyHeader(rw.Header(), result.Res.Header)
	}

	rw.Header().Add(HeaderContentType, MergeContentType)

	rw.WriteHeader(http.StatusOK)

	rw.Write([]byte("{"))

	for index, result := range results {
		rw.Write([]byte("\""))
		rw.Write([]byte(result.Node.AttrName))
		rw.Write([]byte("\":"))
		p.copyResponse(rw, result.Res.Body)
		result.Res.Body.Close() // close now, instead of defer, to populate res.Trailer
		if index < count-1 {
			rw.Write([]byte(","))
		}
	}

	rw.Write([]byte("}"))
}

func (p *Proxy) doProxy(rw http.ResponseWriter, req *http.Request, wg *sync.WaitGroup, result *model.RouteResult) {
	if nil != wg {
		defer wg.Done()
	}

	svr := result.Svr

	if nil == svr {
		result.Err = ErrNoServer
		result.Code = http.StatusServiceUnavailable
		return
	}

	transport := p.transport

	outreq, err := copyRequest(req)
	if err != nil {
		log.ErrorError(err)
	}

	//process client connect has gone before backend responsed
	if closeNotifier, ok := rw.(http.CloseNotifier); ok {
		if requestCanceler, ok := transport.(requestCanceler); ok {
			reqDone := make(chan struct{})
			defer close(reqDone)

			clientGone := closeNotifier.CloseNotify()

			outreq.Body = struct {
				io.Reader
				io.Closer
			}{
				Reader: &runOnFirstRead{
					Reader: outreq.Body,
					fn: func() {
						go func() {
							select {
							case <-clientGone:
								requestCanceler.CancelRequest(outreq)
							case <-reqDone:
							}
						}()
					},
				},
				Closer: outreq.Body,
			}
		}
	}

	path := req.URL.Path
	// change url
	if result.Node != nil {
		path = result.Node.URL
	}

	outreq.URL.Scheme = svr.Schema
	outreq.URL.Host = svr.Addr
	outreq.URL.Path = path

	outreq.RequestURI = outreq.URL.RequestURI()

	c := &filterContext{
		rw:         rw,
		req:        req,
		outreq:     outreq,
		result:     result,
		rb:         p.routeTable,
		runtimeVar: make(map[string]string),
	}

	// pre filters
	filterName, code, err := p.doPreFilters(c)
	if nil != err {
		log.WarnErrorf(err, "Proxy Filter-Pre<%s> fail", filterName)
		result.Err = err
		result.Code = code
		return
	}

	c.startAt = time.Now().UnixNano()
	res, err := transport.RoundTrip(outreq)
	c.endAt = time.Now().UnixNano()

	result.Res = res

	if err != nil || res.StatusCode >= http.StatusInternalServerError {
		resCode := http.StatusServiceUnavailable

		if nil != err {
			log.InfoErrorf(err, "Proxy Fail <%s>", svr.Addr)
		} else {
			resCode = res.StatusCode
			log.InfoErrorf(err, "Proxy Fail <%s>, Code <%d>", svr.Addr, res.StatusCode)
		}

		// 用户取消，不计算为错误
		if nil == err || !strings.HasPrefix(err.Error(), ErrPrefixRequestCancel) {
			p.doPostErrFilters(c)
		}

		result.Err = err
		result.Code = resCode
		return
	}

	// post filters
	filterName, code, err = p.doPostFilters(c)
	if nil != err {
		log.InfoErrorf(err, "Proxy Filter-Post<%s> fail: %s ", filterName, err.Error())

		result.Err = err
		result.Code = code
		return
	}
}

func (p *Proxy) writeResult(rw http.ResponseWriter, res *http.Response) {
	rw.WriteHeader(res.StatusCode)
	if len(res.Trailer) > 0 {
		// Force chunking if we saw a response trailer.
		// This prevents net/http from calculating the length for short
		// bodies and adding a Content-Length.
		if fl, ok := rw.(http.Flusher); ok {
			fl.Flush()
		}
	}

	p.copyResponse(rw, res.Body)
	res.Body.Close() // close now, instead of defer, to populate res.Trailer
	copyHeader(rw.Header(), res.Trailer)
}
