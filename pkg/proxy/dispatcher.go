package proxy

import (
	"sync"
	"time"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/fagongzi/gateway/pkg/store"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/fagongzi/goetty"
	"github.com/fagongzi/util/task"
	"github.com/valyala/fasthttp"
)

type dispathNode struct {
	api   *model.API
	node  *model.Node
	dest  *serverRuntime
	res   *fasthttp.Response
	err   error
	code  int
	merge bool
}

func (dn *dispathNode) release() {
	if nil != dn.res {
		fasthttp.ReleaseResponse(dn.res)
	}
}

func (dn *dispathNode) needRewrite() bool {
	return dn.node.Rewrite != ""
}

func (dn *dispathNode) rewiteURL(req *fasthttp.Request) string {
	return dn.api.RewriteURL(req, dn.node)
}

type dispatcher struct {
	sync.RWMutex

	cnf         *Cfg
	binds       map[string]*model.Bind
	clusters    map[string]*clusterRuntime
	servers     map[string]*serverRuntime
	mapping     map[string]map[string]*clusterRuntime // map[server id]map[cluster id]cluster
	checkerC    chan string
	apis        map[string]*model.API
	routings    map[string]*model.Routing
	watchStopC  chan bool
	watchEventC chan *store.Evt
	analysiser  *model.Analysis
	store       store.Store
	httpClient  *util.FastHTTPClient
	tw          *goetty.TimeoutWheel
	taskRunner  *task.Runner
}

func newRouteTable(cnf *Cfg, db store.Store, taskRunner *task.Runner) *dispatcher {
	tw := goetty.NewTimeoutWheel(goetty.WithTickInterval(time.Second))
	tw.Start()

	rt := &dispatcher{
		cnf:        cnf,
		tw:         tw,
		store:      db,
		taskRunner: taskRunner,

		analysiser: model.NewAnalysis(taskRunner),
		httpClient: util.NewFastHTTPClient(),

		binds:       make(map[string]*model.Bind),
		clusters:    make(map[string]*clusterRuntime),
		servers:     make(map[string]*serverRuntime),
		mapping:     make(map[string]map[string]*clusterRuntime),
		apis:        make(map[string]*model.API),
		routings:    make(map[string]*model.Routing),
		checkerC:    make(chan string, 1024),
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
		if api.Matches(req) {
			dispathes = make([]*dispathNode, len(api.Nodes))

			for index, node := range api.Nodes {
				dispathes[index] = &dispathNode{
					api:  api,
					node: node,
					dest: r.selectServer(req, r.selectClusterByRouting(req, r.clusters[node.ClusterName])),
				}
			}
		}
	}

	r.RUnlock()
	return dispathes
}

func (r *dispatcher) selectClusterByRouting(req *fasthttp.Request, src *clusterRuntime) *clusterRuntime {
	targetCluster := src

	for _, routing := range r.routings {
		if routing.Matches(req) {
			targetCluster = r.clusters[routing.Cluster]
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
