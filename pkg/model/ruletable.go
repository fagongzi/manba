package model

import (
	"errors"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/net"
	"net/http"
	"sync"
	"time"
)

var (
	ERR_SERVER_EXISTS      = errors.New("Server already exist.")
	ERR_CLUSTER_EXISTS     = errors.New("Cluster already exist.")
	ERR_BIND_EXISTS        = errors.New("Bind already exist")
	ERR_AGGREGATION_EXISTS = errors.New("Aggregation already exist.")

	ERR_SERVER_NOT_FOUND      = errors.New("Server not found.")
	ERR_CLUSTER_NOT_FOUND     = errors.New("Cluster not found.")
	ERR_BIND_NOT_FOUND        = errors.New("Bind not found")
	ERR_AGGREGATION_NOT_FOUND = errors.New("Aggregation not found.")
)

type RouteResult struct {
	Node  *Node
	Svr   *Server
	Err   error
	Code  int
	Res   *http.Response
	Merge bool
}

type RouteTable struct {
	rwLock *sync.RWMutex

	clusters     map[string]*Cluster
	svrs         map[string]*Server
	mapping      map[string]map[string]*Cluster
	aggregations map[string]*Aggregation

	tw             *net.HashedTimeWheel
	evtChan        chan *Server
	store          Store
	watchStopCh    chan bool
	watchReceiveCh chan *Evt

	analysiser *Analysis
}

func NewRouteTable(store Store) *RouteTable {
	tw := net.NewHashedTimeWheel(time.Second, 60, 3)
	tw.Start()

	rt := &RouteTable{
		tw:         tw,
		store:      store,
		analysiser: newAnalysis(),

		rwLock: &sync.RWMutex{},

		clusters:     make(map[string]*Cluster),
		svrs:         make(map[string]*Server),
		aggregations: make(map[string]*Aggregation),
		mapping:      make(map[string]map[string]*Cluster), // serverAddr -> map[clusterName]*Cluster

		evtChan:        make(chan *Server, 1024),
		watchStopCh:    make(chan bool),
		watchReceiveCh: make(chan *Evt),
	}

	go rt.changed()
	go rt.watch()

	return rt
}

func (self *RouteTable) GetServer(addr string) *Server {
	return self.svrs[addr]
}

func (self *RouteTable) AddNewAggregation(ang *Aggregation) error {
	self.rwLock.Lock()
	defer self.rwLock.Unlock()

	_, ok := self.aggregations[ang.Url]

	if ok {
		return ERR_AGGREGATION_EXISTS
	}

	self.aggregations[ang.Url] = ang

	return nil
}

func (self *RouteTable) UpdateAggregation(ang *Aggregation) error {
	self.rwLock.Lock()
	defer self.rwLock.Unlock()

	old, ok := self.aggregations[ang.Url]

	if !ok {
		return ERR_AGGREGATION_NOT_FOUND
	}

	old.updateFrom(ang)

	return nil
}

func (self *RouteTable) DeleteAggregation(url string) error {
	self.rwLock.Lock()
	defer self.rwLock.Unlock()

	_, ok := self.aggregations[url]

	if !ok {
		return ERR_AGGREGATION_NOT_FOUND
	}

	delete(self.aggregations, url)

	return nil
}

func (self *RouteTable) UpdateServer(svr *Server) error {
	self.rwLock.Lock()
	defer self.rwLock.Unlock()

	old, ok := self.svrs[svr.Addr]

	if !ok {
		return ERR_SERVER_NOT_FOUND
	}

	old.updateFrom(svr)

	return nil
}

