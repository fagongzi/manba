package proxy

import (
	"errors"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/fagongzi/gateway/pkg/store"
	"github.com/fagongzi/log"
	"github.com/fagongzi/util/json"
)

var (
	errServerExists    = errors.New("Server already exist")
	errClusterExists   = errors.New("Cluster already exist")
	errBindExists      = errors.New("Bind already exist")
	errAPIExists       = errors.New("API already exist")
	errRoutingExists   = errors.New("Routing already exist")
	errServerNotFound  = errors.New("Server not found")
	errClusterNotFound = errors.New("Cluster not found")
	errBindNotFound    = errors.New("Bind not found")
	errAPINotFound     = errors.New("API not found")
	errRoutingNotFound = errors.New("Routing not found")
)

func (r *dispatcher) load() {
	r.loadClusters()
	r.loadServers()
	r.loadBinds()
	r.loadAPIs()
	r.loadRoutings()

	go r.watch()
}

func (r *dispatcher) loadClusters() {
	clusters, err := r.store.GetClusters()
	if nil != err {
		log.Errorf("load clusters failed, errors:\n%+v",
			err)
		return
	}

	log.Infof("load %d clusters", len(clusters))
	for _, cluster := range clusters {
		if err := r.addCluster(cluster); nil != err {
			log.Fatalf("cluster <%s> add failed, errors:\n%+v",
				json.MustMarshal(cluster),
				err)
		}
	}
}

func (r *dispatcher) loadServers() {
	servers, err := r.store.GetServers()
	if nil != err {
		log.Errorf("load servers from store failed, errors:\n%+v",
			err)
		return
	}

	log.Infof("load %d servers", len(servers))
	for _, server := range servers {
		if err := r.addServer(server); nil != err {
			log.Fatalf("server <%s> add failed, errors:\n%+v",
				json.MustMarshal(server),
				err)
		}
	}
}

func (r *dispatcher) loadRoutings() {
	routings, err := r.store.GetRoutings()
	if nil != err {
		log.Errorf("load routings from store failed, errors:\n%+v",
			err)
		return
	}

	log.Infof("load %d routings", len(routings))
	for _, route := range routings {
		if err := r.addRouting(route); nil != err {
			log.Fatalf("routing <%s> add failed, errors:\n%+v",
				json.MustMarshal(route),
				err)
		}
	}
}

func (r *dispatcher) loadBinds() {
	binds, err := r.store.GetBinds()
	if nil != err {
		log.Errorf("load binds from store failed, errors:\n%+v",
			err)
		return
	}

	log.Infof("load %d binds", len(binds))
	for _, b := range binds {
		err := r.addBind(b)
		if nil != err {
			log.Fatalf("bind <%s> add failed, errors:\n%+v",
				json.MustMarshal(b),
				err)
		}
	}
}

func (r *dispatcher) loadAPIs() {
	apis, err := r.store.GetAPIs()
	if nil != err {
		log.Errorf("load apis from store failed, errors:\n%+v",
			err)
		return
	}

	log.Infof("load %d apis", len(apis))
	for _, api := range apis {
		if err := r.addAPI(api); nil != err {
			log.Fatalf("api <%s> add failed, errors:\n%+v",
				json.MustMarshal(api),
				err)
		}
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
		} else {
			log.Warnf("unknown event <%+v>", evt)
		}
	}
}

func (r *dispatcher) doRoutingEvent(evt *store.Evt) {
	routing, _ := evt.Value.(*model.Routing)

	if evt.Type == store.EventTypeNew {
		r.addRouting(routing)
	} else if evt.Type == store.EventTypeDelete {
		r.removeRouting(evt.Key)
	} else if evt.Type == store.EventTypeUpdate {
		// TODO: impl
	}
}

func (r *dispatcher) doAPIEvent(evt *store.Evt) {
	api, _ := evt.Value.(*model.API)

	if evt.Type == store.EventTypeNew {
		r.addAPI(api)
	} else if evt.Type == store.EventTypeDelete {
		r.removeAPI(evt.Key)
	} else if evt.Type == store.EventTypeUpdate {
		r.updateAPI(api)
	}
}

func (r *dispatcher) doClusterEvent(evt *store.Evt) {
	cluster, _ := evt.Value.(*model.Cluster)

	if evt.Type == store.EventTypeNew {
		r.addCluster(cluster)
	} else if evt.Type == store.EventTypeDelete {
		r.removeCluster(evt.Key)
	} else if evt.Type == store.EventTypeUpdate {
		r.updateCluster(cluster)
	}
}

func (r *dispatcher) doServerEvent(evt *store.Evt) {
	svr, _ := evt.Value.(*model.Server)

	if evt.Type == store.EventTypeNew {
		r.addServer(svr)
	} else if evt.Type == store.EventTypeDelete {
		r.removeServer(evt.Key)
	} else if evt.Type == store.EventTypeUpdate {
		r.updateServer(svr)
	}
}

func (r *dispatcher) doBindEvent(evt *store.Evt) {
	bind, _ := evt.Value.(*model.Bind)

	if evt.Type == store.EventTypeNew {
		r.addBind(bind)
	} else if evt.Type == store.EventTypeDelete {
		r.removeBind(evt.Key)
	}
}

func (r *dispatcher) addRouting(routing *model.Routing) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.routings[routing.ID]; ok {
		return errRoutingExists
	}

	r.routings[routing.ID] = routing
	log.Infof("routing <%s> added, data <%s>",
		routing.ID,
		json.MustMarshal(routing))

	return nil
}

