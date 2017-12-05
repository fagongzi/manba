package store

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/fagongzi/log"
	fjson "github.com/fagongzi/util/json"
	"github.com/fagongzi/util/task"
	"golang.org/x/net/context"
)

var (
	// ErrHasBind error has bind into, can not delete
	ErrHasBind = errors.New("Has bind info, can not delete")
)

const (
	// DefaultTimeout default timeout
	DefaultTimeout = time.Second * 3
	// DefaultRequestTimeout default request timeout
	DefaultRequestTimeout = 10 * time.Second
	// DefaultSlowRequestTime default slow request time
	DefaultSlowRequestTime = time.Second * 1
)

// slowLogTxn wraps etcd transaction and log slow one.
type slowLogTxn struct {
	clientv3.Txn
	cancel context.CancelFunc
}

func newSlowLogTxn(client *clientv3.Client) clientv3.Txn {
	ctx, cancel := context.WithTimeout(client.Ctx(), DefaultRequestTimeout)
	return &slowLogTxn{
		Txn:    client.Txn(ctx),
		cancel: cancel,
	}
}

func (t *slowLogTxn) If(cs ...clientv3.Cmp) clientv3.Txn {
	return &slowLogTxn{
		Txn:    t.Txn.If(cs...),
		cancel: t.cancel,
	}
}

func (t *slowLogTxn) Then(ops ...clientv3.Op) clientv3.Txn {
	return &slowLogTxn{
		Txn:    t.Txn.Then(ops...),
		cancel: t.cancel,
	}
}

// Commit implements Txn Commit interface.
func (t *slowLogTxn) Commit() (*clientv3.TxnResponse, error) {
	start := time.Now()
	resp, err := t.Txn.Commit()
	t.cancel()

	cost := time.Now().Sub(start)
	if cost > DefaultSlowRequestTime {
		log.Warn("slow: txn runs too slow, resp=<%v> cost=<%s> errors:\n %+v",
			resp,
			cost,
			err)
	}

	return resp, err
}

// EtcdStore etcd store impl
type EtcdStore struct {
	prefix      string
	clustersDir string
	serversDir  string
	bindsDir    string
	apisDir     string
	proxiesDir  string
	routingsDir string

	cli                *clientv3.Client
	evtCh              chan *Evt
	watchMethodMapping map[EvtSrc]func(EvtType, *mvccpb.KeyValue) *Evt

	taskRunner *task.Runner
}

// NewEtcdStore create a etcd store
func NewEtcdStore(etcdAddrs []string, prefix string, taskRunner *task.Runner) (Store, error) {
	store := &EtcdStore{
		prefix:             prefix,
		clustersDir:        fmt.Sprintf("%s/clusters", prefix),
		serversDir:         fmt.Sprintf("%s/servers", prefix),
		bindsDir:           fmt.Sprintf("%s/binds", prefix),
		apisDir:            fmt.Sprintf("%s/apis", prefix),
		proxiesDir:         fmt.Sprintf("%s/proxy", prefix),
		routingsDir:        fmt.Sprintf("%s/routings", prefix),
		watchMethodMapping: make(map[EvtSrc]func(EvtType, *mvccpb.KeyValue) *Evt),
		taskRunner:         taskRunner,
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   etcdAddrs,
		DialTimeout: DefaultTimeout,
	})

	if err != nil {
		return nil, err
	}

	store.cli = cli

	store.init()
	return store, nil
}

func (e *EtcdStore) txn() clientv3.Txn {
	return newSlowLogTxn(e.cli)
}

// SaveBind save bind to store
func (e *EtcdStore) SaveBind(bind *model.Bind) error {
	svr, err := e.GetServer(bind.ServerID)
	if err != nil {
		return err
	}
	svr.AddBind(bind)

	cluster, err := e.GetCluster(bind.ClusterID)
	if err != nil {
		return err
	}
	cluster.AddBind(bind)

	bindKey := fmt.Sprintf("%s/%s", e.bindsDir, bind.ID)
	svrKey := fmt.Sprintf("%s/%s", e.serversDir, svr.Addr)
	clusterKey := fmt.Sprintf("%s/%s", e.clustersDir, cluster.Name)

	opBind := clientv3.OpPut(bindKey, string(fjson.MustMarshal(bind)))
	opSvr := clientv3.OpPut(svrKey, string(fjson.MustMarshal(svr)))
	opCluster := clientv3.OpPut(clusterKey, string(fjson.MustMarshal(cluster)))

	_, err = e.txn().Then(opBind, opSvr, opCluster).Commit()
	return err
}

