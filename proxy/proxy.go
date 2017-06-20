package proxy

import (
	"container/list"
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/gateway/pkg/conf"
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/fagongzi/gateway/pkg/util"
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

	cnf             *conf.Conf
	filters         *list.List
	fastHTTPClients map[string]*util.FastHTTPClient
	routeTable      *model.RouteTable
}

// NewProxy create a new proxy
func NewProxy(config *conf.Conf) *Proxy {
	p := &Proxy{
		fastHTTPClients: make(map[string]*util.FastHTTPClient),
		cnf:             config,
		filters:         list.New(),
	}

	p.init()

	return p
}

func (p *Proxy) init() {
	err := p.initRouteTable()
	if err != nil {
		log.PanicError(err, "init etcd store error")
	}

	p.initFilters()
}

func (p *Proxy) initRouteTable() error {
	store, err := model.GetStoreFrom(p.cnf.RegistryAddr, p.cnf.Prefix)

	if err != nil {
		return err
	}

	register, _ := store.(model.Register)

	register.Registry(&model.ProxyInfo{
		Conf: p.cnf,
	})

	p.routeTable = model.NewRouteTable(p.cnf, store)
	p.routeTable.Load()

	return nil
}

func (p *Proxy) initFilters() {
	for _, filter := range p.cnf.Filers {
		f, err := newFilter(filter)
		if nil != err {
			log.Panicf("Proxy unknow filter <%+v>.", filter)
		}

		p.filters.PushBack(f)
	}
}

// Start start proxy
func (p *Proxy) Start() {
	err := p.startRPCServer()

	if nil != err {
		log.PanicErrorf(err, "Proxy start rpc at <%s> fail.", p.cnf.MgrAddr)
	}

	log.ErrorErrorf(fasthttp.ListenAndServe(p.cnf.Addr, p.ReverseProxyHandler), "Proxy exit at %s", p.cnf.Addr)
}

// ReverseProxyHandler http reverse handler
func (p *Proxy) ReverseProxyHandler(ctx *fasthttp.RequestCtx) {
	results := p.routeTable.Select(&ctx.Request)

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
			result.Merge = merge

			go func(result *model.RouteResult) {
				p.doProxy(ctx, wg, result)
			}(result)
		}

		wg.Wait()
	} else {
		p.doProxy(ctx, nil, results[0])
	}

	for _, result := range results {
		if result.Err != nil {
			if result.API.Mock != nil {
				result.API.RenderMock(ctx)
				result.Release()
				return
			}

			ctx.SetStatusCode(result.Code)
			result.Release()
			return
		}

		if !merge {
			p.writeResult(ctx, result.Res)
			result.Release()
			return
		}
	}

	for _, result := range results {
		for _, h := range MergeRemoveHeaders {
			result.Res.Header.Del(h)
		}
		result.Res.Header.CopyTo(&ctx.Response.Header)
	}

	ctx.Response.Header.SetContentType(MergeContentType)
	ctx.SetStatusCode(fasthttp.StatusOK)

	ctx.WriteString("{")

	for index, result := range results {
		ctx.WriteString("\"")
		ctx.WriteString(result.Node.AttrName)
		ctx.WriteString("\":")
		ctx.Write(result.Res.Body())
		if index < count-1 {
			ctx.WriteString(",")
		}

		result.Release()
	}

	ctx.WriteString("}")
}

func (p *Proxy) doProxy(ctx *fasthttp.RequestCtx, wg *sync.WaitGroup, result *model.RouteResult) {
	if nil != wg {
		defer wg.Done()
	}

	svr := result.Svr

	if nil == svr {
		result.Err = ErrNoServer
		result.Code = http.StatusServiceUnavailable
		return
	}

	outreq := copyRequest(&ctx.Request)

	// change url
	if result.NeedRewrite() {
		// if not use rewrite, it only change uri path and query string
		realPath := result.GetRewritePath(&ctx.Request)
		if "" != realPath {
			log.Infof("URL Rewrite from <%s> to <%s>", string(ctx.URI().FullURI()), realPath)
			outreq.SetRequestURI(realPath)
			outreq.SetHost(svr.Addr)
		} else {
			log.Warnf("URL Rewrite<%s> not matches <%s>", string(ctx.URI().FullURI()), result.Node.Rewrite)
			result.Err = ErrRewriteNotMatch
			result.Code = http.StatusBadRequest
			return
		}
	}

	c := newContext(p.routeTable, ctx, outreq, result)

	// pre filters
	filterName, code, err := p.doPreFilters(c)
	if nil != err {
		log.WarnErrorf(err, "Proxy Filter-Pre<%s> fail", filterName)
		result.Err = err
		result.Code = code
		return
	}

	c.SetStartAt(time.Now().UnixNano())
	res, err := p.getClient(svr.Addr).Do(outreq, svr.Addr)
	c.SetEndAt(time.Now().UnixNano())

	result.Res = res

	if err != nil || res.StatusCode() >= fasthttp.StatusInternalServerError {
		resCode := http.StatusServiceUnavailable

		if nil != err {
			log.InfoErrorf(err, "Proxy Fail <%s>", svr.Addr)
		} else {
			resCode = res.StatusCode()
			log.InfoErrorf(err, "Proxy Fail <%s>, Code <%d>", svr.Addr, res.StatusCode())
		}

		// 用户取消，不计算为错误
		if nil == err || !strings.HasPrefix(err.Error(), ErrPrefixRequestCancel) {
			p.doPostErrFilters(c)
		}

		result.Err = err
		result.Code = resCode
		return
	}

	log.Infof("Backend server[%s] responsed, code <%d>, body<%s>", svr.Addr, res.StatusCode(), res.Body())

	// post filters
	filterName, code, err = p.doPostFilters(c)
	if nil != err {
		log.InfoErrorf(err, "Proxy Filter-Post<%s> fail: %s ", filterName, err.Error())

		result.Err = err
		result.Code = code
		return
	}
}

func (p *Proxy) writeResult(ctx *fasthttp.RequestCtx, res *fasthttp.Response) {
	ctx.SetStatusCode(res.StatusCode())
	ctx.Write(res.Body())
}

func (p *Proxy) getClient(addr string) *util.FastHTTPClient {
	p.RLock()
	c, ok := p.fastHTTPClients[addr]
	if ok {
		p.RUnlock()
		return c
	}
	p.RUnlock()

	p.Lock()
	c, ok = p.fastHTTPClients[addr]
	if ok {
		p.Unlock()
		return c
	}

	p.fastHTTPClients[addr] = util.NewFastHTTPClient(p.cnf)
	p.Unlock()
	return c
}
