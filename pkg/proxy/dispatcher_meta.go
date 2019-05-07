package proxy

import (
	"errors"
	"sort"
	"time"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/fagongzi/log"
)

var (
	errServerExists    = errors.New("Server already exist")
	errClusterExists   = errors.New("Cluster already exist")
	errBindExists      = errors.New("Bind already exist")
	errAPIExists       = errors.New("API already exist")
	errProxyExists     = errors.New("Proxy already exist")
	errRoutingExists   = errors.New("Routing already exist")
	errServerNotFound  = errors.New("Server not found")
	errClusterNotFound = errors.New("Cluster not found")
	errBindNotFound    = errors.New("Bind not found")
	errProxyNotFound   = errors.New("Proxy not found")
	errAPINotFound     = errors.New("API not found")
	errRoutingNotFound = errors.New("Routing not found")

	limit = int64(32)
)

func (r *dispatcher) load() {
	go r.watch()

	r.loadProxies()
	r.loadClusters()
	r.loadServers()
	r.loadBinds()
	r.loadAPIs()
	r.loadRoutings()
}

func (r *dispatcher) loadProxies() {
	log.Infof("load proxies")

	err := r.store.GetProxies(limit, func(value *metapb.Proxy) error {
		return r.addProxy(value)
	})
	if nil != err {
		log.Errorf("load proxies failed, errors:\n%+v",
			err)
		return
	}
}

func (r *dispatcher) loadClusters() {
	log.Infof("load clusters")

	err := r.store.GetClusters(limit, func(value interface{}) error {
		return r.addCluster(value.(*metapb.Cluster))
	})
	if nil != err {
		log.Errorf("load clusters failed, errors:\n%+v",
			err)
		return
	}
}

func (r *dispatcher) loadServers() {
	log.Infof("load servers")

	err := r.store.GetServers(limit, func(value interface{}) error {
		return r.addServer(value.(*metapb.Server))
	})
	if nil != err {
		log.Errorf("load servers failed, errors:\n%+v",
			err)
		return
	}
}

func (r *dispatcher) loadRoutings() {
	log.Infof("load routings")

	err := r.store.GetRoutings(limit, func(value interface{}) error {
		return r.addRouting(value.(*metapb.Routing))
	})
	if nil != err {
		log.Errorf("load servers failed, errors:\n%+v",
			err)
		return
	}
}

func (r *dispatcher) loadBinds() {
	log.Infof("load binds")

	for clusterID := range r.clusters {
		servers, err := r.store.GetBindServers(clusterID)
		if nil != err {
			log.Errorf("load binds from store failed, errors:\n%+v",
				err)
			return
		}

		for _, serverID := range servers {
			b := &metapb.Bind{
				ClusterID: clusterID,
				ServerID:  serverID,
			}
			err = r.addBind(b)
			if nil != err {
				log.Fatalf("bind <%s> add failed, errors:\n%+v",
					b.String(),
					err)
			}
		}
	}
}

func (r *dispatcher) loadAPIs() {
	log.Infof("load apis")

	err := r.store.GetAPIs(limit, func(value interface{}) error {
		return r.addAPI(value.(*metapb.API))
	})
	if nil != err {
		log.Errorf("load apis failed, errors:\n%+v",
			err)
		return
	}
}

func (r *dispatcher) addRouting(meta *metapb.Routing) error {
	if _, ok := r.routings[meta.ID]; ok {
		return errRoutingExists
	}

	newValues := r.copyRoutings(0)
	newValues[meta.ID] = newRoutingRuntime(meta)
	r.routings = newValues
	log.Infof("routing <%d> added, data <%s>",
		meta.ID,
		meta.String())

	return nil
}

func (r *dispatcher) updateRouting(meta *metapb.Routing) error {
	rt, ok := r.routings[meta.ID]
	if !ok {
		return errRoutingNotFound
	}

	newValues := r.copyRoutings(0)
	rt = newValues[meta.ID]
	rt.updateMeta(meta)
	r.routings = newValues

	log.Infof("routing <%d> updated, data <%s>",
		meta.ID,
		meta.String())
	return nil
}

func (r *dispatcher) removeRouting(id uint64) error {
	if _, ok := r.routings[id]; !ok {
		return errRoutingNotFound
	}

	newValues := r.copyRoutings(id)
	r.routings = newValues

	log.Infof("routing <%d> deleted",
		id)
	return nil
}

