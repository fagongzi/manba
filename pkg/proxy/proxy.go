package proxy

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
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
	// MultiResultsContentType merge operation using content-type
	MultiResultsContentType = "application/json; charset=utf-8"
	// MultiResultsRemoveHeaders merge operation need to remove headers
	MultiResultsRemoveHeaders = []string{
		"Content-Length",
		"Content-Type",
		"Date",
	}
)

// Proxy Proxy
type Proxy struct {
	sync.RWMutex

	dispatchIndex, copyIndex uint64
	dispatches               []chan *dispathNode
	copies                   []chan *copyReq

	cfg        *Cfg
	filtersMap map[string]filter.Filter
	filters    []filter.Filter
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
func NewProxy(cfg *Cfg) *Proxy {
	p := &Proxy{
		client: util.NewFastHTTPClientOption(&util.HTTPOption{
			MaxConnDuration:     cfg.Option.LimitDurationConnKeepalive,
			MaxIdleConnDuration: cfg.Option.LimitDurationConnIdle,
			ReadTimeout:         cfg.Option.LimitTimeoutRead,
			WriteTimeout:        cfg.Option.LimitTimeoutWrite,
			MaxResponseBodySize: cfg.Option.LimitBytesBody,
			WriteBufferSize:     cfg.Option.LimitBufferWrite,
			ReadBufferSize:      cfg.Option.LimitBufferRead,
			MaxConns:            cfg.Option.LimitCountConn,
		}),
		cfg:           cfg,
		filtersMap:    make(map[string]filter.Filter),
		stopC:         make(chan struct{}),
		runner:        task.NewRunner(),
		copies:        make([]chan *copyReq, cfg.Option.LimitCountCopyWorker, cfg.Option.LimitCountCopyWorker),
		dispatches:    make([]chan *dispathNode, cfg.Option.LimitCountDispatchWorker, cfg.Option.LimitCountDispatchWorker),
		dispatchIndex: 0,
		copyIndex:     0,
	}

	p.init()

	return p
}

// Start start proxy
func (p *Proxy) Start() {
	go p.listenToStop()

	p.readyToCopy()
	p.readyToDispatch()

	log.Infof("gateway proxy started at <%s>", p.cfg.Addr)
	err := fasthttp.ListenAndServe(p.cfg.Addr, p.ReverseProxyHandler)
	if err != nil {
		log.Fatalf("gateway proxy start failed, errors:\n%+v",
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
	err := p.initDispatcher()
	if err != nil {
		log.Fatalf("init route table failed, errors:\n%+v",
			err)
	}

	p.initFilters()

	err = p.dispatcher.store.RegistryProxy(&metapb.Proxy{
		Addr:    p.cfg.Addr,
		AddrRPC: p.cfg.AddrRPC,
	}, p.cfg.TTLProxy)
	if err != nil {
		log.Fatalf("init route table failed, errors:\n%+v",
			err)
	}

	p.dispatcher.load()
}

func (p *Proxy) initDispatcher() error {
	s, err := store.GetStoreFrom(p.cfg.AddrStore, p.cfg.Namespace)

	if err != nil {
		return err
	}

	p.dispatcher = newDispatcher(p.cfg, s, p.runner)
	return nil
}

func (p *Proxy) initFilters() {
	for _, filter := range p.cfg.Filers {
		f, err := p.newFilter(filter)
		if nil != err {
			log.Fatalf("init filter failed, filter=<%+v> errors:\n%+v",
				filter,
				err)
		}

		err = f.Init(filter.ExternalCfg)
		if nil != err {
			log.Fatalf("init filter failed, filter=<%+v> errors:\n%+v",
				filter,
				err)
		}

		log.Infof("filter added, filter=<%+v>", filter)
		p.filters = append(p.filters, f)
		p.filtersMap[f.Name()] = f
	}
}

func (p *Proxy) readyToDispatch() {
	for i := uint64(0); i < p.cfg.Option.LimitCountDispatchWorker; i++ {
		c := make(chan *dispathNode, 1024)
		p.dispatches[i] = c

		_, err := p.runner.RunCancelableTask(func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				case dn := <-c:
					if dn != nil {
						p.doProxy(dn)
					}
				}
			}
		})
		if err != nil {
			log.Fatalf("init dispatch workers failed, errors:\n%+v", err)
		}
	}
}

func (p *Proxy) readyToCopy() {
	for i := uint64(0); i < p.cfg.Option.LimitCountCopyWorker; i++ {
		c := make(chan *copyReq, 1024)
		p.copies[i] = c

		_, err := p.runner.RunCancelableTask(func(ctx context.Context) {
			for {
				select {
				case <-ctx.Done():
					return
				case req := <-c:
					if req != nil {
						p.doCopy(req)
					}
				}
			}
		})
		if err != nil {
			log.Fatalf("init copy workers failed, errors:\n%+v", err)
		}
	}
}

