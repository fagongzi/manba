package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/fagongzi/util/task"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/watch"
)

type consulStore struct {
	consulAddr string
	client     *api.Client

	prefix      string
	clustersDir string
	serversDir  string
	bindsDir    string
	apisDir     string
	proxiesDir  string
	routingsDir string

	taskRunner *task.Runner
}

// NewConsulStore returns a consul implemention store
func NewConsulStore(consulAddr string, prefix string, taskRunner *task.Runner) (Store, error) {
	if strings.HasPrefix(prefix, "/") {
		prefix = prefix[1:]
	}

	store := &consulStore{
		consulAddr:  consulAddr,
		prefix:      prefix,
		clustersDir: fmt.Sprintf("%s/clusters", prefix),
		serversDir:  fmt.Sprintf("%s/servers", prefix),
		bindsDir:    fmt.Sprintf("%s/binds", prefix),
		apisDir:     fmt.Sprintf("%s/apis", prefix),
		proxiesDir:  fmt.Sprintf("%s/proxy", prefix),
		routingsDir: fmt.Sprintf("%s/routings", prefix),
		taskRunner:  taskRunner,
	}

	conf := api.DefaultConfig()
	conf.Address = consulAddr

	client, err := api.NewClient(conf)
	if err != nil {
		return nil, err
	}

	store.client = client
	if err != nil {
		return nil, err
	}

	return store, nil
}

func (s *consulStore) SaveBind(bind *model.Bind) error {
	svr, err := s.GetServer(bind.ServerID)
	if err != nil {
		return err
	}
	svr.AddBind(bind)

	cluster, err := s.GetCluster(bind.ClusterID)
	if err != nil {
		return err
	}
	cluster.AddBind(bind)

	bindKey := fmt.Sprintf("%s/%s", s.bindsDir, bind.ID)
	svrKey := fmt.Sprintf("%s/%s", s.serversDir, svr.ID)
	clusterKey := fmt.Sprintf("%s/%s", s.clustersDir, cluster.ID)

	txn := api.KVTxnOps{
		&api.KVTxnOp{
			Verb:  api.KVSet,
			Key:   bindKey,
			Value: util.MustMarshal(bind),
		},
		&api.KVTxnOp{
			Verb:  api.KVSet,
			Key:   svrKey,
			Value: util.MustMarshal(svr),
		},
		&api.KVTxnOp{
			Verb:  api.KVSet,
			Key:   clusterKey,
			Value: util.MustMarshal(cluster),
		},
	}

	ok, _, _, err := s.client.KV().Txn(txn, nil)
	if err != nil {
		return err
	} else if !ok {
		return errors.New("transaction should have failed")
	}

	return nil
}

func (s *consulStore) UnBind(id string) error {
	bind, err := s.GetBind(id)
	if err != nil {
		return err
	}

	svr, err := s.GetServer(bind.ServerID)
	if err != nil {
		return err
	}
	svr.RemoveBind(bind.ClusterID)

	c, err := s.GetCluster(bind.ClusterID)
	if err != nil {
		return err
	}
	c.RemoveBind(bind.ServerID)

	bindKey := fmt.Sprintf("%s/%s", s.bindsDir, id)
	svrKey := fmt.Sprintf("%s/%s", s.serversDir, svr.Addr)
	clusterKey := fmt.Sprintf("%s/%s", s.clustersDir, c.Name)
	txn := api.KVTxnOps{
		&api.KVTxnOp{
			Verb:  api.KVDelete,
			Key:   bindKey,
			Value: []byte(""),
		},
		&api.KVTxnOp{
			Verb:  api.KVSet,
			Key:   svrKey,
			Value: util.MustMarshal(svr),
		},
		&api.KVTxnOp{
			Verb:  api.KVSet,
			Key:   clusterKey,
			Value: util.MustMarshal(c),
		},
	}

	ok, _, _, err := s.client.KV().Txn(txn, nil)
	if err != nil {
		return err
	} else if !ok {
		return errors.New("transaction should have failed")
	}

	return nil
}

