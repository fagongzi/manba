package proxy

import (
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/route"
)

func (r *dispatcher) copyServers(exclude uint64) map[uint64]*serverRuntime {
	values := make(map[uint64]*serverRuntime)
	for key, value := range r.servers {
		if key != exclude {
			values[key] = value.clone()
		}
	}
	return values
}

func (r *dispatcher) copyClusters(exclude uint64) map[uint64]*clusterRuntime {
	values := make(map[uint64]*clusterRuntime)
	for key, value := range r.clusters {
		if key != exclude {
			values[key] = value.clone()
		}

	}
	return values
}

func (r *dispatcher) copyRoutings(exclude uint64) map[uint64]*routingRuntime {
	values := make(map[uint64]*routingRuntime)
	for key, value := range r.routings {
		if key != exclude {
			values[key] = value.clone()
		}
	}
	return values
}

func (r *dispatcher) copyAPIs(exclude uint64) (*route.Route, map[uint64]*apiRuntime) {
	route := route.NewRoute()
	values := make(map[uint64]*apiRuntime)
	for key, value := range r.apis {
		if key != exclude {
			values[key] = value.clone()
			route.Add(values[key].meta)
		}
	}

	return route, values
}

func (r *dispatcher) copyBinds(exclude metapb.Bind) map[uint64]*binds {
	values := make(map[uint64]*binds)
	for key, bindsInfo := range r.binds {
		if exclude.ClusterID != key {
			newBindsInfo := &binds{}
			for _, info := range bindsInfo.servers {
				if info.svrID != exclude.ServerID {
					newBindsInfo.servers = append(newBindsInfo.servers, &bindInfo{
						svrID:  info.svrID,
						status: info.status,
					})
				}
			}

			for _, info := range bindsInfo.actives {
				if info.ID != exclude.ServerID {
					newBindsInfo.actives = append(newBindsInfo.actives, info)
				}
			}

			values[key] = newBindsInfo
		}
	}

	return values
}
