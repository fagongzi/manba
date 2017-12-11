package store

import (
	"fmt"
	"strings"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/log"
	"github.com/fagongzi/util/format"
	"github.com/fagongzi/util/protoc"
)

// Watch watch event from etcd
func (e *EtcdStore) Watch(evtCh chan *Evt, stopCh chan bool) error {
	e.evtCh = evtCh

	log.Infof("watch event at: <%s>",
		e.prefix)

	e.doWatch()

	return nil
}

func (e EtcdStore) doWatch() {
	watcher := clientv3.NewWatcher(e.rawClient)
	defer watcher.Close()

	ctx := e.rawClient.Ctx()
	for {
		rch := watcher.Watch(ctx, e.prefix, clientv3.WithPrefix())
		for wresp := range rch {
			if wresp.Canceled {
				return
			}

			for _, ev := range wresp.Events {
				var evtSrc EvtSrc
				var evtType EvtType

				switch ev.Type {
				case mvccpb.DELETE:
					evtType = EventTypeDelete
				case mvccpb.PUT:
					if ev.IsCreate() {
						evtType = EventTypeNew
					} else if ev.IsModify() {
						evtType = EventTypeUpdate
					}
				}

				key := string(ev.Kv.Key)
				if strings.HasPrefix(key, e.clustersDir) {
					evtSrc = EventSrcCluster
				} else if strings.HasPrefix(key, e.serversDir) {
					evtSrc = EventSrcServer
				} else if strings.HasPrefix(key, e.bindsDir) {
					evtSrc = EventSrcBind
				} else if strings.HasPrefix(key, e.apisDir) {
					evtSrc = EventSrcAPI
				} else if strings.HasPrefix(key, e.routingsDir) {
					evtSrc = EventSrcRouting
				} else {
					continue
				}

				log.Debugf("watch event: <%s, %v>",
					key,
					evtType)
				e.evtCh <- e.watchMethodMapping[evtSrc](evtType, ev.Kv)
			}
		}

		select {
		case <-ctx.Done():
			// server closed, return
			return
		default:
		}
	}
}

func (e *EtcdStore) doWatchWithCluster(evtType EvtType, kv *mvccpb.KeyValue) *Evt {
	value := &metapb.Cluster{}
	if len(kv.Value) > 0 {
		protoc.MustUnmarshal(value, []byte(kv.Value))
	}

	return &Evt{
		Src:   EventSrcCluster,
		Type:  evtType,
		Key:   strings.Replace(string(kv.Key), fmt.Sprintf("%s/", e.clustersDir), "", 1),
		Value: value,
	}
}

func (e *EtcdStore) doWatchWithServer(evtType EvtType, kv *mvccpb.KeyValue) *Evt {
	value := &metapb.Server{}
	if len(kv.Value) > 0 {
		protoc.MustUnmarshal(value, []byte(kv.Value))
	}

	return &Evt{
		Src:   EventSrcServer,
		Type:  evtType,
		Key:   strings.Replace(string(kv.Key), fmt.Sprintf("%s/", e.serversDir), "", 1),
		Value: value,
	}
}

func (e *EtcdStore) doWatchWithBind(evtType EvtType, kv *mvccpb.KeyValue) *Evt {
	// bind key is: bindsDir/clusterID/serverID
	key := strings.Replace(string(kv.Key), fmt.Sprintf("%s/", e.bindsDir), "", 1)
	infos := strings.SplitN(key, "/", 2)

	return &Evt{
		Src:  EventSrcBind,
		Type: evtType,
		Key:  string(kv.Key),
		Value: &metapb.Bind{
			ClusterID: format.MustParseStrUInt64(infos[0]),
			ServerID:  format.MustParseStrUInt64(infos[1]),
		},
	}
}

func (e *EtcdStore) doWatchWithAPI(evtType EvtType, kv *mvccpb.KeyValue) *Evt {
	value := &metapb.API{}
	if len(kv.Value) > 0 {
		protoc.MustUnmarshal(value, []byte(kv.Value))
	}

	return &Evt{
		Src:   EventSrcAPI,
		Type:  evtType,
		Key:   strings.Replace(string(kv.Key), fmt.Sprintf("%s/", e.apisDir), "", 1),
		Value: value,
	}
}

func (e *EtcdStore) doWatchWithRouting(evtType EvtType, kv *mvccpb.KeyValue) *Evt {
	value := &metapb.Routing{}
	if len(kv.Value) > 0 {
		protoc.MustUnmarshal(value, []byte(kv.Value))
	}

	return &Evt{
		Src:   EventSrcRouting,
		Type:  evtType,
		Key:   strings.Replace(string(kv.Key), fmt.Sprintf("%s/", e.routingsDir), "", 1),
		Value: value,
	}
}

func (e *EtcdStore) init() {
	e.watchMethodMapping[EventSrcBind] = e.doWatchWithBind
	e.watchMethodMapping[EventSrcServer] = e.doWatchWithServer
	e.watchMethodMapping[EventSrcCluster] = e.doWatchWithCluster
	e.watchMethodMapping[EventSrcAPI] = e.doWatchWithAPI
	e.watchMethodMapping[EventSrcRouting] = e.doWatchWithRouting
}