func (r *dispatcher) addProxy(meta *metapb.Proxy) error {
	key := util.GetAddrFormat(meta.Addr)

	if _, ok := r.proxies[key]; ok {
		return errProxyExists
	}

	r.proxies[key] = meta
	r.refreshAllQPS()

	log.Infof("proxy <%s> added", key)
	return nil
}

func (r *dispatcher) removeProxy(addr string) error {
	if _, ok := r.proxies[addr]; !ok {
		return errProxyNotFound
	}

	delete(r.proxies, addr)
	r.refreshAllQPS()

	log.Infof("proxy <%s> deleted", addr)
	return nil
}

func (r *dispatcher) addAPI(api *metapb.API) error {
	if _, ok := r.apis[api.ID]; ok {
		return errAPIExists
	}

	a := newAPIRuntime(api, r.tw, r.refreshQPS(api.MaxQPS))
	newValues := r.copyAPIs(0)
	newValues[api.ID] = a
	newKeys := sortAPIs(newValues)

	if a.cb != nil {
		r.addAnalysis(api.ID, a.cb)
	}

	r.apis = newValues
	r.apiSortedKeys = newKeys
	log.Infof("api <%d> added, data <%s>",
		api.ID,
		api.String())

	return nil
}

func (r *dispatcher) updateAPI(api *metapb.API) error {
	_, ok := r.apis[api.ID]
	if !ok {
		return errAPINotFound
	}

	newValues := r.copyAPIs(0)
	rt := newValues[api.ID]
	rt.activeQPS = r.refreshQPS(api.MaxQPS)
	rt.updateMeta(api)
	newKeys := sortAPIs(newValues)

	if rt.cb != nil {
		r.addAnalysis(rt.meta.ID, rt.meta.CircuitBreaker)
	}

	r.apis = newValues
	r.apiSortedKeys = newKeys
	log.Infof("api <%d> updated, data <%s>",
		api.ID,
		api.String())

	return nil
}

func (r *dispatcher) removeAPI(id uint64) error {
	if _, ok := r.apis[id]; !ok {
		return errAPINotFound
	}

	newValues := r.copyAPIs(id)
	newKeys := sortAPIs(newValues)

	r.apiSortedKeys = newKeys
	r.apis = newValues
	log.Infof("api <%d> removed", id)

	return nil
}

func (r *dispatcher) refreshAllQPS() {
	for _, svr := range r.servers {
		svr.activeQPS = r.refreshQPS(svr.meta.MaxQPS)
		svr.updateMeta(svr.meta)
		r.addToCheck(svr)
	}

	for _, api := range r.apis {
		api.activeQPS = r.refreshQPS(api.meta.MaxQPS)
		api.updateMeta(api.meta)
	}
}

func (r *dispatcher) refreshQPS(value int64) int64 {
	activeQPS := value
	if len(r.proxies) > 0 {
		activeQPS = value / int64(len(r.proxies))
	}
	if activeQPS <= 0 {
		activeQPS = 1
	}
	return activeQPS
}

func (r *dispatcher) addServer(svr *metapb.Server) error {
	if _, ok := r.servers[svr.ID]; ok {
		return errServerExists
	}

	newValues := r.copyServers(0)
	rt := newServerRuntime(svr, r.tw, r.refreshQPS(svr.MaxQPS))
	newValues[svr.ID] = rt
	r.addAnalysis(rt.meta.ID, rt.meta.CircuitBreaker)
	r.addToCheck(rt)

	r.servers = newValues
	log.Infof("server <%d> added, data <%s>",
		svr.ID,
		svr.String())

	return nil
}

func (r *dispatcher) updateServer(meta *metapb.Server) error {
	rt, ok := r.servers[meta.ID]
	if !ok {
		return errServerNotFound
	}

	// stop old heath check
	rt.heathTimeout.Stop()

	newValues := r.copyServers(0)
	rt = newValues[meta.ID]
	rt.activeQPS = r.refreshQPS(meta.MaxQPS)
	rt.updateMeta(meta)
	r.addAnalysis(rt.meta.ID, rt.meta.CircuitBreaker)
	r.addToCheck(rt)

	r.servers = newValues
	log.Infof("server <%d> updated, data <%s>",
		meta.ID,
		meta.String())

	return nil
}

func (r *dispatcher) removeServer(id uint64) error {
	rt, ok := r.servers[id]
	if !ok {
		return errServerNotFound
	}

	// stop old heath check
	rt.heathTimeout.Stop()

	newValues := r.copyServers(id)
	newBinds := r.copyBinds(metapb.Bind{
		ServerID: id,
	})

	r.servers = newValues
	r.binds = newBinds
	log.Infof("server <%d> removed",
		rt.meta.ID)
	return nil
}

