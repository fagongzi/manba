package model

import (
	"errors"
	"regexp"
	"sync"
	"time"

	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/goetty"
	"github.com/valyala/fasthttp"
)

var (
	// ErrServerExists Server already exist
	ErrServerExists = errors.New("Server already exist")
	// ErrClusterExists Cluster already exist
	ErrClusterExists = errors.New("Cluster already exist")
	// ErrBindExists Bind already exist
	ErrBindExists = errors.New("Bind already exist")
	// ErrAggregationExists Aggregation already exist
	ErrAggregationExists = errors.New("Aggregation already exist")
	// ErrRoutingExists Routing already exist
	ErrRoutingExists = errors.New("Routing already exist")
	// ErrServerNotFound Server not found
	ErrServerNotFound = errors.New("Server not found")
	// ErrClusterNotFound Cluster not found
	ErrClusterNotFound = errors.New("Cluster not found")
	// ErrBindNotFound Bind not found
	ErrBindNotFound = errors.New("Bind not found")
	// ErrAggregationNotFound Aggregation not found
	ErrAggregationNotFound = errors.New("Aggregation not found")
	// ErrRoutingNotFound Routing not found
	ErrRoutingNotFound = errors.New("Routing not found")
)

// RouteResult RouteResult
type RouteResult struct {
	Aggregation *Aggregation
	Node        *Node
	Svr         *Server
	Err         error
	Code        int
	Res         *fasthttp.Response
	Merge       bool
}

// Release release resp
func (result *RouteResult) Release() {
	if nil != result.Res {
		fasthttp.ReleaseResponse(result.Res)
	}
}

// NeedRewrite need rewrite
func (result *RouteResult) NeedRewrite() bool {
	return result.Node != nil && result.Node.Rewrite != ""
}

// GetRealPath get real path
func (result *RouteResult) GetRealPath(req *fasthttp.Request) string {
	if nil != result.Node {
		return result.Aggregation.getNodeURL(req, result.Node)
	}

	return ""
}

// RouteTable route table
type RouteTable struct {
	rwLock *sync.RWMutex

	clusters     map[string]*Cluster
	svrs         map[string]*Server
	mapping      map[string]map[string]*Cluster
	aggregations map[string]*Aggregation
	routings     map[string]*Routing

	tw             *goetty.HashedTimeWheel
	evtChan        chan *Server
	store          Store
	watchStopCh    chan bool
	watchReceiveCh chan *Evt

	analysiser *Analysis
}

// NewRouteTable create a new RouteTable
func NewRouteTable(store Store) *RouteTable {
	tw := goetty.NewHashedTimeWheel(time.Second, 60, 3)
	tw.Start()

	rt := &RouteTable{
		tw:         tw,
		store:      store,
		analysiser: newAnalysis(),

		rwLock: &sync.RWMutex{},

		clusters:     make(map[string]*Cluster),
		svrs:         make(map[string]*Server),
		aggregations: make(map[string]*Aggregation),
		routings:     make(map[string]*Routing),
		mapping:      make(map[string]map[string]*Cluster), // serverAddr -> map[clusterName]*Cluster

		evtChan:        make(chan *Server, 1024),
		watchStopCh:    make(chan bool),
		watchReceiveCh: make(chan *Evt),
	}

	go rt.changed()
	go rt.watch()

	return rt
}

// GetServer return server
func (r *RouteTable) GetServer(addr string) *Server {
	return r.svrs[addr]
}

// AddNewRouting add a new route
func (r *RouteTable) AddNewRouting(routing *Routing) error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()

	err := routing.Check()

	if nil != err {
		return err
	}

	_, ok := r.routings[routing.ID]

	if ok {
		return ErrRoutingExists
	}

	r.routings[routing.ID] = routing

	log.Infof("Routing <%s> added", routing.Cfg)

	return nil
}

// DeleteRouting delete a route
func (r *RouteTable) DeleteRouting(id string) error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()

	route, ok := r.routings[id]

	if !ok {
		return ErrRoutingNotFound
	}

	delete(r.routings, id)

	log.Infof("Routing <%s> deleted", route.Cfg)

	return nil
}

// AddNewAggregation add a new aggregation
func (r *RouteTable) AddNewAggregation(ang *Aggregation) error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()

	_, ok := r.aggregations[ang.URL]

	if ok {
		return ErrAggregationExists
	}

	ang.Pattern = regexp.MustCompile(ang.URL)

	r.aggregations[ang.URL] = ang

	log.Infof("Aggregation <%s> added", ang.URL)

	return nil
}

// UpdateAggregation update aggregation
func (r *RouteTable) UpdateAggregation(ang *Aggregation) error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()

	old, ok := r.aggregations[ang.URL]

	if !ok {
		return ErrAggregationNotFound
	}

	old.updateFrom(ang)

	log.Infof("Aggregation <%s> updated", ang.URL)

	return nil
}