func (self *RouteTable) DeleteServer(serverAddr string) error {
	self.rwLock.Lock()
	defer self.rwLock.Unlock()

	log.Infof("Server delete start: <%s>", serverAddr)

	svr, ok := self.svrs[serverAddr]

	if !ok {
		return ERR_SERVER_NOT_FOUND
	}

	delete(self.svrs, serverAddr)

	// TODO: delete aggregations

	svr.stopCheck()
	self.removeFromCheck(svr)

	binded, _ := self.mapping[svr.Addr]
	delete(self.mapping, svr.Addr)
	log.Infof("Bind <%s> stored all removed.", svr.Addr)

	for _, cluster := range binded {
		cluster.unbind(svr)
	}

	return nil
}

func (self *RouteTable) AddNewServer(svr *Server) error {
	self.rwLock.Lock()
	defer self.rwLock.Unlock()

	_, ok := self.svrs[svr.Addr]

	if ok {
		return ERR_SERVER_EXISTS
	}

	svr.prevStatus = DOWN
	svr.Status = DOWN
	svr.useCheckDuration = svr.CheckDuration
	self.svrs[svr.Addr] = svr

	binded := make(map[string]*Cluster)
	self.mapping[svr.Addr] = binded

	svr.init()

	// start check
	self.addToCheck(svr)

	self.analysiser.addNewAnalysis(svr.Addr)
	// 1 secs default add to use
	self.analysiser.AddRecentCount(svr.Addr, 1)

	log.Infof("Server <%s> added", svr.Addr)

	return nil
}

func (self *RouteTable) UpdateCluster(cluster *Cluster) error {
	self.rwLock.Lock()
	defer self.rwLock.Unlock()

	old, ok := self.clusters[cluster.Name]

	if !ok {
		return ERR_CLUSTER_NOT_FOUND
	}

	old.updateFrom(cluster)

	return nil
}

func (self *RouteTable) DeleteCluster(clusterName string) error {
	self.rwLock.Lock()
	defer self.rwLock.Unlock()

	cluster, ok := self.clusters[clusterName]

	if !ok {
		return ERR_CLUSTER_NOT_FOUND
	}

	cluster.doInEveryBindServers(func(addr string) {
		svr, ok := self.svrs[addr]
		if ok {
			self.doUnBind(svr, cluster, false)
		}
	})

	delete(self.clusters, cluster.Name)

	// TODO Aggregation node loose cluster

	return nil
}

func (self *RouteTable) AddNewCluster(cluster *Cluster) error {
	self.rwLock.Lock()
	defer self.rwLock.Unlock()

	_, ok := self.clusters[cluster.Name]

	if ok {
		log.Errorf("Cluster <%v> added fail: %s", cluster, ERR_CLUSTER_EXISTS.Error())
		return ERR_CLUSTER_EXISTS
	}

	self.clusters[cluster.Name] = cluster

	log.Infof("Cluster <%s> added", cluster.Name)

	return nil
}

func (self *RouteTable) Bind(svrAddr string, clusterName string) error {
	self.rwLock.Lock()
	defer self.rwLock.Unlock()

	svr, ok := self.svrs[svrAddr]
	if !ok {
		log.Errorf("Bind <%s,%s> fail: %s", svrAddr, clusterName, ERR_SERVER_NOT_FOUND.Error())

		return ERR_SERVER_NOT_FOUND
	}

	cluster, ok := self.clusters[clusterName]
	if !ok {
		log.Errorf("Bind <%s,%s> fail: %s", svrAddr, clusterName, ERR_CLUSTER_NOT_FOUND.Error())

		return ERR_CLUSTER_NOT_FOUND
	}

	binded, _ := self.mapping[svr.Addr]
	bindCluster, ok := binded[cluster.Name]

	if ok && bindCluster.Name == clusterName {
		log.Errorf("Bind <%s,%s> fail: %s", svrAddr, clusterName, ERR_BIND_EXISTS.Error())

		return ERR_BIND_EXISTS
	}

	binded[cluster.Name] = cluster

	log.Infof("Bind <%s,%s> stored.", svrAddr, clusterName)

	if svr.Status == UP {
		cluster.bind(svr)
	}

	return nil
}

