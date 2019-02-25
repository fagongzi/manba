package proxy

import (
	"errors"
	"time"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/store"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/fagongzi/log"
	"github.com/fagongzi/util/format"
)

var (
	errServerExists    = errors.New("Server already exist")
	errClusterExists   = errors.New("Cluster already exist")
	errBindExists      = errors.New("Bind already exist")
	errAPIExists       = errors.New("API already exist")
	errProxyExists     = errors.New("Proxy already exist")
	errPluginExists    = errors.New("Plugin already exist")
	errRoutingExists   = errors.New("Routing already exist")
	errServerNotFound  = errors.New("Server not found")
	errClusterNotFound = errors.New("Cluster not found")
	errBindNotFound    = errors.New("Bind not found")
	errProxyNotFound   = errors.New("Proxy not found")
	errAPINotFound     = errors.New("API not found")
	errRoutingNotFound = errors.New("Routing not found")
	errPluginNotFound  = errors.New("Plugin not found")

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
	r.loadPlugins()
	r.loadAppliedPlugins()
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

func (r *dispatcher) loadPlugins() {
	log.Infof("load plugins")

	err := r.store.GetPlugins(limit, func(value interface{}) error {
		return r.addPlugin(value.(*metapb.Plugin))
	})
	if nil != err {
		log.Errorf("load plugins failed, errors:\n%+v",
			err)
		return
	}
}

func (r *dispatcher) loadAppliedPlugins() {
	log.Infof("load applied plugins")

	applied, err := r.store.GetAppliedPlugins()
	if nil != err {
		log.Errorf("load applied plugins failed, errors:\n%+v",
			err)
		return
	}

	err = r.updateAppliedPlugin(applied)
	if nil != err {
		log.Errorf("updated applied plugins failed, errors:\n%+v",
			err)
		return
	}
}

func (r *dispatcher) watch() {
	log.Info("router start watch meta data")

	go r.readyToReceiveWatchEvent()
	err := r.store.Watch(r.watchEventC, r.watchStopC)
	log.Errorf("router watch failed, errors:\n%+v",
		err)
}

func (r *dispatcher) readyToReceiveWatchEvent() {
	for {
		evt := <-r.watchEventC

		if evt.Src == store.EventSrcCluster {
			r.doClusterEvent(evt)
		} else if evt.Src == store.EventSrcServer {
			r.doServerEvent(evt)
		} else if evt.Src == store.EventSrcBind {
			r.doBindEvent(evt)
		} else if evt.Src == store.EventSrcAPI {
			r.doAPIEvent(evt)
		} else if evt.Src == store.EventSrcRouting {
			r.doRoutingEvent(evt)
		} else if evt.Src == store.EventSrcProxy {
			r.doProxyEvent(evt)
		} else if evt.Src == store.EventSrcPlugin {
			r.doPluginEvent(evt)
		} else if evt.Src == store.EventSrcApplyPlugin {
			r.doApplyPluginEvent(evt)
		} else {
			log.Warnf("unknown event <%+v>", evt)
		}
	}
}

func (r *dispatcher) doRoutingEvent(evt *store.Evt) {
	routing, _ := evt.Value.(*metapb.Routing)

	if evt.Type == store.EventTypeNew {
		r.addRouting(routing)
	} else if evt.Type == store.EventTypeDelete {
		r.removeRouting(format.MustParseStrUInt64(evt.Key))
	} else if evt.Type == store.EventTypeUpdate {
		r.updateRouting(routing)
	}
}

func (r *dispatcher) doProxyEvent(evt *store.Evt) {
	proxy, _ := evt.Value.(*metapb.Proxy)

	if evt.Type == store.EventTypeNew {
		r.addProxy(proxy)
	} else if evt.Type == store.EventTypeDelete {
		r.removeProxy(evt.Key)
	}
}

func (r *dispatcher) doAPIEvent(evt *store.Evt) {
	api, _ := evt.Value.(*metapb.API)

	if evt.Type == store.EventTypeNew {
		r.addAPI(api)
	} else if evt.Type == store.EventTypeDelete {
		r.removeAPI(format.MustParseStrUInt64(evt.Key))
	} else if evt.Type == store.EventTypeUpdate {
		r.updateAPI(api)
	}
}

func (r *dispatcher) doClusterEvent(evt *store.Evt) {
	cluster, _ := evt.Value.(*metapb.Cluster)

	if evt.Type == store.EventTypeNew {
		r.addCluster(cluster)
	} else if evt.Type == store.EventTypeDelete {
		r.removeCluster(format.MustParseStrUInt64(evt.Key))
	} else if evt.Type == store.EventTypeUpdate {
		r.updateCluster(cluster)
	}
}

func (r *dispatcher) doServerEvent(evt *store.Evt) {
	svr, _ := evt.Value.(*metapb.Server)

	if evt.Type == store.EventTypeNew {
		r.addServer(svr)
	} else if evt.Type == store.EventTypeDelete {
		r.removeServer(format.MustParseStrUInt64(evt.Key))
	} else if evt.Type == store.EventTypeUpdate {
		r.updateServer(svr)
	}
}

func (r *dispatcher) doBindEvent(evt *store.Evt) {
	bind, _ := evt.Value.(*metapb.Bind)

	if evt.Type == store.EventTypeNew {
		r.addBind(bind)
	} else if evt.Type == store.EventTypeDelete {
		r.removeBind(bind)
	}
}

func (r *dispatcher) doPluginEvent(evt *store.Evt) {
	value, _ := evt.Value.(*metapb.Plugin)

	if evt.Type == store.EventTypeNew {
		r.addPlugin(value)
	} else if evt.Type == store.EventTypeDelete {
		r.removePlugin(format.MustParseStrUInt64(evt.Key))
	} else if evt.Type == store.EventTypeUpdate {
		r.updatePlugin(value)
	}
}

func (r *dispatcher) doApplyPluginEvent(evt *store.Evt) {
	value, _ := evt.Value.(*metapb.AppliedPlugins)

	if evt.Type == store.EventTypeNew {
		r.updateAppliedPlugin(value)
	} else if evt.Type == store.EventTypeDelete {
		r.removeAppliedPlugin()
	} else if evt.Type == store.EventTypeUpdate {
		r.updateAppliedPlugin(value)
	}
}

func (r *dispatcher) addRouting(meta *metapb.Routing) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.routings[meta.ID]; ok {
		return errRoutingExists
	}

	r.routings[meta.ID] = newRoutingRuntime(meta)
	log.Infof("routing <%d> added, data <%s>",
		meta.ID,
		meta.String())

	return nil
}

