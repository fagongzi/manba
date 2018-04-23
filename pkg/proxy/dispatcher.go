package proxy

import (
	"sync"
	"time"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/store"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/fagongzi/goetty"
	"github.com/fagongzi/util/task"
	"github.com/valyala/fasthttp"
)

type copyReq struct {
	origin *fasthttp.Request
	api    *apiRuntime
	node   *apiNode
	to     *serverRuntime
}

func (req *copyReq) prepare() {
	if req.needRewrite() {
		// if not use rewrite, it only change uri path and query string
		realPath := req.rewiteURL()
		if "" != realPath {
			req.origin.SetRequestURI(realPath)
			req.origin.SetHost(req.to.meta.Addr)
		}
	}
}

func (req *copyReq) needRewrite() bool {
	return req.node.meta.URLRewrite != ""
}

func (req *copyReq) rewiteURL() string {
	return req.api.rewriteURL(req.origin, req.node.meta.URLRewrite)
}

type dispathNode struct {
	ctx *fasthttp.RequestCtx
	wg  *sync.WaitGroup

	api                  *apiRuntime
	node                 *apiNode
	dest                 *serverRuntime
	copyTo               *serverRuntime
	res                  *fasthttp.Response
	cachedBody, cachedCT []byte
	err                  error
	code                 int
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
	return dn.api.rewriteURL(req, dn.node.meta.URLRewrite)
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

func (dn *dispathNode) getResponseBody() []byte {
	if dn.node.meta.UseDefault ||
		(dn.hasError() && dn.hasDefaultValue()) {
		return dn.node.meta.DefaultValue.Body
	}

	if len(dn.cachedBody) > 0 {
		return dn.cachedBody
	}

	if nil != dn.res {
		return dn.res.Body()
	}

	return nil
}

func (dn *dispathNode) maybeDone() {
	if nil != dn.wg {
		defer dn.wg.Done()
	}
}

type dispatcher struct {
	sync.RWMutex

	cnf         *Cfg
	routings    map[uint64]*routingRuntime
	apis        map[uint64]*apiRuntime
	clusters    map[uint64]*clusterRuntime
	servers     map[uint64]*serverRuntime
	binds       map[uint64]map[uint64]*clusterRuntime
	proxies     map[string]*metapb.Proxy
	checkerC    chan uint64
	watchStopC  chan bool
	watchEventC chan *store.Evt
	analysiser  *util.Analysis
	store       store.Store
	httpClient  *util.FastHTTPClient
	tw          *goetty.TimeoutWheel
	runner      *task.Runner
}

func newDispatcher(cnf *Cfg, db store.Store, runner *task.Runner) *dispatcher {
	tw := goetty.NewTimeoutWheel(goetty.WithTickInterval(time.Second))
	rt := &dispatcher{
		cnf:         cnf,
		tw:          tw,
		store:       db,
		runner:      runner,
		analysiser:  util.NewAnalysis(tw),
		httpClient:  util.NewFastHTTPClient(),
		clusters:    make(map[uint64]*clusterRuntime),
		servers:     make(map[uint64]*serverRuntime),
		apis:        make(map[uint64]*apiRuntime),
		routings:    make(map[uint64]*routingRuntime),
		binds:       make(map[uint64]map[uint64]*clusterRuntime),
		proxies:     make(map[string]*metapb.Proxy),
		checkerC:    make(chan uint64, 1024),
		watchStopC:  make(chan bool),
		watchEventC: make(chan *store.Evt),
	}

	rt.readyToHeathChecker()
	return rt
}

func (r *dispatcher) dispatch(req *fasthttp.Request) (*apiRuntime, []*dispathNode) {
	r.RLock()

	var targetAPI *apiRuntime
	var dispathes []*dispathNode
	for _, api := range r.apis {
		if api.matches(req) {
			targetAPI = api
			if api.meta.UseDefault {
				break
			}

			for _, node := range api.nodes {
				dn := &dispathNode{
					api:  api,
					node: node,
					dest: r.selectServer(req, r.clusters[node.meta.ClusterID]),
				}

				r.routingOpt(api.meta.ID, req, dn)
				dispathes = append(dispathes, dn)
			}
			break
		}
	}

	r.RUnlock()
	return targetAPI, dispathes
}

func (r *dispatcher) routingOpt(apiID uint64, req *fasthttp.Request, node *dispathNode) {
	for _, routing := range r.routings {
		if routing.matches(apiID, req) {
			switch routing.meta.Strategy {
			case metapb.Split:
				node.dest = r.selectServer(req, r.clusters[routing.meta.ClusterID])
			case metapb.Copy:
				node.copyTo = r.selectServer(req, r.clusters[routing.meta.ClusterID])
			}
			break
		}
	}
}

func (r *dispatcher) selectServer(req *fasthttp.Request, cluster *clusterRuntime) *serverRuntime {
	id := cluster.selectServer(req)
	svr, _ := r.servers[id]
	return svr
}