// UnBind delete bind from store
func (e *EtcdStore) UnBind(id string) error {
	bind, err := e.GetBind(id)
	if err != nil {
		return err
	}

	svr, err := e.GetServer(bind.ServerID)
	if err != nil {
		return err
	}
	svr.RemoveBind(bind.ClusterID)

	c, err := e.GetCluster(bind.ClusterID)
	if err != nil {
		return err
	}
	c.RemoveBind(bind.ServerID)

	bindKey := fmt.Sprintf("%s/%s", e.bindsDir, bind.ID)
	svrKey := fmt.Sprintf("%s/%s", e.serversDir, svr.Addr)
	clusterKey := fmt.Sprintf("%s/%s", e.clustersDir, c.Name)

	opBind := clientv3.OpDelete(bindKey)
	opSvr := clientv3.OpPut(svrKey, string(fjson.MustMarshal(svr)))
	opCluster := clientv3.OpPut(clusterKey, string(fjson.MustMarshal(c)))
	_, err = e.txn().Then(opBind, opSvr, opCluster).Commit()
	return err
}

// GetBind return bind info
func (e *EtcdStore) GetBind(id string) (*model.Bind, error) {
	key := fmt.Sprintf("%s/%s", e.bindsDir, id)

	var value *model.Bind
	err := e.getList(key, func(item *mvccpb.KeyValue) {
		value = &model.Bind{}
		fjson.MustUnmarshal(value, item.Value)
	})
	if err != nil {
		return nil, err
	}

	if nil == value {
		return nil, fmt.Errorf("bind <%s> not found", id)
	}

	return value, err
}

// GetBinds return binds info
func (e *EtcdStore) GetBinds() ([]*model.Bind, error) {
	var values []*model.Bind
	err := e.getList(e.bindsDir, func(item *mvccpb.KeyValue) {
		value := &model.Bind{}
		fjson.MustUnmarshal(value, item.Value)
		values = append(values, value)
	})

	return values, err
}

// SaveCluster save a cluster to store
func (e *EtcdStore) SaveCluster(cluster *model.Cluster) error {
	return e.doUpdateCluster(cluster)
}

// UpdateCluster update a cluster to store
func (e EtcdStore) UpdateCluster(cluster *model.Cluster) error {
	if _, err := e.GetCluster(cluster.ID); nil != err {
		return err
	}

	return e.doUpdateCluster(cluster)
}

func (e *EtcdStore) doUpdateCluster(cluster *model.Cluster) error {
	key := fmt.Sprintf("%s/%s", e.clustersDir, cluster.ID)
	return e.put(key, string(fjson.MustMarshal(cluster)))
}

// DeleteCluster delete a cluster from store
func (e *EtcdStore) DeleteCluster(id string) error {
	c, err := e.GetCluster(id)
	if err != nil {
		return err
	}

	if c.HasBind() {
		return ErrHasBind
	}

	key := fmt.Sprintf("%s/%s", e.clustersDir, id)
	return e.delete(key)
}

// GetClusters return clusters in store
func (e *EtcdStore) GetClusters() ([]*model.Cluster, error) {
	var values []*model.Cluster
	err := e.getList(e.clustersDir, func(item *mvccpb.KeyValue) {
		c := &model.Cluster{}
		fjson.MustUnmarshal(c, item.Value)
		values = append(values, c)
	})

	return values, err
}

// GetCluster return cluster info
func (e *EtcdStore) GetCluster(id string) (*model.Cluster, error) {
	key := fmt.Sprintf("%s/%s", e.clustersDir, id)

	var value *model.Cluster
	err := e.getList(key, func(item *mvccpb.KeyValue) {
		value = &model.Cluster{}
		fjson.MustUnmarshal(value, item.Value)
	})
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, fmt.Errorf("cluster <%s> not found", id)
	}

	return value, err
}

// SaveServer save a server to store
func (e *EtcdStore) SaveServer(svr *model.Server) error {
	return e.doUpdateServer(svr)
}