func (r *dispatcher) updateRouting(meta *metapb.Routing) error {
	r.Lock()
	defer r.Unlock()

	rt, ok := r.routings[meta.ID]
	if !ok {
		return errRoutingNotFound
	}

	rt.updateMeta(meta)
	log.Infof("routing <%d> updated, data <%s>",
		meta.ID,
		meta.String())

	return nil
}

func (r *dispatcher) removeRouting(id uint64) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.routings[id]; !ok {
		return errRoutingNotFound
	}

	delete(r.routings, id)
	log.Infof("routing <%d> deleted",
		id)

	return nil
}

func (r *dispatcher) addProxy(meta *metapb.Proxy) error {
	r.Lock()
	defer r.Unlock()

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
	r.Lock()
	defer r.Unlock()

	if _, ok := r.proxies[addr]; !ok {
		return errProxyNotFound
	}

	delete(r.proxies, addr)
	r.refreshAllQPS()

	log.Infof("proxy <%s> deleted", addr)
	return nil
}

func (r *dispatcher) addAPI(api *metapb.API) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.apis[api.ID]; ok {
		return errAPIExists
	}

	err := r.route.Add(api)
	if err != nil {
		return err
	}

	a := newAPIRuntime(api, r.tw, r.refreshQPS(api.MaxQPS))
	r.apis[api.ID] = a

	if a.cb != nil {
		r.addAnalysis(api.ID, a.cb)
	}

	log.Infof("api <%d> added, data <%s>",
		api.ID,
		api.String())

	return nil
}