// DeleteAggregation delete a aggregation using url
func (r *RouteTable) DeleteAggregation(url string) error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()

	_, ok := r.aggregations[url]

	if !ok {
		return ErrAggregationNotFound
	}

	delete(r.aggregations, url)

	log.Infof("Aggregation <%s> deleted", url)

	return nil
}

// UpdateServer update server
func (r *RouteTable) UpdateServer(svr *Server) error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()

	old, ok := r.svrs[svr.Addr]

	if !ok {
		return ErrServerNotFound
	}

	old.updateFrom(svr)

	log.Infof("Server <%s> updated", svr.Addr)

	return nil
}

// DeleteServer delete a server
func (r *RouteTable) DeleteServer(serverAddr string) error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()

	svr, ok := r.svrs[serverAddr]

	if !ok {
		return ErrServerNotFound
	}

	delete(r.svrs, serverAddr)

	// TODO: delete aggregations

	svr.stopCheck()
	r.removeFromCheck(svr)

	binded, _ := r.mapping[svr.Addr]
	delete(r.mapping, svr.Addr)
	log.Infof("Bind <%s> stored all removed.", svr.Addr)

	for _, cluster := range binded {
		cluster.unbind(svr)
	}

	log.Infof("Server <%s> deleted", svr.Addr)

	return nil
}

// AddNewServer add a new server
func (r *RouteTable) AddNewServer(svr *Server) error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()

	_, ok := r.svrs[svr.Addr]

	if ok {
		return ErrServerExists
	}

	svr.prevStatus = Down
	svr.Status = Down
	svr.useCheckDuration = svr.CheckDuration
	r.svrs[svr.Addr] = svr

	binded := make(map[string]*Cluster)
	r.mapping[svr.Addr] = binded

	svr.init()

	// start check
	r.addToCheck(svr)

	r.analysiser.addNewAnalysis(svr.Addr)
	// 1 secs default add to use
	r.analysiser.AddRecentCount(svr.Addr, 1)

	log.Infof("Server <%s> added", svr.Addr)

	return nil
}

// UpdateCluster update cluster
func (r *RouteTable) UpdateCluster(cluster *Cluster) error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()

	old, ok := r.clusters[cluster.Name]

	if !ok {
		return ErrClusterNotFound
	}

	old.updateFrom(cluster)

	log.Infof("Cluster <%s> updated", cluster.Name)

	return nil
}

// DeleteCluster delete a cluster
func (r *RouteTable) DeleteCluster(clusterName string) error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()

	cluster, ok := r.clusters[clusterName]

	if !ok {
		return ErrClusterNotFound
	}

	cluster.doInEveryBindServers(func(addr string) {
		if svr, ok := r.svrs[addr]; ok {
			r.doUnBind(svr, cluster, false)
		}
	})

	delete(r.clusters, cluster.Name)

	// TODO: Aggregation node loose cluster

	log.Infof("Cluster <%s> deleted", cluster.Name)

	return nil
}

// AddNewCluster add new cluster
func (r *RouteTable) AddNewCluster(cluster *Cluster) error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()

	_, ok := r.clusters[cluster.Name]

	if ok {
		log.Errorf("Cluster <%v> added fail: %s", cluster, ErrClusterExists.Error())
		return ErrClusterExists
	}

	r.clusters[cluster.Name] = cluster

	log.Infof("Cluster <%s> added", cluster.Name)

	return nil
}

// Bind bind server and cluster
func (r *RouteTable) Bind(svrAddr string, clusterName string) error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()

	svr, ok := r.svrs[svrAddr]
	if !ok {
		log.Errorf("Bind <%s,%s> fail: %s", svrAddr, clusterName, ErrServerNotFound.Error())

		return ErrServerNotFound
	}

	cluster, ok := r.clusters[clusterName]
	if !ok {
		log.Errorf("Bind <%s,%s> fail: %s", svrAddr, clusterName, ErrClusterNotFound.Error())

		return ErrClusterNotFound
	}

	binded, _ := r.mapping[svr.Addr]
	bindCluster, ok := binded[cluster.Name]

	if ok && bindCluster.Name == clusterName {
		log.Errorf("Bind <%s,%s> fail: %s", svrAddr, clusterName, ErrBindExists.Error())

		return ErrBindExists
	}

	binded[cluster.Name] = cluster

	log.Infof("Bind <%s,%s> stored.", svrAddr, clusterName)

	if svr.Status == Up {
		cluster.bind(svr)
	}

	return nil
}

