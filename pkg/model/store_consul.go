package model

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"

	"encoding/json"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/watch"
)

const (
	consulCustomEvent = "consul-custom-event"
)

type event struct {
	Data []*eventData `json:"data"`
}

type eventData struct {
	Key   string  `json:"key"`
	Type  EvtType `json:"type"`
	Src   EvtSrc  `json:"src"`
	Value []byte  `json:"value"`
}

func (e *event) add(data *eventData) {
	e.Data = append(e.Data, data)
}

func (e *event) marshal() []byte {
	value, _ := json.Marshal(e)
	return value
}

func unMarshal(data []byte) *event {
	e := new(event)
	json.Unmarshal(data, e)

	return e
}

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
			Value: cluster.Marshal(),
		},
	}

	ok, _, _, err := s.client.KV().Txn(txn, nil)
	if err != nil {
		return err
	} else if !ok {
		return errors.New("transaction should have failed")
	}

	evt := &event{}
	evt.add(&eventData{
		Key:  bindKey,
		Src:  EventSrcBind,
		Type: EventTypeNew,
	})
	evt.add(&eventData{
		Key:   svrKey,
		Src:   EventSrcServer,
		Type:  EventTypeNew,
		Value: svr.Marshal(),
	})
	evt.add(&eventData{
		Key:   clusterKey,
		Src:   EventSrcCluster,
		Type:  EventTypeNew,
		Value: cluster.Marshal(),
	})

	s.client.Event().Fire(&api.UserEvent{
		Name:    consulCustomEvent,
		Payload: evt.marshal(),
	}, nil)

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

	evt := &event{}
	evt.add(&eventData{
		Key:  bindKey,
		Src:  EventSrcBind,
		Type: EventTypeDelete,
	})
	evt.add(&eventData{
		Key:   svrKey,
		Src:   EventSrcServer,
		Type:  EventTypeUpdate,
		Value: svr.Marshal(),
	})
	evt.add(&eventData{
		Key:   clusterKey,
		Src:   EventSrcCluster,
		Type:  EventTypeUpdate,
		Value: c.Marshal(),
	})

	s.client.Event().Fire(&api.UserEvent{
		Name:    consulCustomEvent,
		Payload: evt.marshal(),
	}, nil)

	return nil
}

func (s *consulStore) GetServer(serverAddr string) (*Server, error) {
	key := fmt.Sprintf("%s/%s", s.serversDir, serverAddr)

	pair, _, err := s.client.KV().Get(key, nil)
	if err != nil {
		return nil, err
	}

	return UnMarshalServer(pair.Value), nil
}

