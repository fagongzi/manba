package proxy

import (
	"sync"
	"time"

	"github.com/fagongzi/gateway/pkg/plugin"

	"github.com/fagongzi/gateway/pkg/expr"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/route"
	"github.com/fagongzi/gateway/pkg/store"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/fagongzi/goetty"
	"github.com/fagongzi/log"
	"github.com/fagongzi/util/hack"
	"github.com/fagongzi/util/task"
	"github.com/valyala/fasthttp"
)

type copyReq struct {
	origin     *fasthttp.Request
	api        *apiRuntime
	node       *apiNode
	to         *serverRuntime
	params     map[string][]byte
	idx        int
	requestTag string
}

func (req *copyReq) prepare() {
	if req.needRewrite() {
		// if not use rewrite, it only change uri path and query string
		realPath := req.rewriteURL()
		if "" != realPath {
			req.origin.SetRequestURI(realPath)
			req.origin.SetHost(req.to.meta.Addr)

			log.Infof("%s: dispatch node %d rewrite url to %s for copy",
				req.requestTag,
				req.idx,
				realPath)
		}
	}
}

func (req *copyReq) needRewrite() bool {
	return req.node.meta.URLRewrite != ""
}

func (req *copyReq) rewriteURL() string {
	ctx := &expr.Ctx{}
	ctx.Origin = req.origin
	ctx.Params = req.params
	return hack.SliceToString(expr.Exec(ctx, req.node.parsedExprs...))
}

type dispatchNode struct {
	rd       *render
	ctx      *fasthttp.RequestCtx
	multiCtx *multiContext
	exprCtx  *expr.Ctx
	wg       *sync.WaitGroup

	requestTag           string
	idx                  int
	api                  *apiRuntime
	node                 *apiNode
	dest                 *serverRuntime
	copyTo               *serverRuntime
	res                  *fasthttp.Response
	cachedBody, cachedCT []byte
	err                  error
	code                 int
}

func (dn *dispatchNode) setHost(forwardReq *fasthttp.Request) {
	switch dn.node.meta.HostType {
	case metapb.HostOrigin:
		forwardReq.SetHostBytes(dn.ctx.Request.Host())
	case metapb.HostServerAddress:
		forwardReq.SetHost(dn.dest.meta.Addr)
	case metapb.HostCustom:
		forwardReq.SetHost(dn.node.meta.CustemHost)
	}
}

func (dn *dispatchNode) reset() {
	*dn = emptyDispathNode
}

func (dn *dispatchNode) hasRetryStrategy() bool {
	return dn.retryStrategy() != nil
}

func (dn *dispatchNode) matchRetryStrategy(target int32) bool {
	for _, code := range dn.retryStrategy().Codes {
		if code == target {
			return true
		}
	}

	return false
}

func (dn *dispatchNode) matchAllRetryStrategy() bool {
	return len(dn.retryStrategy().Codes) == 0
}

func (dn *dispatchNode) httpOption() *util.HTTPOption {
	return &dn.node.httpOption
}

func (dn *dispatchNode) retryStrategy() *metapb.RetryStrategy {
	return dn.node.meta.RetryStrategy
}

func (dn *dispatchNode) hasError() bool {
	return dn.err != nil ||
		dn.code >= fasthttp.StatusBadRequest
}

func (dn *dispatchNode) hasDefaultValue() bool {
	return dn.node.meta.DefaultValue != nil
}

func (dn *dispatchNode) release() {
	if nil != dn.res {
		fasthttp.ReleaseResponse(dn.res)
	}
}

func (dn *dispatchNode) needRewrite() bool {
	return dn.node.meta.URLRewrite != ""
}

func (dn *dispatchNode) getResponseContentType() []byte {
	if len(dn.cachedCT) > 0 {
		return dn.cachedCT
	}

	if nil != dn.res {
		return dn.res.Header.ContentType()
	}

	return nil
}

func (dn *dispatchNode) getResponseBody() []byte {
	if len(dn.cachedBody) > 0 {
		return dn.cachedBody
	}

	if dn.node.meta.UseDefault ||
		(dn.hasError() && dn.hasDefaultValue()) {
		return dn.node.meta.DefaultValue.Body
	}

	if nil != dn.res {
		return dn.res.Body()
	}

	return nil
}

func (dn *dispatchNode) copyHeaderTo(ctx *fasthttp.RequestCtx) {
	if dn.node.meta.UseDefault ||
		(dn.hasError() && dn.hasDefaultValue()) {
		for _, hd := range dn.node.meta.DefaultValue.Headers {
			(&ctx.Response.Header).Add(hd.Name, hd.Value)
		}

		for _, ck := range dn.node.defaultCookies {
			(&ctx.Response.Header).SetCookie(ck)
		}
		return
	}

	if dn.res != nil {
		for _, h := range MultiResultsRemoveHeaders {
			dn.res.Header.Del(h)
		}
		dn.res.Header.CopyTo(&ctx.Response.Header)
	}
}