func (r *dispatcher) updateAPI(api *metapb.API) error {
	r.Lock()
	defer r.Unlock()

	rt, ok := r.apis[api.ID]
	if !ok {
		return errAPINotFound
	}

	err := r.route.Update(api)
	if err != nil {
		return err
	}

	rt.activeQPS = r.refreshQPS(api.MaxQPS)
	rt.updateMeta(api)

	if rt.cb != nil {
		r.addAnalysis(rt.meta.ID, rt.meta.CircuitBreaker)
	}

	log.Infof("api <%d> updated, data <%s>",
		api.ID,
		api.String())

	return nil
}

func (r *dispatcher) removeAPI(id uint64) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.apis[id]; !ok {
		return errAPINotFound
	}

	r.route.Remove(id)
	delete(r.apis, id)

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
	if len(r.proxies) > 0 {
		return value / int64(len(r.proxies))
	}
	return value
}

func (r *dispatcher) addServer(svr *metapb.Server) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.servers[svr.ID]; ok {
		return errServerExists
	}

	rt := newServerRuntime(svr, r.tw, r.refreshQPS(svr.MaxQPS))
	r.servers[svr.ID] = rt

	r.addAnalysis(rt.meta.ID, rt.meta.CircuitBreaker)
	r.addToCheck(rt)

	log.Infof("server <%d> added, data <%s>",
		svr.ID,
		svr.String())

	return nil
}

func (r *dispatcher) updateServer(meta *metapb.Server) error {
	r.Lock()
	defer r.Unlock()

	rt, ok := r.servers[meta.ID]
	if !ok {
		return errServerNotFound
	}

	rt.activeQPS = r.refreshQPS(meta.MaxQPS)
	rt.updateMeta(meta)

	r.addAnalysis(rt.meta.ID, rt.meta.CircuitBreaker)
	r.addToCheck(rt)

	log.Infof("server <%d> updated, data <%s>",
		meta.ID,
		meta.String())

	return nil
}

func (r *dispatcher) removeServer(id uint64) error {
	r.Lock()
	defer r.Unlock()

	svr, ok := r.servers[id]
	if !ok {
		return errServerNotFound
	}

	svr.heathTimeout.Stop()
	delete(r.servers, id)
	for _, cluster := range r.clusters {
		cluster.remove(id)
	}

	log.Infof("server <%d> removed",
		svr.meta.ID)
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
	r.Lock()
	defer r.Unlock()

	if _, ok := r.clusters[cluster.ID]; ok {
		return errClusterExists
	}

	r.clusters[cluster.ID] = newClusterRuntime(cluster)
	log.Infof("cluster <%d> added, data <%s>",
		cluster.ID,
		cluster.String())

	return nil
}

func (r *dispatcher) updateCluster(meta *metapb.Cluster) error {
	r.Lock()
	defer r.Unlock()

	rt, ok := r.clusters[meta.ID]
	if !ok {
		return errClusterNotFound
	}

	rt.updateMeta(meta)
	log.Infof("cluster <%d> updated, data <%s>",
		meta.ID,
		meta.String())

	return nil
}

func (r *dispatcher) removeCluster(id uint64) error {
	r.Lock()
	defer r.Unlock()

	cluster, ok := r.clusters[id]
	if !ok {
		return errClusterNotFound
	}

	// TODO: check API node loose cluster
	for _, clusters := range r.binds {
		delete(clusters, id)
	}

	delete(r.clusters, cluster.meta.ID)
	log.Infof("cluster <%d> removed",
		cluster.meta.ID)

	return nil
}