func (s *consulStore) GetCluster(clusterName string) (*Cluster, error) {
	key := fmt.Sprintf("%s/%s", s.clustersDir, clusterName)
	pair, _, err := s.client.KV().Get(key, nil)
	if err != nil {
		return nil, err
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
	key := fmt.Sprintf("%s/%s", s.clustersDir, cluster.Name)
	_, err := s.client.KV().Put(&api.KVPair{
		Key:   key,
		Value: cluster.Marshal(),
	}, nil)

	if err != nil {
		return err
	}

	evt := &event{}
	evt.add(&eventData{
		Key:   key,
		Src:   EventSrcCluster,
		Type:  EventTypeUpdate,
		Value: cluster.Marshal(),
	})

	s.client.Event().Fire(&api.UserEvent{
		Name:    consulCustomEvent,
		Payload: evt.marshal(),
	}, nil)

	return nil
}

func (s *consulStore) UpdateCluster(cluster *Cluster) error {
	old, err := s.GetCluster(cluster.Name)
	if nil != err {
		return err
	}

	old.updateFrom(cluster)
	return s.SaveCluster(old)
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

	evt := &event{}
	evt.add(&eventData{
		Key:  key,
		Src:  EventSrcCluster,
		Type: EventTypeDelete,
	})

	s.client.Event().Fire(&api.UserEvent{
		Name:    consulCustomEvent,
		Payload: evt.marshal(),
	}, nil)

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
	key := fmt.Sprintf("%s/%s", s.serversDir, svr.Addr)
	_, err := s.client.KV().Put(&api.KVPair{
		Key:   key,
		Value: svr.Marshal(),
	}, nil)

	if err != nil {
		return err
	}

	evt := &event{}
	evt.add(&eventData{
		Key:   key,
		Src:   EventSrcServer,
		Type:  EventTypeUpdate,
		Value: svr.Marshal(),
	})

	s.client.Event().Fire(&api.UserEvent{
		Name:    consulCustomEvent,
		Payload: evt.marshal(),
	}, nil)

	return nil
}

func (s *consulStore) UpdateServer(svr *Server) error {
	old, err := s.GetServer(svr.Addr)

	if nil != err {
		return err
	}

	old.updateFrom(svr)
	return s.SaveServer(old)
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

	evt := &event{}
	evt.add(&eventData{
		Key:  key,
		Src:  EventSrcServer,
		Type: EventTypeDelete,
	})

	s.client.Event().Fire(&api.UserEvent{
		Name:    consulCustomEvent,
		Payload: evt.marshal(),
	}, nil)

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
	key := fmt.Sprintf("%s/%s", s.apisDir, getAPIKey(ap.URL, ap.Method))
	_, err := s.client.KV().Put(&api.KVPair{
		Key:   key,
		Value: ap.Marshal(),
	}, nil)

	if err != nil {
		return err
	}

	evt := &event{}
	evt.add(&eventData{
		Key:   key,
		Src:   EventSrcAPI,
		Type:  EventTypeUpdate,
		Value: ap.Marshal(),
	})

	s.client.Event().Fire(&api.UserEvent{
		Name:    consulCustomEvent,
		Payload: evt.marshal(),
	}, nil)

	return nil
}

func (s *consulStore) UpdateAPI(api *API) error {
	return s.SaveAPI(api)
}

func (s *consulStore) DeleteAPI(url string, method string) error {
	key := fmt.Sprintf("%s/%s", s.apisDir, getAPIKey(url, method))
	_, err := s.client.KV().Delete(key, nil)
	if err != nil {
		return nil
	}

	evt := &event{}
	evt.add(&eventData{
		Key:  key,
		Src:  EventSrcAPI,
		Type: EventTypeDelete,
	})

	s.client.Event().Fire(&api.UserEvent{
		Name:    consulCustomEvent,
		Payload: evt.marshal(),
	}, nil)

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
	key := fmt.Sprintf("%s/%s", s.routingsDir, routing.ID)

	_, err := s.client.KV().Put(&api.KVPair{
		Key:   key,
		Value: routing.Marshal(),
	}, nil)

	if err != nil {
		return nil
	}

	evt := &event{}
	evt.add(&eventData{
		Key:   key,
		Src:   EventSrcRouting,
		Type:  EventTypeUpdate,
		Value: routing.Marshal(),
	})

	s.client.Event().Fire(&api.UserEvent{
		Name:    consulCustomEvent,
		Payload: evt.marshal(),
	}, nil)

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

func (s *consulStore) Watch(evtCh chan *Evt, stopCh chan bool) error {
	plan, err := watch.Parse(makeParams(fmt.Sprintf(`{"type":"event", "name":"%s"}`, consulCustomEvent)))
	if err != nil {
		return err
	}

	plan.Handler = func(idx uint64, val interface{}) {
		if val == nil {
			return
		}

		events := val.([]*api.UserEvent)

		for _, e := range events {
			ent := unMarshal(e.Payload)

			for _, data := range ent.Data {
				evtCh <- &Evt{
					Src:   data.Src,
					Type:  data.Type,
					Key:   data.Key,
					Value: data.Value,
				}
			}
		}
	}

	go plan.Run(s.consulAddr)
	go func() {
		<-stopCh
		plan.Stop()
	}()

	return nil
}

func (s *consulStore) Clean() error {
	_, err := s.client.KV().DeleteTree(s.prefix, nil)
	return err
}

func (s *consulStore) GC() error {
	return nil
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