func (self *RouteTable) UnBind(svrAddr string, clusterName string) error {
	self.rwLock.Lock()
	defer self.rwLock.Unlock()

	svr, ok := self.svrs[svrAddr]
	if !ok {
		log.Errorf("UnBind <%s,%s> fail: %s", svrAddr, clusterName, ERR_SERVER_NOT_FOUND.Error())

		return ERR_SERVER_NOT_FOUND
	}

	cluster, ok := self.clusters[clusterName]
	if !ok {
		log.Errorf("UnBind <%s,%s> fail: %s", svrAddr, clusterName, ERR_CLUSTER_NOT_FOUND.Error())

		return ERR_CLUSTER_NOT_FOUND
	}

	self.doUnBind(svr, cluster, true)

	return nil
}

func (self *RouteTable) doUnBind(svr *Server, cluster *Cluster, withLock bool) {
	binded, ok := self.mapping[svr.Addr]
	if ok {
		delete(binded, cluster.Name)
		log.Infof("Bind <%s,%s> stored removed.", svr.Addr, cluster.Name)
		if withLock {
			cluster.unbind(svr)
		} else {
			cluster.doUnBind(svr)
		}
	}
}

func (self *RouteTable) Select(req *http.Request) []*RouteResult {
	self.rwLock.RLock()
	defer self.rwLock.RUnlock()

	matches, results := self.selectAggregation(req)

	if matches {
		return results
	}

	for _, cluster := range self.clusters {
		svr := self.selectServer(req, cluster)

		if nil != svr {
			return []*RouteResult{&RouteResult{Svr: svr}}
		}
	}

	return nil
}

func (self *RouteTable) selectAggregation(req *http.Request) (matches bool, results []*RouteResult) {
	matches = false

	for url, agn := range self.aggregations {
		if url == req.URL.Path {
			matches = true
			results = make([]*RouteResult, len(agn.Nodes))

			for index, node := range agn.Nodes {
				results[index] = &RouteResult{
					Node: node,
					Svr:  self.selectServer(req, self.clusters[node.ClusterName]),
				}
			}
		}
	}

	return matches, results
}

func (self *RouteTable) selectServer(req *http.Request, cluster *Cluster) *Server {
	if cluster.Matches(req) {
		addr := cluster.Select(req) // 这里有可能会被锁住，会被正在修改bind关系的cluster锁住
		svr, _ := self.svrs[addr]
		return svr
	}

	return nil
}

func (self *RouteTable) GetAnalysis() *Analysis {
	return self.analysiser
}

func (self *RouteTable) GetTimeWheel() *net.HashedTimeWheel {
	return self.tw
}

func (self *RouteTable) watch() {
	log.Info("RouteTable start watch.")

	go self.doEvtReceive()
	err := self.store.Watch(self.watchReceiveCh, self.watchStopCh)

	log.Errorf("RouteTable watch err: %s", err)
}

func (self *RouteTable) doEvtReceive() {
	for {
		evt := <-self.watchReceiveCh

		if evt.Src == EVT_SRC_CLUSTER {
			self.doReceiveCluster(evt)
		} else if evt.Src == EVT_SRC_SERVER {
			self.doReceiveServer(evt)
		} else if evt.Src == EVT_SRC_BIND {
			self.doReceiveBind(evt)
		} else if evt.Src == EVT_STC_AGGREGATION {
			self.doReceiveAggregation(evt)
		} else {
			log.Warnf("EVT unknown <%+v>", evt)
		}
	}
}

func (self *RouteTable) doReceiveAggregation(evt *Evt) {
	ang, _ := evt.Value.(*Aggregation)

	if evt.Type == EVT_TYPE_NEW {
		self.AddNewAggregation(ang)
	} else if evt.Type == EVT_TYPE_DELETE {
		self.DeleteAggregation(evt.Key)
	} else if evt.Type == EVT_TYPE_UPDATE {
		self.UpdateAggregation(ang)
	}
}

