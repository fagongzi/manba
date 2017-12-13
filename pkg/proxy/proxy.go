package proxy

import (
	"container/list"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fagongzi/gateway/pkg/store"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/fagongzi/log"
	"github.com/fagongzi/util/task"
	"github.com/valyala/fasthttp"
)

var (
	// ErrPrefixRequestCancel user cancel request error
	ErrPrefixRequestCancel = "request canceled"
	// ErrNoServer no server
	ErrNoServer = errors.New("has no server")
	// ErrRewriteNotMatch rewrite not match request url
	ErrRewriteNotMatch = errors.New("rewrite not match request url")
)

var (
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
	sync.RWMutex

	cnf        *Cfg
	filters    *list.List
	client     *util.FastHTTPClient
	dispatcher *dispatcher

	rpcListener net.Listener

	runner   *task.Runner
	stopped  int32
	stopC    chan struct{}
	stopOnce sync.Once
	stopWG   sync.WaitGroup
}

// NewProxy create a new proxy
func NewProxy(cnf *Cfg) *Proxy {
	p := &Proxy{
		client: util.NewFastHTTPClientOption(&util.HTTPOption{
			MaxConnDuration:     cnf.Option.LimitDurationConnKeepalive,
			MaxIdleConnDuration: cnf.Option.LimitDurationConnIdle,
			ReadTimeout:         cnf.Option.LimitTimeoutRead,
			WriteTimeout:        cnf.Option.LimitTimeoutWrite,
			MaxResponseBodySize: cnf.Option.LimitBytesBody,
			WriteBufferSize:     cnf.Option.LimitBufferWrite,
			ReadBufferSize:      cnf.Option.LimitBufferRead,
			MaxConns:            cnf.Option.LimitCountConn,
		}),
		cnf:     cnf,
		filters: list.New(),
		stopC:   make(chan struct{}),
		runner:  task.NewRunner(),
	}

	p.init()

	return p
}

// Start start proxy
func (p *Proxy) Start() {
	go p.listenToStop()

	log.Infof("gateway proxy started at <%s>", p.cnf.Addr)
	err := fasthttp.ListenAndServe(p.cnf.Addr, p.ReverseProxyHandler)
	if err != nil {
		log.Errorf("gateway proxy start failed, errors:\n%+v",
			err)
		return
	}
}

// Stop stop the proxy
func (p *Proxy) Stop() {
	log.Infof("stop: start to stop gateway proxy")

	p.stopWG.Add(1)
	p.stopC <- struct{}{}
	p.stopWG.Wait()

	log.Infof("stop: gateway proxy stopped")
}

func (p *Proxy) listenToStop() {
	<-p.stopC
	p.doStop()
}

func (p *Proxy) doStop() {
	p.stopOnce.Do(func() {
		defer p.stopWG.Done()
		p.setStopped()
		p.runner.Stop()
	})
}

func (p *Proxy) stopRPC() error {
	return p.rpcListener.Close()
}

func (p *Proxy) setStopped() {
	atomic.StoreInt32(&p.stopped, 1)
}

func (p *Proxy) isStopped() bool {
	return atomic.LoadInt32(&p.stopped) == 1
}

func (p *Proxy) init() {
	p.initFilters()

	err := p.initDispatcher()
	if err != nil {
		log.Fatalf("init route table failed, errors:\n%+v",
			err)
	}
}

func (p *Proxy) initDispatcher() error {
	s, err := store.GetStoreFrom(p.cnf.AddrStore, p.cnf.Namespace)

	if err != nil {
		return err
	}

	p.dispatcher = newDispatcher(p.cnf, s, p.runner)
	p.dispatcher.load()

	return nil
}

func (p *Proxy) initFilters() {
	for _, filter := range p.cnf.Filers {
		f, err := newFilter(filter)
		if nil != err {
			log.Fatalf("init filter failed, filter=<%+v> errors:\n%+v",
				filter,
				err)
		}

		log.Infof("filter added, filter=<%+v>", filter)
		p.filters.PushBack(f)
	}
}