func (dn *dispatchNode) maybeDone() {
	if nil != dn.multiCtx {
		dn.multiCtx.completePart(dn.node.meta.AttrName, dn.getResponseBody())
		if nil != dn.wg {
			dn.wg.Done()
		}
	}
}

type dispatcher struct {
	cnf            *Cfg
	routings       map[uint64]*routingRuntime
	route          *route.Route
	apis           map[uint64]*apiRuntime
	clusters       map[uint64]*clusterRuntime
	servers        map[uint64]*serverRuntime
	binds          map[uint64]*binds
	proxies        map[string]*metapb.Proxy
	plugins        map[uint64]*metapb.Plugin
	appliedPlugins *metapb.AppliedPlugins
	jsEngineFunc   func(*plugin.Engine)
	checkerC       chan uint64
	watchStopC     chan bool
	watchEventC    chan *store.Evt
	analysiser     *util.Analysis
	store          store.Store
	httpClient     *util.FastHTTPClient
	tw             *goetty.TimeoutWheel
	runner         *task.Runner
}

func newDispatcher(cnf *Cfg, db store.Store, runner *task.Runner, jsEngineFunc func(*plugin.Engine)) *dispatcher {
	tw := goetty.NewTimeoutWheel(goetty.WithTickInterval(time.Second))
	rt := &dispatcher{
		cnf:          cnf,
		tw:           tw,
		store:        db,
		runner:       runner,
		analysiser:   util.NewAnalysis(tw),
		httpClient:   util.NewFastHTTPClient(),
		clusters:     make(map[uint64]*clusterRuntime),
		servers:      make(map[uint64]*serverRuntime),
		route:        route.NewRoute(),
		apis:         make(map[uint64]*apiRuntime),
		routings:     make(map[uint64]*routingRuntime),
		binds:        make(map[uint64]*binds),
		proxies:      make(map[string]*metapb.Proxy),
		plugins:      make(map[uint64]*metapb.Plugin),
		jsEngineFunc: jsEngineFunc,
		checkerC:     make(chan uint64, 1024),
		watchStopC:   make(chan bool),
		watchEventC:  make(chan *store.Evt),
	}

	rt.readyToHeathChecker()
	return rt
}

func (r *dispatcher) dispatch(reqCtx *fasthttp.RequestCtx, requestTag string) (*apiRuntime, []*dispatchNode, *expr.Ctx) {
	req := &reqCtx.Request
	dispatcherRoute := r.route
	var targetAPI *apiRuntime
	var dispatches []*dispatchNode

	exprCtx := acquireExprCtx()
	exprCtx.Origin = req

	id, ok := dispatcherRoute.Find(req.URI().Path(), hack.SliceToString(req.Header.Method()), exprCtx.AddParam)
	if ok {
		if api, ok := r.apis[id]; ok && api.matches(req) {
			targetAPI = api
		}
	}

	if targetAPI == nil {
		return targetAPI, dispatches, exprCtx
	}

	if targetAPI.meta.UseDefault {
		log.Debugf("%s: match api %s, and use default force",
			requestTag,
			targetAPI.meta.Name)
	} else {
		for idx, node := range targetAPI.nodes {
			dn := acquireDispathNode()
			dn.idx = idx
			dn.api = targetAPI
			dn.node = node
			dn.exprCtx = exprCtx
			r.selectServer(reqCtx, dn, requestTag)
			dispatches = append(dispatches, dn)
		}
	}

	return targetAPI, dispatches, exprCtx
}

func (r *dispatcher) selectServer(reqCtx *fasthttp.RequestCtx, dn *dispatchNode, requestTag string) {
	dn.dest = r.selectServerFromCluster(reqCtx, dn.node.meta.ClusterID)
	r.adjustByRouting(dn.api.meta.ID, reqCtx, dn, requestTag)
}

func (r *dispatcher) adjustByRouting(apiID uint64, reqCtx *fasthttp.RequestCtx, dn *dispatchNode, requestTag string) {
	routings := r.routings

	for _, routing := range routings {
		if routing.isUp() && routing.matches(apiID, &reqCtx.Request, requestTag) {
			log.Infof("%s: match routing %s, %s traffic to cluster %d",
				requestTag,
				routing.meta.Name,
				routing.meta.Status.String(),
				routing.meta.ClusterID)

			svr := r.selectServerFromCluster(reqCtx, routing.meta.ClusterID)

			switch routing.meta.Strategy {
			case metapb.Split:
				dn.dest = svr
			case metapb.Copy:
				dn.copyTo = svr
			}
			break
		}
	}
}

func (r *dispatcher) selectServerFromCluster(ctx *fasthttp.RequestCtx, id uint64) *serverRuntime {
	cluster, ok := r.clusters[id]
	if !ok {
		return nil
	}

	if bindsInfo, ok := r.binds[id]; ok {
		return r.servers[cluster.selectServer(ctx, bindsInfo.actives)]
	}

	return nil
}