// UpdateServer update a server to store
func (e *EtcdStore) UpdateServer(svr *model.Server) error {
	if _, err := e.GetServer(svr.ID); nil != err {
		return err
	}

	return e.doUpdateServer(svr)
}

func (e *EtcdStore) doUpdateServer(svr *model.Server) error {
	key := fmt.Sprintf("%s/%s", e.serversDir, svr.ID)
	return e.put(key, string(fjson.MustMarshal(svr)))
}

// DeleteServer delete a server from store
func (e *EtcdStore) DeleteServer(id string) error {
	svr, err := e.GetServer(id)
	if err != nil {
		return err
	}

	if svr.HasBind() {
		return ErrHasBind
	}

	key := fmt.Sprintf("%s/%s", e.serversDir, id)
	return e.delete(key)
}

// GetServers return server from store
func (e *EtcdStore) GetServers() ([]*model.Server, error) {
	var values []*model.Server
	err := e.getList(e.serversDir, func(item *mvccpb.KeyValue) {
		svr := &model.Server{}
		fjson.MustUnmarshal(svr, item.Value)
		values = append(values, svr)
	})

	return values, err
}

// GetServer return spec server
func (e *EtcdStore) GetServer(id string) (*model.Server, error) {
	key := fmt.Sprintf("%s/%s", e.serversDir, id)

	var value *model.Server
	err := e.getList(key, func(item *mvccpb.KeyValue) {
		value = &model.Server{}
		fjson.MustUnmarshal(value, item.Value)
	})
	if err != nil {
		return nil, err
	}

	if nil == value {
		return nil, fmt.Errorf("server <%s> not found", id)
	}

	return value, err
}

// SaveAPI save a api in store
func (e *EtcdStore) SaveAPI(api *model.API) error {
	return e.UpdateAPI(api)
}

// UpdateAPI update a api in store
func (e *EtcdStore) UpdateAPI(api *model.API) error {
	if _, err := e.GetAPI(api.ID); nil != err {
		return err
	}

	key := fmt.Sprintf("%s/%s", e.apisDir, api.ID)
	return e.put(key, string(fjson.MustMarshal(api)))
}

// DeleteAPI delete a api from store
func (e *EtcdStore) DeleteAPI(id string) error {
	key := fmt.Sprintf("%s/%s", e.apisDir, id)
	return e.delete(key)
}

// GetAPIs return api list from store
func (e *EtcdStore) GetAPIs() ([]*model.API, error) {
	var values []*model.API
	err := e.getList(e.apisDir, func(item *mvccpb.KeyValue) {
		value := &model.API{}
		fjson.MustUnmarshal(value, item.Value)
		values = append(values, value)
	})

	return values, err
}

// GetAPI return api by url from store
func (e *EtcdStore) GetAPI(id string) (*model.API, error) {
	key := fmt.Sprintf("%s/%s", e.apisDir, id)

	var value *model.API
	err := e.getList(key, func(item *mvccpb.KeyValue) {
		value = &model.API{}
		fjson.MustUnmarshal(value, item.Value)
	})
	if err != nil {
		return nil, err
	}

	if value == nil {
		return nil, fmt.Errorf("api <%s> not found", id)
	}

	return value, err
}

// SaveRouting save route to store
func (e *EtcdStore) SaveRouting(routing *model.Routing) error {
	key := fmt.Sprintf("%s/%s", e.routingsDir, routing.ID)
	return e.put(key, string(fjson.MustMarshal(routing)))
}

// GetRoutings return routes in store
func (e *EtcdStore) GetRoutings() ([]*model.Routing, error) {
	var values []*model.Routing
	err := e.getList(e.routingsDir, func(item *mvccpb.KeyValue) {
		value := &model.Routing{}
		fjson.MustUnmarshal(value, item.Value)
		values = append(values, value)
	})

	return values, err
}

// Clean clean data in store
func (e *EtcdStore) Clean() error {
	_, err := e.txn().Then(clientv3.OpDelete(e.prefix, clientv3.WithPrefix())).Commit()
	return err
}

// Watch watch event from etcd
func (e *EtcdStore) Watch(evtCh chan *Evt, stopCh chan bool) error {
	e.evtCh = evtCh

	log.Infof("meta: etcd watch at: <%s>",
		e.prefix)

	e.doWatch()

	return nil
}