// ReverseProxyHandler http reverse handler
func (p *Proxy) ReverseProxyHandler(ctx *fasthttp.RequestCtx) {
	if p.isStopped() {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		return
	}

	results := p.dispatcher.dispatch(&ctx.Request)

	if nil == results || len(results) == 0 {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}

	count := len(results)
	merge := count > 1

	if merge {
		wg := &sync.WaitGroup{}
		wg.Add(count)

		for _, result := range results {
			go func(result *dispathNode) {
				p.doProxy(ctx, wg, result)
			}(result)
		}

		wg.Wait()
	} else {
		p.doProxy(ctx, nil, results[0])
	}

	for _, result := range results {
		if result.err != nil {
			if result.api.meta.DefaultValue != nil {
				result.api.renderDefault(ctx)
				result.release()
				return
			}

			ctx.SetStatusCode(result.code)
			result.release()
			return
		}

		if !merge {
			p.writeResult(ctx, result.res)
			result.release()
			return
		}
	}

	for _, result := range results {
		for _, h := range MergeRemoveHeaders {
			result.res.Header.Del(h)
		}
		result.res.Header.CopyTo(&ctx.Response.Header)
	}

	ctx.Response.Header.SetContentType(MergeContentType)
	ctx.SetStatusCode(fasthttp.StatusOK)

	ctx.WriteString("{")

	for index, result := range results {
		ctx.WriteString("\"")
		ctx.WriteString(result.node.meta.AttrName)
		ctx.WriteString("\":")
		ctx.Write(result.res.Body())
		if index < count-1 {
			ctx.WriteString(",")
		}

		result.release()
	}

	ctx.WriteString("}")
}

func (p *Proxy) doProxy(ctx *fasthttp.RequestCtx, wg *sync.WaitGroup, result *dispathNode) {
	if nil != wg {
		defer wg.Done()
	}

	svr := result.dest

	if nil == svr {
		result.err = ErrNoServer
		result.code = http.StatusServiceUnavailable
		return
	}

	forwardReq := copyRequest(&ctx.Request)

	// change url
	if result.needRewrite() {
		// if not use rewrite, it only change uri path and query string
		realPath := result.rewiteURL(&ctx.Request)
		if "" != realPath {
			if log.DebugEnabled() {
				log.Debugf("proxy: rewrite, from=<%s> to=<%s>",
					string(ctx.URI().FullURI()),
					realPath)
			}

			forwardReq.SetRequestURI(realPath)
			forwardReq.SetHost(svr.meta.Addr)
		} else {
			log.Warnf("proxy: rewrite not matches, origin=<%s> pattern=<%s>",
				string(ctx.URI().FullURI()),
				result.node.meta.URLRewrite)

			result.err = ErrRewriteNotMatch
			result.code = http.StatusBadRequest
			return
		}
	}

	c := newContext(p.dispatcher, ctx, forwardReq, result)

	// pre filters
	filterName, code, err := p.doPreFilters(c)
	if nil != err {
		log.Warnf("proxy: call pre filter failed, filter=<%s> errors:\n%+v",
			filterName,
			err)

		result.err = err
		result.code = code
		return
	}

	res, err := p.client.Do(forwardReq, svr.meta.Addr, nil)
	c.SetEndAt(time.Now())

	result.res = res

	if err != nil || res.StatusCode() >= fasthttp.StatusInternalServerError {
		resCode := http.StatusServiceUnavailable

		if nil != err {
			log.Warnf("proxy: failed, target=<%s> errors:\n%+v",
				svr.meta.Addr,
				err)
		} else {
			resCode = res.StatusCode()
			log.Warnf("proxy: returns error code, target=<%s> code=<%d>",
				svr.meta.Addr,
				res.StatusCode())
		}

		if nil == err || !strings.HasPrefix(err.Error(), ErrPrefixRequestCancel) {
			p.doPostErrFilters(c)
		}

		result.err = err
		result.code = resCode
		return
	}

	if log.DebugEnabled() {
		log.Debugf("proxy: return, target=<%s> code=<%d> body=<%d>",
			svr.meta.Addr,
			res.StatusCode(),
			res.Body())
	}

	// post filters
	filterName, code, err = p.doPostFilters(c)
	if nil != err {
		log.Warnf("proxy: call post filter failed, filter=<%s> errors:\n%+v",
			filterName,
			err)

		result.err = err
		result.code = code
		return
	}
}

func (p *Proxy) writeResult(ctx *fasthttp.RequestCtx, res *fasthttp.Response) {
	ctx.SetStatusCode(res.StatusCode())
	ctx.Write(res.Body())
}