// ReverseProxyHandler http reverse handler
func (p *Proxy) ReverseProxyHandler(ctx *fasthttp.RequestCtx) {
	if p.isStopped() {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		return
	}

	dispatches, template := p.dispatcher.dispatch(&ctx.Request)
	if len(dispatches) == 0 {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		return
	}

	log.Infof("api(%s): %s %s",
		dispatches[0].api.meta.Name,
		ctx.Method(),
		ctx.RequestURI())

	rd := newRender(dispatches, template)

	for _, dn := range dispatches {
		dn.wg = rd.wg
		dn.ctx = ctx

		if dn.copyTo != nil {
			p.copies[getIndex(&p.copyIndex, p.cfg.Option.LimitCountCopyWorker)] <- &copyReq{
				origin: copyRequest(&ctx.Request),
				to:     dn.copyTo,
				api:    dn.api,
				node:   dn.node,
			}
		}

		if rd.multi {
			p.dispatches[getIndex(&p.dispatchIndex, p.cfg.Option.LimitCountDispatchWorker)] <- dn
		} else {
			p.doProxy(dn)
		}
	}

	rd.render(ctx)
}

func (p *Proxy) doCopy(req *copyReq) {
	svr := req.to

	if nil == svr {
		return
	}

	req.prepare()

	log.Debugf("copy: copy to %s", req.to.meta.Addr)

	res, err := p.client.Do(req.origin, svr.meta.Addr, nil)
	if err != nil {
		log.Errorf("copy: copy to %s failed, errors:\n%+v", req.to.meta.Addr, err)
		return
	}

	if res != nil {
		fasthttp.ReleaseResponse(res)
	}

	fasthttp.ReleaseRequest(req.origin)
}

func (p *Proxy) doProxy(dn *dispathNode) {
	ctx := dn.ctx
	svr := dn.dest

	if nil == svr {
		dn.err = ErrNoServer
		dn.code = http.StatusServiceUnavailable
		dn.maybeDone()
		return
	}

	forwardReq := copyRequest(&ctx.Request)
	forwardReq.SetHost(svr.meta.Addr)

	// change url
	if dn.needRewrite() {
		// if not use rewrite, it only change uri path and query string
		realPath := dn.rewiteURL(&ctx.Request)
		if "" != realPath {
			if log.DebugEnabled() {
				log.Debugf("dispatch: rewrite, from=<%s> to=<%s>",
					string(ctx.URI().FullURI()),
					realPath)
			}

			forwardReq.SetRequestURI(realPath)
		} else {
			log.Warnf("dispatch: rewrite not matches, origin=<%s> pattern=<%s>",
				string(ctx.URI().FullURI()),
				dn.node.meta.URLRewrite)

			dn.err = ErrRewriteNotMatch
			dn.code = http.StatusBadRequest
			dn.maybeDone()
			return
		}
	}

	c := newContext(p.dispatcher, ctx, forwardReq, dn)

	// pre filters
	filterName, code, err := p.doPreFilters(c)
	if nil != err {
		log.Warnf("dispatch: call pre filter failed, filter=<%s> errors:\n%+v",
			filterName,
			err)

		dn.err = err
		dn.code = code
		dn.maybeDone()
		return
	}

	// hit cache
	if value := c.GetAttr(cacheHit); nil != value {
		log.Debugf("dispatch: hit cahce for %s", string(forwardReq.RequestURI()))
		dn.cachedCT, dn.cachedBody = parseCachedValue(value.([]byte))
		dn.maybeDone()
		return
	}

	res, err := p.client.Do(forwardReq, svr.meta.Addr, nil)
	c.(*proxyContext).setEndAt(time.Now())

	dn.res = res

	if err != nil || res.StatusCode() >= fasthttp.StatusInternalServerError {
		resCode := http.StatusServiceUnavailable

		if nil != err {
			log.Warnf("dispatch: failed, target=<%s> errors:\n%+v",
				svr.meta.Addr,
				err)
		} else {
			resCode = res.StatusCode()
			log.Warnf("dispatch: returns error code, target=<%s> code=<%d>",
				svr.meta.Addr,
				res.StatusCode())
		}

		if nil == err || !strings.HasPrefix(err.Error(), ErrPrefixRequestCancel) {
			p.doPostErrFilters(c)
		}

		dn.err = err
		dn.code = resCode
		dn.maybeDone()
		return
	}

	if log.DebugEnabled() {
		log.Debugf("dispatch: return, target=<%s> code=<%d> body=<%s>",
			svr.meta.Addr,
			res.StatusCode(),
			string(res.Body()))
	}

	// post filters
	filterName, code, err = p.doPostFilters(c)
	if nil != err {
		log.Warnf("dispatch: call post filter failed, filter=<%s> errors:\n%+v",
			filterName,
			err)

		dn.err = err
		dn.code = code
		dn.maybeDone()
		return
	}

	dn.maybeDone()
}

func getIndex(opt *uint64, size uint64) int {
	return int(atomic.AddUint64(opt, 1) % size)
}