func (s *consulStore) GetServer(id string) (*model.Server, error) {
	key := fmt.Sprintf("%s/%s", s.serversDir, id)

	pair, _, err := s.client.KV().Get(key, nil)
	if err != nil {
		return nil, err
	}

	if nil == pair {
		return nil, fmt.Errorf("server <%s> not found", id)
	}

	svr := &model.Server{}
	util.MustUnmarshal(svr, pair.Value)
	return svr, nil
}

func (s *consulStore) GetCluster(id string) (*model.Cluster, error) {
	key := fmt.Sprintf("%s/%s", s.clustersDir, id)
	pair, _, err := s.client.KV().Get(key, nil)
	if err != nil {
		return nil, err
	}

	if nil == pair {
		return nil, fmt.Errorf("cluster <%s> not found", id)
	}

	c := &model.Cluster{}
	util.MustUnmarshal(c, pair.Value)
	return c, nil
}

func (s *consulStore) GetBind(id string) (*model.Bind, error) {
	key := fmt.Sprintf("%s/%s", s.bindsDir, id)
	pair, _, err := s.client.KV().Get(key, nil)
	if err != nil {
		return nil, err
	}

	if nil == pair {
		return nil, fmt.Errorf("bind <%s> not found", id)
	}

	b := &model.Bind{}
	util.MustUnmarshal(b, pair.Value)
	return b, nil
}

func (s *consulStore) GetBinds() ([]*model.Bind, error) {
	pairs, _, err := s.client.KV().List(s.bindsDir, nil)

	if nil != err {
		return nil, err
	}

	values := make([]*model.Bind, len(pairs))
	i := 0

	for _, pair := range pairs {
		key := strings.Replace(pair.Key, fmt.Sprintf("%s/", s.bindsDir), "", 1)
		infos := strings.SplitN(key, "-", 2)

		values[i] = &model.Bind{
			ServerID:  infos[0],
			ClusterID: infos[1],
		}

		i++
	}

	return values, nil
}

func (s *consulStore) SaveCluster(cluster *model.Cluster) error {
	return s.doPutCluster(cluster, EventTypeNew)
}

func (s *consulStore) doPutCluster(cluster *model.Cluster, et EvtType) error {
	key := fmt.Sprintf("%s/%s", s.clustersDir, cluster.ID)
	_, err := s.client.KV().Put(&api.KVPair{
		Key:   key,
		Value: util.MustMarshal(cluster),
	}, nil)

	if err != nil {
		return err
	}

	return nil
}

func (s *consulStore) UpdateCluster(cluster *model.Cluster) error {
	if _, err := s.GetCluster(cluster.ID); nil != err {
		return err
	}

	return s.doPutCluster(cluster, EventTypeUpdate)
}

