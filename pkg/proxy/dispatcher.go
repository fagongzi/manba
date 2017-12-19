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

type dispathNode struct {
	api  *apiRuntime
	node *apiNode
	dest *serverRuntime
	res  *fasthttp.Response
	err  error
	code int
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

func (r *dispatcher) dispatch(req *fasthttp.Request) []*dispathNode {
	r.RLock()

	var dispathes []*dispathNode
	for _, api := range r.apis {
		if api.matches(req) {
			for _, node := range api.nodes {
				dispathes = append(dispathes, &dispathNode{
					api:  api,
					node: node,
					dest: r.selectServer(req, r.selectClusterByRouting(req, r.clusters[node.meta.ClusterID])),
				})
			}
			break
		}
	}

	r.RUnlock()
	return dispathes
}

func (r *dispatcher) selectClusterByRouting(req *fasthttp.Request, src *clusterRuntime) *clusterRuntime {
	targetCluster := src

	for _, routing := range r.routings {
		if routing.matches(req) {
			targetCluster = r.clusters[routing.meta.ClusterID]
			break
		}
	}

	return targetCluster
}

func (r *dispatcher) selectServer(req *fasthttp.Request, cluster *clusterRuntime) *serverRuntime {
	id := cluster.selectServer(req)
	svr, _ := r.servers[id]
	return svr
}