func (r *dispatcher) addBind(bind *metapb.Bind) error {
	r.Lock()
	defer r.Unlock()

	server, ok := r.servers[bind.ServerID]
	if !ok {
		log.Warnf("bind failed, server <%d> not found",
			bind.ServerID)
		return errServerNotFound
	}

	cluster, ok := r.clusters[bind.ClusterID]
	if !ok {
		log.Warnf("add bind failed, cluster <%d> not found",
			bind.ClusterID)
		return errClusterNotFound
	}

	if _, ok := r.binds[bind.ServerID]; !ok {
		r.binds[bind.ServerID] = make(map[uint64]*clusterRuntime)
	}

	clusters := r.binds[bind.ServerID]
	clusters[bind.ClusterID] = cluster

	log.Infof("bind <%d,%d> created", bind.ClusterID, bind.ServerID)

	if server.status == metapb.Up {
		cluster.add(server.meta)
	}
	return nil
}

func (r *dispatcher) removeBind(bind *metapb.Bind) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.servers[bind.ServerID]; !ok {
		log.Errorf("remove bind failed: server <%d> not found",
			bind.ServerID)
		return errServerNotFound
	}

	cluster, ok := r.clusters[bind.ClusterID]
	if !ok {
		log.Errorf("remove bind failed: cluster <%d> not found",
			bind.ClusterID)
		return errClusterNotFound
	}

	cluster.remove(bind.ServerID)

	if clusters, ok := r.binds[bind.ServerID]; ok {
		delete(clusters, bind.ClusterID)
		log.Infof("bind <%d,%d> removed", bind.ClusterID, bind.ServerID)
	}

	return nil
}

func (r *dispatcher) addPlugin(value *metapb.Plugin) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.plugins[value.ID]; ok {
		return errPluginExists
	}

	r.plugins[value.ID] = value

	log.Infof("plugin <%d/%s:%d> added",
		value.ID,
		value.Name,
		value.Version)

	return nil
}

func (r *dispatcher) updatePlugin(value *metapb.Plugin) error {
	r.Lock()
	defer r.Unlock()

	_, ok := r.plugins[value.ID]
	if !ok {
		return errPluginNotFound
	}

	if r.appliedPlugins != nil {
		for _, id := range r.appliedPlugins.AppliedIDs {
			if id == value.ID {
				err := r.jsEngine.UpdatePlugin(value)
				if err != nil {
					return err
				}
			}
		}
	}

	r.plugins[value.ID] = value
	log.Infof("plugin <%d/%s:%d> updated",
		value.ID,
		value.Name,
		value.Version)

	return nil
}

func (r *dispatcher) removePlugin(id uint64) error {
	r.Lock()
	defer r.Unlock()

	value, ok := r.plugins[id]
	if !ok {
		return errPluginNotFound
	}

	delete(r.plugins, id)
	log.Infof("plugin <%d/%s:%d> removed",
		value.ID,
		value.Name,
		value.Version)
	return nil
}

func (r *dispatcher) updateAppliedPlugin(value *metapb.AppliedPlugins) error {
	r.Lock()
	defer r.Unlock()

	var plugins []*metapb.Plugin
	for _, id := range value.AppliedIDs {
		plugin, ok := r.plugins[id]
		if !ok {
			return errPluginNotFound
		}

		plugins = append(plugins, plugin)
	}

	err := r.jsEngine.ApplyPlugins(plugins...)
	if err != nil {
		return err
	}

	r.appliedPlugins = value
	log.Infof("plugins applied with %+v",
		value.AppliedIDs)
	return nil
}

func (r *dispatcher) removeAppliedPlugin() error {
	err := r.jsEngine.ApplyPlugins()
	if err != nil {
		return err
	}

	r.appliedPlugins = &metapb.AppliedPlugins{}
	log.Infof("plugins applied removed")
	return nil
}
