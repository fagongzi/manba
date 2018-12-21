package proxy

import (
	"sync"
	"time"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/store"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/fagongzi/goetty"
	"github.com/fagongzi/log"
	"github.com/fagongzi/util/task"
	"github.com/valyala/fasthttp"
)

type copyReq struct {
	origin     *fasthttp.Request
	api        *apiRuntime
	node       *apiNode
	to         *serverRuntime
	idx        int
	requestTag string
}

func (req *copyReq) prepare() {
	if req.needRewrite() {
		// if not use rewrite, it only change uri path and query string
		realPath := req.rewiteURL()
		if "" != realPath {
			req.origin.SetRequestURI(realPath)
			req.origin.SetHost(req.to.meta.Addr)

			log.Infof("%s: dipatch node %d rewrite url to %s for copy",
				req.requestTag,
				req.idx,
				realPath)
		}
	}
}

func (req *copyReq) needRewrite() bool {
	return req.node.meta.URLRewrite != ""
}

func (req *copyReq) rewiteURL() string {
	return req.api.rewriteURL(req.origin, req.node, nil)
}

type dispathNode struct {
	rd       *render
	ctx      *fasthttp.RequestCtx
	multiCtx *multiContext
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

func (dn *dispathNode) reset() {
	*dn = emptyDispathNode
}

func (dn *dispathNode) hasRetryStrategy() bool {
	return dn.retryStrategy() != nil
}

func (dn *dispathNode) matchRetryStrategy(target int32) bool {
	for _, code := range dn.retryStrategy().Codes {
		if code == target {
			return true
		}
	}

	return false
}

func (dn *dispathNode) matchAllRetryStrategy() bool {
	return len(dn.retryStrategy().Codes) == 0
}

func (dn *dispathNode) httpOption() *util.HTTPOption {
	return &dn.node.httpOption
}

func (dn *dispathNode) retryStrategy() *metapb.RetryStrategy {
	return dn.node.meta.RetryStrategy
}

func (dn *dispathNode) hasError() bool {
	return dn.err != nil ||
		dn.code >= fasthttp.StatusBadRequest
}

func (dn *dispathNode) hasDefaultValue() bool {
	return dn.node.meta.DefaultValue != nil
}

func (dn *dispathNode) release() {
	if nil != dn.res {
		fasthttp.ReleaseResponse(dn.res)
	}
}

func (dn *dispathNode) needRewrite() bool {
	return dn.node.meta.URLRewrite != ""
}

func (dn *dispathNode) rewiteURL(req *fasthttp.Request) string {
	return dn.api.rewriteURL(req, dn.node, dn.multiCtx)
}

func (dn *dispathNode) getResponseContentType() []byte {
	if len(dn.cachedCT) > 0 {
		return dn.cachedCT
	}

	if nil != dn.res {
		return dn.res.Header.ContentType()
	}

	return nil
}

func (dn *dispathNode) getResponseBody() []byte {
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

func (dn *dispathNode) copyHeaderTo(ctx *fasthttp.RequestCtx) {
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

func (dn *dispathNode) maybeDone() {
	if nil != dn.wg {
		dn.multiCtx.completePart(dn.node.meta.AttrName, dn.getResponseBody())
		dn.wg.Done()
	}
}

type dispatcher struct {
	sync.RWMutex

	cnf           *Cfg
	routings      map[uint64]*routingRuntime
	apis          map[uint64]*apiRuntime
	apiSortedKeys []uint64
	clusters      map[uint64]*clusterRuntime
	servers       map[uint64]*serverRuntime
	binds         map[uint64]map[uint64]*clusterRuntime
	proxies       map[string]*metapb.Proxy
	checkerC      chan uint64
	watchStopC    chan bool
	watchEventC   chan *store.Evt
	analysiser    *util.Analysis
	store         store.Store
	httpClient    *util.FastHTTPClient
	tw            *goetty.TimeoutWheel
	runner        *task.Runner
}

func newDispatcher(cnf *Cfg, db store.Store, runner *task.Runner) *dispatcher {
	tw := goetty.NewTimeoutWheel(goetty.WithTickInterval(time.Second))
	rt := &dispatcher{
		cnf:           cnf,
		tw:            tw,
		store:         db,
		runner:        runner,
		analysiser:    util.NewAnalysis(tw),
		httpClient:    util.NewFastHTTPClient(),
		clusters:      make(map[uint64]*clusterRuntime),
		servers:       make(map[uint64]*serverRuntime),
		apis:          make(map[uint64]*apiRuntime),
		apiSortedKeys: make([]uint64, 0),
		routings:      make(map[uint64]*routingRuntime),
		binds:         make(map[uint64]map[uint64]*clusterRuntime),
		proxies:       make(map[string]*metapb.Proxy),
		checkerC:      make(chan uint64, 1024),
		watchStopC:    make(chan bool),
		watchEventC:   make(chan *store.Evt),
	}

	rt.readyToHeathChecker()
	return rt
}

func (r *dispatcher) dispatchCompleted() {
	r.RUnlock()
}

func (r *dispatcher) dispatch(req *fasthttp.Request, requestTag string) (*apiRuntime, []*dispathNode) {
	r.RLock()

	var targetAPI *apiRuntime
	var dispathes []*dispathNode
	for _, apiKey := range r.apiSortedKeys {
		api := r.apis[apiKey]
		if api.matches(req) {
			targetAPI = api
			if api.meta.UseDefault {
				log.Debugf("%s: match api %s, and use default force",
					requestTag,
					api.meta.Name)
				break
			}

			for idx, node := range api.nodes {
				dn := acquireDispathNode()
				dn.idx = idx
				dn.api = api
				dn.node = node
				r.selectServer(req, dn, requestTag)
				dispathes = append(dispathes, dn)
			}
			break
		}
	}

	return targetAPI, dispathes
}

func (r *dispatcher) selectServer(req *fasthttp.Request, dn *dispathNode, requestTag string) {
	dn.dest = r.selectServerFromCluster(req, dn.node.meta.ClusterID)
	r.adjustByRouting(dn.api.meta.ID, req, dn, requestTag)
}

func (r *dispatcher) adjustByRouting(apiID uint64, req *fasthttp.Request, dn *dispathNode, requestTag string) {
	for _, routing := range r.routings {
		if routing.isUp() && routing.matches(apiID, req, requestTag) {
			log.Infof("%s: match routing %s, %s traffic to cluster %d",
				requestTag,
				routing.meta.Name,
				routing.meta.Status.String(),
				routing.meta.ClusterID)

			svr := r.selectServerFromCluster(req, routing.meta.ClusterID)

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

func (r *dispatcher) selectServerFromCluster(req *fasthttp.Request, id uint64) *serverRuntime {
	cluster, ok := r.clusters[id]
	if !ok {
		return nil
	}

	sid := cluster.selectServer(req)
	return r.servers[sid]
}