func (e EtcdStore) doWatch() {
	watcher := clientv3.NewWatcher(e.cli)
	defer watcher.Close()

	ctx := e.cli.Ctx()
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

				log.Infof("meta: etcd changed: <%s, %v>",
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
	cluster := &model.Cluster{}
	fjson.MustUnmarshal(cluster, []byte(kv.Value))

	return &Evt{
		Src:   EventSrcCluster,
		Type:  evtType,
		Key:   strings.Replace(string(kv.Key), fmt.Sprintf("%s/", e.clustersDir), "", 1),
		Value: cluster,
	}
}

func (e *EtcdStore) doWatchWithServer(evtType EvtType, kv *mvccpb.KeyValue) *Evt {
	svr := &model.Server{}
	fjson.MustUnmarshal(svr, kv.Value)

	return &Evt{
		Src:   EventSrcServer,
		Type:  evtType,
		Key:   strings.Replace(string(kv.Key), fmt.Sprintf("%s/", e.serversDir), "", 1),
		Value: svr,
	}
}

func (e *EtcdStore) doWatchWithBind(evtType EvtType, kv *mvccpb.KeyValue) *Evt {
	key := strings.Replace(string(kv.Key), fmt.Sprintf("%s/", e.bindsDir), "", 1)
	infos := strings.SplitN(key, "-", 2)

	return &Evt{
		Src:  EventSrcBind,
		Type: evtType,
		Key:  string(kv.Key),
		Value: &model.Bind{
			ServerID:  infos[0],
			ClusterID: infos[1],
		},
	}
}

func (e *EtcdStore) doWatchWithAPI(evtType EvtType, kv *mvccpb.KeyValue) *Evt {
	api := &model.API{}
	fjson.MustUnmarshal(api, []byte(kv.Value))
	value := strings.Replace(string(kv.Key), fmt.Sprintf("%s/", e.apisDir), "", 1)

	return &Evt{
		Src:   EventSrcAPI,
		Type:  evtType,
		Key:   value,
		Value: api,
	}
}

func (e *EtcdStore) doWatchWithRouting(evtType EvtType, kv *mvccpb.KeyValue) *Evt {
	routing := &model.Routing{}
	fjson.MustUnmarshal(routing, []byte(kv.Value))

	return &Evt{
		Src:   EventSrcRouting,
		Type:  evtType,
		Key:   strings.Replace(string(kv.Key), fmt.Sprintf("%s/", e.routingsDir), "", 1),
		Value: routing,
	}
}

func (e *EtcdStore) init() {
	e.watchMethodMapping[EventSrcBind] = e.doWatchWithBind
	e.watchMethodMapping[EventSrcServer] = e.doWatchWithServer
	e.watchMethodMapping[EventSrcCluster] = e.doWatchWithCluster
	e.watchMethodMapping[EventSrcAPI] = e.doWatchWithAPI
	e.watchMethodMapping[EventSrcRouting] = e.doWatchWithRouting
}

func (e *EtcdStore) put(key, value string) error {
	_, err := e.txn().Then(clientv3.OpPut(key, value)).Commit()
	return err
}

func (e *EtcdStore) putTTL(key, value string, ttl int64) error {
	lessor := clientv3.NewLease(e.cli)
	defer lessor.Close()

	ctx, cancel := context.WithTimeout(e.cli.Ctx(), DefaultRequestTimeout)
	leaseResp, err := lessor.Grant(ctx, ttl)
	cancel()

	if err != nil {
		return err
	}

	_, err = e.txn().Then(clientv3.OpPut(key, value, clientv3.WithLease(leaseResp.ID))).Commit()
	return err
}

func (e *EtcdStore) delete(key string) error {
	_, err := e.txn().Then(clientv3.OpDelete(key)).Commit()
	return err
}

func (e *EtcdStore) getList(key string, fn func(*mvccpb.KeyValue)) error {
	ctx, cancel := context.WithTimeout(e.cli.Ctx(), DefaultRequestTimeout)
	defer cancel()

	resp, err := clientv3.NewKV(e.cli).Get(ctx, key, clientv3.WithPrefix())
	if nil != err {
		return err
	}

	for _, item := range resp.Kvs {
		fn(item)
	}

	return nil
}