func (self *RouteTable) doReceiveCluster(evt *Evt) {
	cluster, _ := evt.Value.(*Cluster)

	if evt.Type == EVT_TYPE_NEW {
		self.AddNewCluster(cluster)
	} else if evt.Type == EVT_TYPE_DELETE {
		self.DeleteCluster(evt.Key)
	} else if evt.Type == EVT_TYPE_UPDATE {
		self.UpdateCluster(cluster)
	}
}

func (self *RouteTable) doReceiveServer(evt *Evt) {
	svr, _ := evt.Value.(*Server)

	if evt.Type == EVT_TYPE_NEW {
		self.AddNewServer(svr)
	} else if evt.Type == EVT_TYPE_DELETE {
		self.DeleteServer(evt.Key)
	} else if evt.Type == EVT_TYPE_UPDATE {
		self.UpdateServer(svr)
	}
}

func (self *RouteTable) doReceiveBind(evt *Evt) {
	bind, _ := evt.Value.(*Bind)

	if evt.Type == EVT_TYPE_NEW {
		self.Bind(bind.ServerAddr, bind.ClusterName)
	} else if evt.Type == EVT_TYPE_DELETE {
		self.UnBind(bind.ServerAddr, bind.ClusterName)
	}
}

func (self *RouteTable) Load() {
	self.loadClusters()
	self.loadServers()
	self.loadBinds()
	self.loadAggregations()
}

func (self *RouteTable) loadClusters() {
	clusters, err := self.store.GetClusters()
	if nil != err {
		log.PanicError(err, "Load clusters fail.")
	}

	for _, cluster := range clusters {
		err := self.AddNewCluster(cluster)
		if nil != err {
			log.PanicErrorf(err, "Server <%s> add fail.", cluster.Name)
		}
	}
}

func (self *RouteTable) loadServers() {
	servers, err := self.store.GetServers()
	if nil != err {
		log.PanicError(err, "Load servers from etcd fail.")
	}

	for _, server := range servers {
		err := self.AddNewServer(server)

		if nil != err {
			log.PanicErrorf(err, "Server <%s> add fail.", server.Addr)
		}
	}
}

func (self *RouteTable) loadBinds() {
	binds, err := self.store.GetBinds()
	if nil != err {
		log.PanicError(err, "Load binds from etcd fail.")
	}

	for _, b := range binds {
		err := self.Bind(b.ServerAddr, b.ClusterName)
		if nil != err {
			log.PanicError(err, "Bind <%s, %s> add fail.", b.ServerAddr, b.ClusterName)
		}
	}
}

func (self *RouteTable) loadAggregations() {
	angs, err := self.store.GetAggregations()
	if nil != err {
		log.PanicError(err, "Load aggregations from etcd fail.")
	}

	for _, ang := range angs {
		err := self.AddNewAggregation(ang)
		if nil != err {
			log.PanicError(err, "Aggregation <%s> add fail.", ang.Url)
		}
	}
}

func (self *RouteTable) removeFromCheck(svr *Server) {
	self.tw.Cancel(svr.Addr)
}

func (self *RouteTable) addToCheck(svr *Server) {
	self.tw.AddWithId(time.Duration(svr.useCheckDuration)*time.Second, svr.Addr, self.check)
}

func (self *RouteTable) check(addr string) {
	svr, _ := self.svrs[addr]

	if svr.check(self.addToCheck) {
		svr.changeTo(UP)

		if svr.statusChanged() {
			log.Infof("Server <%s> UP.", svr.Addr)
		}
	} else {
		svr.changeTo(DOWN)

		if svr.statusChanged() {
			log.Warnf("Server <%s, %s> DOWN.", svr.Addr, svr.CheckPath)
		}
	}

	self.evtChan <- svr
}

func (self *RouteTable) changed() {
	for {
		svr := <-self.evtChan

		if svr.statusChanged() {
			binded := self.mapping[svr.Addr]

			if svr.Status == UP {
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