func (r *dispatcher) addAnalysis(id uint64, cb *metapb.CircuitBreaker) {
	r.analysiser.RemoveTarget(id)
	r.analysiser.AddTarget(id, time.Second)
	if cb != nil {
		r.analysiser.AddTarget(id, time.Duration(cb.RateCheckPeriod))
	}
}

func (r *dispatcher) addCluster(cluster *metapb.Cluster) error {
	if _, ok := r.clusters[cluster.ID]; ok {
		return errClusterExists
	}

	newValues := r.copyClusters(0)
	newValues[cluster.ID] = newClusterRuntime(cluster)

	r.clusters = newValues
	log.Infof("cluster <%d> added, data <%s>",
		cluster.ID,
		cluster.String())

	return nil
}

func (r *dispatcher) updateCluster(meta *metapb.Cluster) error {
	_, ok := r.clusters[meta.ID]
	if !ok {
		return errClusterNotFound
	}

	newValues := r.copyClusters(0)
	rt := newValues[meta.ID]
	rt.updateMeta(meta)

	r.clusters = newValues
	log.Infof("cluster <%d> updated, data <%s>",
		meta.ID,
		meta.String())

	return nil
}

func (r *dispatcher) removeCluster(id uint64) error {
	_, ok := r.clusters[id]
	if !ok {
		return errClusterNotFound
	}

	newValues := r.copyClusters(id)
	newBinds := r.copyBinds(metapb.Bind{
		ClusterID: id,
	})
	r.binds = newBinds
	r.clusters = newValues

	log.Infof("cluster <%d> removed",
		id)
	return nil
}

func (r *dispatcher) addBind(bind *metapb.Bind) error {
	server, ok := r.servers[bind.ServerID]
	if !ok {
		log.Warnf("bind failed, server <%d> not found",
			bind.ServerID)
		return errServerNotFound
	}

	if _, ok := r.clusters[bind.ClusterID]; !ok {
		log.Warnf("add bind failed, cluster <%d> not found",
			bind.ClusterID)
		return errClusterNotFound
	}

	status := metapb.Unknown
	if server.meta.HeathCheck == nil {
		status = metapb.Up
	}

	newValues := r.copyBinds(metapb.Bind{})
	if _, ok := newValues[bind.ClusterID]; !ok {
		newValues[bind.ClusterID] = &binds{}
	}

	bindInfos := newValues[bind.ClusterID]
	bindInfos.servers = append(bindInfos.servers, &bindInfo{
		svrID:  bind.ServerID,
		status: status,
	})
	if status == metapb.Up {
		bindInfos.actives = append(bindInfos.actives, *server.meta)
	}

	newValues[bind.ClusterID] = bindInfos
	r.binds = newValues

	log.Infof("bind <%d,%d> created", bind.ClusterID, bind.ServerID)
	return nil
}

func (r *dispatcher) removeBind(bind *metapb.Bind) error {
	if _, ok := r.servers[bind.ServerID]; !ok {
		log.Errorf("remove bind failed: server <%d> not found",
			bind.ServerID)
		return errServerNotFound
	}

	if _, ok := r.clusters[bind.ClusterID]; !ok {
		log.Errorf("remove bind failed: cluster <%d> not found",
			bind.ClusterID)
		return errClusterNotFound
	}

	newValues := r.copyBinds(*bind)
	r.binds = newValues
	log.Infof("bind <%d,%d> removed", bind.ClusterID, bind.ServerID)
	return nil
}

func (r *dispatcher) getServerStatus(id uint64) metapb.Status {
	binds := r.binds
	for _, bindInfos := range binds {
		for _, info := range bindInfos.servers {
			if info.svrID == id {
				return info.status
			}
		}
	}

	return metapb.Unknown
}

func sortAPIs(apis map[uint64]*apiRuntime) []uint64 {
	if len(apis) == 0 {
		return nil
	}

	type kv struct {
		Key   uint64
		Value uint32
	}

	ss := make([]kv, len(apis))

	var i = 0
	for k, v := range apis {
		ss[i] = kv{k, v.position()}
		i++
	}

	// position升序
	sort.SliceStable(ss, func(i, j int) bool {
		return ss[i].Value < ss[j].Value
	})

	keys := make([]uint64, len(ss))
	for i, v := range ss {
		keys[i] = v.Key
	}

	return keys
}
