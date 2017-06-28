package model

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"encoding/json"

	"sync"

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
}

// NewConsulStore returns a consul implemention store
func NewConsulStore(consulAddr string, prefix string) (Store, error) {
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

func (s *consulStore) SaveBind(bind *Bind) error {
	svr, err := s.GetServer(bind.ServerAddr)
	if err != nil {
		return err
	}
	svr.AddBind(bind)

	cluster, err := s.GetCluster(bind.ClusterName)
	if err != nil {
		return err
	}
	cluster.AddBind(bind)

	bindKey := fmt.Sprintf("%s/%s", s.bindsDir, bind.ToString())
	svrKey := fmt.Sprintf("%s/%s", s.serversDir, svr.Addr)
	clusterKey := fmt.Sprintf("%s/%s", s.clustersDir, cluster.Name)

	txn := api.KVTxnOps{
		&api.KVTxnOp{
			Verb:  api.KVSet,
			Key:   bindKey,
			Value: bind.Marshal(),
		},
		&api.KVTxnOp{
			Verb:  api.KVSet,
			Key:   svrKey,
			Value: svr.Marshal(),
		},
		&api.KVTxnOp{
			Verb:  api.KVSet,
			Key:   clusterKey,
			Value: cluster.Marshal(),
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

func (s *consulStore) UnBind(bind *Bind) error {
	svr, err := s.GetServer(bind.ServerAddr)
	if err != nil {
		return err
	}
	svr.RemoveBind(bind.ClusterName)

	c, err := s.GetCluster(bind.ClusterName)
	if err != nil {
		return err
	}
	c.RemoveBind(bind.ServerAddr)

	bindKey := fmt.Sprintf("%s/%s", s.bindsDir, bind.ToString())
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
			Value: svr.Marshal(),
		},
		&api.KVTxnOp{
			Verb:  api.KVSet,
			Key:   clusterKey,
			Value: c.Marshal(),
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

func (s *consulStore) GetServer(serverAddr string) (*Server, error) {
	key := fmt.Sprintf("%s/%s", s.serversDir, serverAddr)

	pair, _, err := s.client.KV().Get(key, nil)
	if err != nil {
		return nil, err
	}

	if nil == pair {
		return new(Server), nil
	}

	return UnMarshalServer(pair.Value), nil
}

func (s *consulStore) GetCluster(clusterName string) (*Cluster, error) {
	key := fmt.Sprintf("%s/%s", s.clustersDir, clusterName)
	pair, _, err := s.client.KV().Get(key, nil)
	if err != nil {
		return nil, err
	}

	if nil == pair {
		return new(Cluster), nil
	}

	return UnMarshalCluster(pair.Value), nil
}

func (s *consulStore) GetBinds() ([]*Bind, error) {
	pairs, _, err := s.client.KV().List(s.bindsDir, nil)

	if nil != err {
		return nil, err
	}

	values := make([]*Bind, len(pairs))
	i := 0

	for _, pair := range pairs {
		key := strings.Replace(pair.Key, fmt.Sprintf("%s/", s.bindsDir), "", 1)
		infos := strings.SplitN(key, "-", 2)

		values[i] = &Bind{
			ServerAddr:  infos[0],
			ClusterName: infos[1],
		}

		i++
	}

	return values, nil
}

func (s *consulStore) SaveCluster(cluster *Cluster) error {
	return s.doPutCluster(cluster, EventTypeNew)
}

func (s *consulStore) doPutCluster(cluster *Cluster, et EvtType) error {
	key := fmt.Sprintf("%s/%s", s.clustersDir, cluster.Name)
	_, err := s.client.KV().Put(&api.KVPair{
		Key:   key,
		Value: cluster.Marshal(),
	}, nil)

	if err != nil {
		return err
	}

	return nil
}

func (s *consulStore) UpdateCluster(cluster *Cluster) error {
	old, err := s.GetCluster(cluster.Name)
	if nil != err {
		return err
	}

	old.updateFrom(cluster)
	return s.doPutCluster(old, EventTypeUpdate)
}

func (s *consulStore) DeleteCluster(name string) error {
	c, err := s.GetCluster(name)
	if err != nil {
		return err
	}

	if c.HasBind() {
		return ErrHasBind
	}

	key := fmt.Sprintf("%s/%s", s.clustersDir, c.Name)
	_, err = s.client.KV().Delete(key, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *consulStore) GetClusters() ([]*Cluster, error) {
	pairs, _, err := s.client.KV().List(s.clustersDir, nil)

	if nil != err {
		return nil, err
	}

	values := make([]*Cluster, len(pairs))
	i := 0

	for _, pair := range pairs {
		values[i] = UnMarshalCluster(pair.Value)

		i++
	}

	return values, nil
}

func (s *consulStore) SaveServer(svr *Server) error {
	return s.doPutServer(svr)
}

func (s *consulStore) doPutServer(svr *Server) error {
	key := fmt.Sprintf("%s/%s", s.serversDir, svr.Addr)
	_, err := s.client.KV().Put(&api.KVPair{
		Key:   key,
		Value: svr.Marshal(),
	}, nil)

	if err != nil {
		return err
	}

	return nil
}

func (s *consulStore) UpdateServer(svr *Server) error {
	old, err := s.GetServer(svr.Addr)

	if nil != err {
		return err
	}

	old.updateFrom(svr)
	return s.doPutServer(old)
}

func (s *consulStore) DeleteServer(addr string) error {
	svr, err := s.GetServer(addr)
	if err != nil {
		return err
	}

	if svr.HasBind() {
		return ErrHasBind
	}

	key := fmt.Sprintf("%s/%s", s.serversDir, addr)
	_, err = s.client.KV().Delete(key, nil)
	if err != nil {
		return err
	}

	return nil
}

func (s *consulStore) GetServers() ([]*Server, error) {
	pairs, _, err := s.client.KV().List(s.serversDir, nil)

	if nil != err {
		return nil, err
	}

	values := make([]*Server, len(pairs))
	i := 0

	for _, pair := range pairs {
		values[i] = UnMarshalServer(pair.Value)
		i++
	}

	return values, nil
}

func (s *consulStore) SaveAPI(ap *API) error {
	return s.doPutAPI(ap, EventTypeNew)
}

func (s *consulStore) doPutAPI(ap *API, et EvtType) error {
	key := fmt.Sprintf("%s/%s", s.apisDir, getAPIKey(ap.URL, ap.Method))
	_, err := s.client.KV().Put(&api.KVPair{
		Key:   key,
		Value: ap.Marshal(),
	}, nil)

	if err != nil {
		return err
	}

	return nil
}

func (s *consulStore) UpdateAPI(api *API) error {
	return s.doPutAPI(api, EventTypeUpdate)
}

func (s *consulStore) DeleteAPI(url string, method string) error {
	key := fmt.Sprintf("%s/%s", s.apisDir, getAPIKey(url, method))
	_, err := s.client.KV().Delete(key, nil)
	if err != nil {
		return nil
	}

	return nil
}

func (s *consulStore) GetAPIs() ([]*API, error) {
	pairs, _, err := s.client.KV().List(s.apisDir, nil)

	if nil != err {
		return nil, err
	}

	values := make([]*API, len(pairs))
	i := 0

	for _, pair := range pairs {
		values[i] = UnMarshalAPI(pair.Value)
		i++
	}

	return values, nil
}

func (s *consulStore) GetAPI(url string, method string) (*API, error) {
	key := fmt.Sprintf("%s/%s", s.apisDir, getAPIKey(url, method))
	pair, _, err := s.client.KV().Get(key, nil)

	if nil != err {
		return nil, err
	}

	return UnMarshalAPI(pair.Value), nil
}

func (s *consulStore) SaveRouting(routing *Routing) error {
	return s.doPutRouting(routing, EventTypeNew)
}

func (s *consulStore) doPutRouting(routing *Routing, et EvtType) error {
	key := fmt.Sprintf("%s/%s", s.routingsDir, routing.ID)

	_, err := s.client.KV().Put(&api.KVPair{
		Key:   key,
		Value: routing.Marshal(),
	}, nil)

	if err != nil {
		return nil
	}

	return nil
}

func (s *consulStore) GetRoutings() ([]*Routing, error) {
	pairs, _, err := s.client.KV().List(s.apisDir, nil)

	if nil != err {
		return nil, err
	}

	values := make([]*Routing, len(pairs))
	i := 0

	for _, pair := range pairs {
		values[i] = UnMarshalRouting(pair.Value)
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
			e.Value = UnMarshalCluster(data)
		}
	})
	if err != nil {
		return err
	}
	plans = append(plans, p)

	p, err = s.watchPrefix(evtCh, EventSrcAPI, s.apisDir, func(data []byte, e *Evt) {
		e.Value = UnMarshalAPI(data)
	})
	if err != nil {
		return err
	}
	plans = append(plans, p)

	p, err = s.watchPrefix(evtCh, EventSrcServer, s.serversDir, func(data []byte, e *Evt) {
		e.Value = UnMarshalServer(data)
	})
	if err != nil {
		return err
	}
	plans = append(plans, p)

	p, err = s.watchPrefix(evtCh, EventSrcRouting, s.routingsDir, func(data []byte, e *Evt) {
		e.Value = UnMarshalRouting(data)
	})
	if err != nil {
		return err
	}
	plans = append(plans, p)

	p, err = s.watchPrefix(evtCh, EventSrcBind, s.bindsDir, func(data []byte, e *Evt) {
		e.Value = UnMarshalBind(data)
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
