package proxy

import (
	"math"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/store"
	"github.com/fagongzi/log"
	"github.com/fagongzi/util/format"
)

var (
	eventTypeStatusChanged = store.EvtType(math.MaxInt32)
	eventSrcStatusChanged  = store.EvtSrc(math.MaxInt32)
)

type statusChanged struct {
	meta   metapb.Server
	status metapb.Status
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
		} else if evt.Src == eventSrcStatusChanged {
			r.doStatusChangedEvent(evt)
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

func (r *dispatcher) doStatusChangedEvent(evt *store.Evt) {
	value := evt.Value.(statusChanged)
	oldStatus := r.getServerStatus(value.meta.ID)

	if oldStatus == value.status {
		return
	}

	newValues := r.copyBinds(metapb.Bind{})
	for _, binds := range newValues {
		hasServer := false
		for _, bind := range binds.servers {
			if bind.svrID == value.meta.ID {
				hasServer = true
				bind.status = value.status
			}
		}

		if hasServer {
			newActives := make([]metapb.Server, len(binds.actives))
			for _, active := range binds.actives {
				if active.ID != value.meta.ID || value.status == metapb.Up {
					newActives = append(newActives, active)
				}
			}

			binds.actives = newActives
		}
	}

	r.binds = newValues
	log.Infof("server <%d> changed to %s", value.meta.ID, value.status.String())
}