// UnBind unbind cluster and server
func (r *RouteTable) UnBind(svrAddr string, clusterName string) error {
	r.rwLock.Lock()
	defer r.rwLock.Unlock()

	svr, ok := r.svrs[svrAddr]
	if !ok {
		log.Errorf("UnBind <%s,%s> fail: %s", svrAddr, clusterName, ErrServerNotFound.Error())

		return ErrServerNotFound
	}

	cluster, ok := r.clusters[clusterName]
	if !ok {
		log.Errorf("UnBind <%s,%s> fail: %s", svrAddr, clusterName, ErrClusterNotFound.Error())

		return ErrClusterNotFound
	}

	r.doUnBind(svr, cluster, true)

	return nil
}

func (r *RouteTable) doUnBind(svr *Server, cluster *Cluster, withLock bool) {
	if binded, ok := r.mapping[svr.Addr]; ok {
		delete(binded, cluster.Name)
		log.Infof("Bind <%s,%s> stored removed.", svr.Addr, cluster.Name)
		if withLock {
			cluster.unbind(svr)
		} else {
			cluster.doUnBind(svr)
		}
	}
}

// Select return route result
func (r *RouteTable) Select(req *fasthttp.Request) []*RouteResult {
	r.rwLock.RLock()

	matches, results := r.selectAggregation(req)

	if matches {
		r.rwLock.RUnlock()
		return results
	}

	var targetCluster *Cluster

	for _, routing := range r.routings {
		if routing.Matches(req) {
			targetCluster = r.clusters[routing.ClusterName]
			break
		}
	}

	if nil != targetCluster {
		r.rwLock.RUnlock()
		return []*RouteResult{&RouteResult{Svr: r.doSelectServer(req, targetCluster)}}
	}

	for _, cluster := range r.clusters {
		svr := r.selectServer(req, cluster)

		if nil != svr {
			r.rwLock.RUnlock()
			return []*RouteResult{&RouteResult{Svr: svr}}
		}
	}

	r.rwLock.RUnlock()
	return nil
}

func (r *RouteTable) selectAggregation(req *fasthttp.Request) (matches bool, results []*RouteResult) {
	matches = false

	for _, agn := range r.aggregations {
		if agn.matches(req) {
			matches = true
			results = make([]*RouteResult, len(agn.Nodes))

			for index, node := range agn.Nodes {
				results[index] = &RouteResult{
					Aggregation: agn,
					Node:        node,
					Svr:         r.selectServer(req, r.clusters[node.ClusterName]),
				}
			}
		}
	}

	return matches, results
}

func (r *RouteTable) selectServer(req *fasthttp.Request, cluster *Cluster) *Server {
	if cluster.Matches(req) {
		return r.doSelectServer(req, cluster)
	}

	return nil
}

func (r *RouteTable) doSelectServer(req *fasthttp.Request, cluster *Cluster) *Server {
	addr := cluster.Select(req) // 这里有可能会被锁住，会被正在修改bind关系的cluster锁住
	svr, _ := r.svrs[addr]
	return svr
}

// GetAnalysis return analysis
func (r *RouteTable) GetAnalysis() *Analysis {
	return r.analysiser
}

// GetTimeWheel return time wheel
func (r *RouteTable) GetTimeWheel() *goetty.HashedTimeWheel {
	return r.tw
}

func (r *RouteTable) watch() {
	log.Info("RouteTable start watch.")

	go r.doEvtReceive()
	err := r.store.Watch(r.watchReceiveCh, r.watchStopCh)

	log.Errorf("RouteTable watch err: %s", err)
}

func (r *RouteTable) doEvtReceive() {
	for {
		evt := <-r.watchReceiveCh

		if evt.Src == EventSrcCluster {
			r.doReceiveCluster(evt)
		} else if evt.Src == EventSrcServer {
			r.doReceiveServer(evt)
		} else if evt.Src == EventSrcBind {
			r.doReceiveBind(evt)
		} else if evt.Src == EventSrcAggregation {
			r.doReceiveAggregation(evt)
		} else if evt.Src == EventSrcRouting {
			r.doReceiveRouting(evt)
		} else {
			log.Warnf("EVT unknown <%+v>", evt)
		}
	}
}

func (r *RouteTable) doReceiveRouting(evt *Evt) {
	routing, _ := evt.Value.(*Routing)

	if evt.Type == EventTypeNew {
		r.AddNewRouting(routing)
	} else if evt.Type == EventTypeDelete {
		r.DeleteRouting(evt.Key)
	} else if evt.Type == EventTypeUpdate {
		// TODO: impl
	}
}