func (s *consulStore) DeleteCluster(id string) error {
	c, err := s.GetCluster(id)
	if err != nil {
		return err
	}

	if c.HasBind() {
		return ErrHasBind
	}

	key := fmt.Sprintf("%s/%s", s.clustersDir, id)
	_, err = s.client.KV().Delete(key, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *consulStore) GetClusters() ([]*model.Cluster, error) {
	pairs, _, err := s.client.KV().List(s.clustersDir, nil)

	if nil != err {
		return nil, err
	}

	values := make([]*model.Cluster, len(pairs))
	i := 0

	for _, pair := range pairs {
		c := &model.Cluster{}
		util.MustUnmarshal(c, pair.Value)
		values[i] = c

		i++
	}

	return values, nil
}

func (s *consulStore) SaveServer(svr *model.Server) error {
	return s.doPutServer(svr)
}

func (s *consulStore) doPutServer(svr *model.Server) error {
	key := fmt.Sprintf("%s/%s", s.serversDir, svr.ID)
	_, err := s.client.KV().Put(&api.KVPair{
		Key:   key,
		Value: util.MustMarshal(svr),
	}, nil)

	if err != nil {
		return err
	}

	return nil
}

func (s *consulStore) UpdateServer(svr *model.Server) error {
	if _, err := s.GetServer(svr.ID); nil != err {
		return err
	}

	return s.doPutServer(svr)
}

func (s *consulStore) DeleteServer(id string) error {
	svr, err := s.GetServer(id)
	if err != nil {
		return err
	}

	if svr.HasBind() {
		return ErrHasBind
	}

	key := fmt.Sprintf("%s/%s", s.serversDir, id)
	_, err = s.client.KV().Delete(key, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *consulStore) GetServers() ([]*model.Server, error) {
	pairs, _, err := s.client.KV().List(s.serversDir, nil)

	if nil != err {
		return nil, err
	}

	values := make([]*model.Server, len(pairs))
	i := 0

	for _, pair := range pairs {
		svr := &model.Server{}
		util.MustUnmarshal(svr, pair.Value)
		values[i] = svr
		i++
	}

	return values, nil
}

func (s *consulStore) SaveAPI(ap *model.API) error {
	return s.doPutAPI(ap, EventTypeNew)
}

func (s *consulStore) doPutAPI(ap *model.API, et EvtType) error {
	key := fmt.Sprintf("%s/%s", s.apisDir, ap.ID)
	_, err := s.client.KV().Put(&api.KVPair{
		Key:   key,
		Value: util.MustMarshal(ap),
	}, nil)

	if err != nil {
		return err
	}

	return nil
}

func (s *consulStore) UpdateAPI(api *model.API) error {
	if _, err := s.GetAPI(api.ID); nil != err {
		return err
	}

	return s.doPutAPI(api, EventTypeUpdate)
}

func (s *consulStore) DeleteAPI(id string) error {
	key := fmt.Sprintf("%s/%s", s.apisDir, id)
	_, err := s.client.KV().Delete(key, nil)
	if err != nil {
		return nil
	}

	return nil
}

func (s *consulStore) GetAPIs() ([]*model.API, error) {
	pairs, _, err := s.client.KV().List(s.apisDir, nil)

	if nil != err {
		return nil, err
	}

	values := make([]*model.API, len(pairs))
	i := 0

	for _, pair := range pairs {
		value := &model.API{}
		util.MustUnmarshal(value, pair.Value)
		values[i] = value
		i++
	}

	return values, nil
}

func (s *consulStore) GetAPI(id string) (*model.API, error) {
	key := fmt.Sprintf("%s/%s", s.apisDir, id)
	pair, _, err := s.client.KV().Get(key, nil)

	if nil != err {
		return nil, err
	}

	if nil == pair {
		return nil, fmt.Errorf("api <%s> not found", id)
	}

	value := &model.API{}
	util.MustUnmarshal(value, pair.Value)
	return value, nil
}

func (s *consulStore) SaveRouting(routing *model.Routing) error {
	return s.doPutRouting(routing, EventTypeNew)
}

func (s *consulStore) doPutRouting(routing *model.Routing, et EvtType) error {
	key := fmt.Sprintf("%s/%s", s.routingsDir, routing.ID)

	_, err := s.client.KV().Put(&api.KVPair{
		Key:   key,
		Value: util.MustMarshal(routing),
	}, nil)

	if err != nil {
		return nil
	}

	return nil
}

func (s *consulStore) GetRoutings() ([]*model.Routing, error) {
	pairs, _, err := s.client.KV().List(s.routingsDir, nil)

	if nil != err {
		return nil, err
	}

	values := make([]*model.Routing, len(pairs))
	i := 0

	for _, pair := range pairs {
		value := &model.Routing{}
		util.MustUnmarshal(value, pair.Value)
		values[i] = value
		i++
	}

	return values, nil
}

func (s *consulStore) watchPrefix(evtCh chan *Evt, src EvtSrc, prefix string, fn func([]byte, *Evt)) (*watch.Plan, error) {
	watchPrefix := fmt.Sprintf("%s/", prefix)
	plan, err := watch.Parse(makeParams(fmt.Sprintf(`{"type":"keyprefix", "prefix":"%s"}`, watchPrefix)))
	if err != nil {
		return nil, err
	}

	var lastIdx uint64
	keys := make(map[string]struct{})

	plan.Handler = func(idx uint64, val interface{}) {
		if val == nil {
			return
		}

		v, ok := val.(api.KVPairs)

		var newest map[string]struct{}
		hasDelete := false
		if lastIdx != 0 && len(keys) != len(v) {
			hasDelete = true
			newest = make(map[string]struct{}, len(v))
		}

		if ok {
			for _, p := range v {
				if hasDelete {
					newest[p.Key] = struct{}{}
				}

				if lastIdx == 0 {
					keys[p.Key] = struct{}{}
				} else {
					if _, ok := keys[p.Key]; !ok {
						keys[p.Key] = struct{}{}
						e := &Evt{
							Src:  src,
							Type: EventTypeNew,
							Key:  strings.Replace(p.Key, watchPrefix, "", 1),
						}
						fn(p.Value, e)
						evtCh <- e
					} else if p.ModifyIndex > p.CreateIndex {
						e := &Evt{
							Src:  src,
							Type: EventTypeUpdate,
							Key:  strings.Replace(p.Key, watchPrefix, "", 1),
						}
						fn(p.Value, e)
						evtCh <- e
					}
				}
			}

			if hasDelete {
				for key := range keys {
					if _, ok := newest[key]; !ok {
						evtCh <- &Evt{
							Src:  src,
							Type: EventTypeDelete,
							Key:  strings.Replace(key, watchPrefix, "", 1),
						}
					}
				}

				keys = newest
			}
		}

		lastIdx = idx
	}

	return plan, nil
}

func (s *consulStore) Watch(evtCh chan *Evt, stopCh chan bool) error {
	var plans []*watch.Plan

	p, err := s.watchPrefix(evtCh, EventSrcCluster, s.clustersDir, func(data []byte, e *Evt) {
		if nil != data {
			c := &model.Cluster{}
			util.MustUnmarshal(c, data)
			e.Value = c
		}
	})
	if err != nil {
		return err
	}
	plans = append(plans, p)

	p, err = s.watchPrefix(evtCh, EventSrcAPI, s.apisDir, func(data []byte, e *Evt) {
		value := &model.API{}
		util.MustUnmarshal(value, data)
		e.Value = value
	})
	if err != nil {
		return err
	}
	plans = append(plans, p)

	p, err = s.watchPrefix(evtCh, EventSrcServer, s.serversDir, func(data []byte, e *Evt) {
		svr := &model.Server{}
		util.MustUnmarshal(svr, data)
		e.Value = svr
	})
	if err != nil {
		return err
	}
	plans = append(plans, p)

	p, err = s.watchPrefix(evtCh, EventSrcRouting, s.routingsDir, func(data []byte, e *Evt) {
		value := &model.Routing{}
		util.MustUnmarshal(value, data)
		e.Value = value
	})
	if err != nil {
		return err
	}
	plans = append(plans, p)

	p, err = s.watchPrefix(evtCh, EventSrcBind, s.bindsDir, func(data []byte, e *Evt) {
		value := &model.Bind{}
		util.MustUnmarshal(value, data)
		e.Value = value
	})
	if err != nil {
		return err
	}
	plans = append(plans, p)

	wg := &sync.WaitGroup{}
	go func() {
		<-stopCh
		for _, p := range plans {
			p.Stop()
			wg.Done()
		}
	}()

	for _, p := range plans {
		wg.Add(1)
		go p.Run(s.consulAddr)
	}

	wg.Wait()
	return nil
}

func (s *consulStore) Clean() error {
	_, err := s.client.KV().DeleteTree(s.prefix, nil)
	return err
}

func makeParams(s string) map[string]interface{} {
	var out map[string]interface{}
	dec := json.NewDecoder(bytes.NewReader([]byte(s)))
	if err := dec.Decode(&out); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return out
}