func (r *dispatcher) removeRouting(id string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.routings[id]; !ok {
		return errRoutingNotFound
	}

	delete(r.routings, id)
	log.Infof("routing <%s> deleted",
		id)

	return nil
}

func (r *dispatcher) addAPI(api *model.API) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.apis[api.URL]; ok {
		return errAPIExists
	}

	r.apis[api.ID] = api
	log.Infof("api <%s> added, data <%s>",
		api.ID,
		json.MustMarshal(api))

	return nil
}

func (r *dispatcher) updateAPI(api *model.API) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.apis[api.ID]; !ok {
		return errAPINotFound
	}

	r.apis[api.ID] = api
	log.Infof("api <%s> updated, data <%s>",
		api.ID,
		json.MustMarshal(api))

	return nil
}

func (r *dispatcher) removeAPI(id string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.apis[id]; !ok {
		return errAPINotFound
	}

	delete(r.apis, id)
	log.Infof("api <%s> removed",
		id)

	return nil
}

func (r *dispatcher) addServer(svr *model.Server) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.servers[svr.ID]; ok {
		return errServerExists
	}

	rt := newServerRuntime(svr, r.tw)

	binded := make(map[string]*clusterRuntime)
	r.servers[svr.ID] = rt
	r.mapping[svr.ID] = binded

	r.addAnalysis(rt)
	r.addToCheck(rt)

	log.Infof("server <%s> added, data <%s>",
		svr.ID,
		json.MustMarshal(svr))

	return nil
}

func (r *dispatcher) updateServer(meta *model.Server) error {
	r.Lock()
	defer r.Unlock()

	rt, ok := r.servers[meta.ID]
	if !ok {
		return errServerNotFound
	}

	rt.updateMeta(meta, func() {
		r.addAnalysis(rt)
		r.addToCheck(rt)
	})

	log.Infof("server <%s> updated, data <%s>",
		meta.ID,
		json.MustMarshal(meta))

	return nil
}

func (r *dispatcher) removeServer(id string) error {
	r.Lock()
	defer r.Unlock()

	svr, ok := r.servers[id]
	if !ok {
		return errServerNotFound
	}

	delete(r.servers, id)

	binded, _ := r.mapping[svr.meta.ID]
	delete(r.mapping, svr.meta.ID)
	log.Infof("all bind of server <%s> removed ",
		svr.meta.ID)

	for _, cluster := range binded {
		cluster.remove(id)
	}
	log.Infof("server <%s> removed",
		svr.meta.ID)

	return nil
}

func (r *dispatcher) addCluster(cluster *model.Cluster) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.clusters[cluster.ID]; ok {
		return errClusterExists
	}

	r.clusters[cluster.ID] = newClusterRuntime(cluster)
	log.Infof("cluster <%s> added, data <%s>",
		cluster.ID,
		json.MustMarshal(cluster))

	return nil
}

func (r *dispatcher) updateCluster(meta *model.Cluster) error {
	r.Lock()
	defer r.Unlock()

	rt, ok := r.clusters[meta.ID]
	if !ok {
		return errClusterNotFound
	}

	rt.updateMeta(meta)
	log.Infof("cluster <%s> updated, data <%s>",
		meta.ID,
		json.MustMarshal(meta))

	return nil
}

func (r *dispatcher) removeCluster(id string) error {
	r.Lock()
	defer r.Unlock()

	cluster, ok := r.clusters[id]
	if !ok {
		return errClusterNotFound
	}

	// TODO: check API node loose cluster
	cluster.foreach(func(id string) {
		if svr, ok := r.servers[id]; ok {
			r.doUnBind(svr, cluster, false)
		}
	})

	delete(r.clusters, cluster.meta.ID)
	log.Infof("cluster <%s> removed",
		cluster.meta.ID)

	return nil
}

func (r *dispatcher) addBind(bind *model.Bind) error {
	r.Lock()
	defer r.Unlock()

	svr, ok := r.servers[bind.ServerID]
	if !ok {
		log.Warnf("bind failed, server <%s> not found",
			bind.ServerID)
		return errServerNotFound
	}

	cluster, ok := r.clusters[bind.ClusterID]
	if !ok {
		log.Warnf("add bind failed, cluster <%s> not found",
			bind.ClusterID)
		return errClusterNotFound
	}

	binded, _ := r.mapping[svr.meta.ID]
	if c, ok := binded[cluster.meta.ID]; ok &&
		c.meta.ID == bind.ClusterID {
		return errBindExists
	}

	r.binds[bind.ID] = bind
	binded[cluster.meta.ID] = cluster

	cluster.add(svr.meta.ID)

	log.Infof("bind <%s,%s> stored",
		bind.ClusterID,
		bind.ServerID)

	return nil
}

func (r *dispatcher) removeBind(id string) error {
	r.Lock()
	defer r.Unlock()

	bind, ok := r.binds[id]
	if !ok {
		return errBindNotFound
	}

	svr, ok := r.servers[bind.ServerID]
	if !ok {
		log.Errorf("remove bind failed: server <%s> not found",
			bind.ServerID)
		return errServerNotFound
	}

	cluster, ok := r.clusters[bind.ClusterID]
	if !ok {
		log.Errorf("remove bind failed: cluster <%s> not found",
			bind.ClusterID)
		return errClusterNotFound
	}

	delete(r.binds, id)
	r.doUnBind(svr, cluster, true)

	log.Infof("bind <%s,%s> removed",
		bind.ClusterID,
		bind.ServerID)

	return nil
}