func (r *RouteTable) doReceiveAggregation(evt *Evt) {
	ang, _ := evt.Value.(*Aggregation)

	if evt.Type == EventTypeNew {
		r.AddNewAggregation(ang)
	} else if evt.Type == EventTypeDelete {
		r.DeleteAggregation(evt.Key)
	} else if evt.Type == EventTypeUpdate {
		r.UpdateAggregation(ang)
	}
}

func (r *RouteTable) doReceiveCluster(evt *Evt) {
	cluster, _ := evt.Value.(*Cluster)

	if evt.Type == EventTypeNew {
		r.AddNewCluster(cluster)
	} else if evt.Type == EventTypeDelete {
		r.DeleteCluster(evt.Key)
	} else if evt.Type == EventTypeUpdate {
		r.UpdateCluster(cluster)
	}
}

func (r *RouteTable) doReceiveServer(evt *Evt) {
	svr, _ := evt.Value.(*Server)

	if evt.Type == EventTypeNew {
		r.AddNewServer(svr)
	} else if evt.Type == EventTypeDelete {
		r.DeleteServer(evt.Key)
	} else if evt.Type == EventTypeUpdate {
		r.UpdateServer(svr)
	}
}

func (r *RouteTable) doReceiveBind(evt *Evt) {
	bind, _ := evt.Value.(*Bind)

	if evt.Type == EventTypeNew {
		r.Bind(bind.ServerAddr, bind.ClusterName)
	} else if evt.Type == EventTypeDelete {
		r.UnBind(bind.ServerAddr, bind.ClusterName)
	}
}

// Load load info from store
func (r *RouteTable) Load() {
	r.loadClusters()
	r.loadServers()
	r.loadBinds()
	r.loadAggregations()
	r.loadRoutings()
}

func (r *RouteTable) loadClusters() {
	clusters, err := r.store.GetClusters()
	if nil != err {
		log.WarnErrorf(err, "Load clusters fail.")
		return
	}

	for _, cluster := range clusters {
		err := r.AddNewCluster(cluster)
		if nil != err {
			log.PanicErrorf(err, "Server <%s> add fail.", cluster.Name)
		}
	}
}

func (r *RouteTable) loadServers() {
	servers, err := r.store.GetServers()
	if nil != err {
		log.WarnErrorf(err, "Load servers from etcd fail.")
		return
	}

	for _, server := range servers {
		err := r.AddNewServer(server)

		if nil != err {
			log.PanicErrorf(err, "Server <%s> add fail.", server.Addr)
		}
	}
}

func (r *RouteTable) loadRoutings() {
	routings, err := r.store.GetRoutings()
	if nil != err {
		log.WarnErrorf(err, "Load routings from etcd fail.")
		return
	}

	for _, route := range routings {
		err := r.AddNewRouting(route)
		if nil != err {
			log.PanicError(err, "Routing <%s> add fail.", route.Cfg)
		}
	}
}

func (r *RouteTable) loadBinds() {
	binds, err := r.store.GetBinds()
	if nil != err {
		log.WarnErrorf(err, "Load binds from etcd fail.")
		return
	}

	for _, b := range binds {
		err := r.Bind(b.ServerAddr, b.ClusterName)
		if nil != err {
			log.WarnErrorf(err, "Bind <%s, %s> add fail.", b.ServerAddr, b.ClusterName)
		}
	}
}

func (r *RouteTable) loadAggregations() {
	angs, err := r.store.GetAggregations()
	if nil != err {
		log.WarnErrorf(err, "Load aggregations from etcd fail.")
		return
	}

	for _, ang := range angs {
		err := r.AddNewAggregation(ang)
		if nil != err {
			log.PanicError(err, "Aggregation <%s> add fail.", ang.URL)
		}
	}
}

func (r *RouteTable) removeFromCheck(svr *Server) {
	r.tw.Cancel(svr.Addr)
}

func (r *RouteTable) addToCheck(svr *Server) {
	r.tw.AddWithId(time.Duration(svr.useCheckDuration)*time.Second, svr.Addr, r.check)
}

func (r *RouteTable) check(addr string) {
	svr, _ := r.svrs[addr]

	if svr.check(r.addToCheck) {
		svr.changeTo(Up)

		if svr.statusChanged() {
			log.Infof("Server <%s> UP.", svr.Addr)
		}
	} else {
		svr.changeTo(Down)

		if svr.statusChanged() {
			log.Warnf("Server <%s, %s> DOWN.", svr.Addr, svr.CheckPath)
		}
	}

	r.evtChan <- svr
}

func (r *RouteTable) changed() {
	for {
		svr := <-r.evtChan

		if svr.statusChanged() {
			binded := r.mapping[svr.Addr]

			if svr.Status == Up {
				for _, c := range binded {
					c.bind(svr)
				}
			} else {
				for _, c := range binded {
					c.unbind(svr)
				}
			}
		}
	}
}
