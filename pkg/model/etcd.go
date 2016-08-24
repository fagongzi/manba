package model

import (
	"container/list"
	"fmt"
	"net/url"
	"strings"

	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/coreos/go-etcd/etcd"
	"github.com/fagongzi/gateway/pkg/util"
)

// EtcdStore etcd store impl
type EtcdStore struct {
	prefix                string
	clustersDir           string
	serversDir            string
	bindsDir              string
	aggregationsDir       string
	proxiesDir            string
	routingsDir           string
	deleteServersDir      string
	deleteClustersDir     string
	deleteAggregationsDir string

	cli *etcd.Client

	watchCh chan *etcd.Response
	evtCh   chan *Evt

	watchMethodMapping map[EvtSrc]func(EvtType, *etcd.Response) *Evt
}

// NewEtcdStore create a etcd store
func NewEtcdStore(etcdAddrs []string, prefix string) (Store, error) {
	store := EtcdStore{
		prefix:                prefix,
		clustersDir:           fmt.Sprintf("%s/clusters", prefix),
		serversDir:            fmt.Sprintf("%s/servers", prefix),
		bindsDir:              fmt.Sprintf("%s/binds", prefix),
		aggregationsDir:       fmt.Sprintf("%s/aggregations", prefix),
		proxiesDir:            fmt.Sprintf("%s/proxy", prefix),
		routingsDir:           fmt.Sprintf("%s/routings", prefix),
		deleteServersDir:      fmt.Sprintf("%s/delete/servers", prefix),
		deleteClustersDir:     fmt.Sprintf("%s/delete/clusters", prefix),
		deleteAggregationsDir: fmt.Sprintf("%s/delete/aggregations", prefix),

		cli:                etcd.NewClient(etcdAddrs),
		watchMethodMapping: make(map[EvtSrc]func(EvtType, *etcd.Response) *Evt),
	}

	store.init()
	return store, nil
}

// SaveAggregation save a aggregation in store
func (e EtcdStore) SaveAggregation(agn *Aggregation) error {
	key := fmt.Sprintf("%s/%s", e.aggregationsDir, url.QueryEscape(agn.URL))
	_, err := e.cli.Create(key, string(agn.Marshal()), 0)

	return err
}

// UpdateAggregation update a aggregation in store
func (e EtcdStore) UpdateAggregation(agn *Aggregation) error {
	key := fmt.Sprintf("%s/%s", e.aggregationsDir, url.QueryEscape(agn.URL))
	_, err := e.cli.Set(key, string(agn.Marshal()), 0)

	return err
}

// DeleteAggregation delete a aggregation from store
func (e EtcdStore) DeleteAggregation(aggregationURL string) error {
	return e.deleteKey(url.QueryEscape(aggregationURL), e.aggregationsDir, e.deleteAggregationsDir)
}

// GetAggregations return aggregations from store
func (e EtcdStore) GetAggregations() ([]*Aggregation, error) {
	rsp, err := e.cli.Get(e.aggregationsDir, true, false)

	if nil != err {
		return nil, err
	}

	l := rsp.Node.Nodes.Len()
	angs := make([]*Aggregation, l)

	for i := 0; i < l; i++ {
		angs[i] = UnMarshalAggregation([]byte(rsp.Node.Nodes[i].Value))
	}

	return angs, nil
}

// SaveServer save a server to store
func (e EtcdStore) SaveServer(svr *Server) error {
	key := fmt.Sprintf("%s/%s", e.serversDir, svr.Addr)
	_, err := e.cli.Create(key, string(svr.Marshal()), 0)

	return err
}

// UpdateServer update a server to store
func (e EtcdStore) UpdateServer(svr *Server) error {
	old, err := e.GetServer(svr.Addr, false)

	if nil != err {
		return err
	}

	old.updateFrom(svr)

	key := fmt.Sprintf("%s/%s", e.serversDir, old.Addr)
	_, err = e.cli.Set(key, string(old.Marshal()), 0)

	return err
}

// DeleteServer delete a server from store
func (e EtcdStore) DeleteServer(addr string) error {
	err := e.deleteKey(addr, e.serversDir, e.deleteServersDir)

	if err != nil {
		return err
	}

	// TODO: delete bind

	return nil
}

// GetServers return server from store
func (e EtcdStore) GetServers() ([]*Server, error) {
	rsp, err := e.cli.Get(e.serversDir, true, false)

	if nil != err {
		return nil, err
	}

	l := rsp.Node.Nodes.Len()
	servers := make([]*Server, l)

	for i := 0; i < l; i++ {
		servers[i] = UnMarshalServer([]byte(rsp.Node.Nodes[i].Value))
	}

	return servers, nil
}

// GetServer return spec server
// if withBinded is true, return with binded cluster
func (e EtcdStore) GetServer(serverAddr string, withBinded bool) (*Server, error) {
	key := fmt.Sprintf("%s/%s", e.serversDir, serverAddr)
	rsp, err := e.cli.Get(key, false, false)

	if nil != err {
		return nil, err
	}

	server := UnMarshalServer([]byte(rsp.Node.Value))

	if withBinded {
		bindsValues, err := e.GetBindedClusters(serverAddr)

		if nil != err {
			return nil, err
		}

		server.BindClusters = bindsValues
	}

	return server, nil
}

// GetBindedServers return cluster bind servers
func (e EtcdStore) GetBindedServers(clusterName string) ([]string, error) {
	return e.getBindedValues(clusterName, 1, 0)
}

// SaveCluster save a cluster to store
func (e EtcdStore) SaveCluster(cluster *Cluster) error {
	key := fmt.Sprintf("%s/%s", e.clustersDir, cluster.Name)
	_, err := e.cli.Create(key, string(cluster.Marshal()), 0)

	return err
}

// UpdateCluster update a cluster to store
func (e EtcdStore) UpdateCluster(cluster *Cluster) error {
	old, err := e.GetCluster(cluster.Name, false)

	if nil != err {
		return err
	}

	old.updateFrom(cluster)

	key := fmt.Sprintf("%s/%s", e.clustersDir, old.Name)
	_, err = e.cli.Set(key, string(old.Marshal()), 0)

	return err
}

// DeleteCluster delete a cluster from store
func (e EtcdStore) DeleteCluster(name string) error {
	err := e.deleteKey(name, e.clustersDir, e.deleteClustersDir)

	if err != nil {
		return err
	}

	// TODO: delete bind

	return nil
}

// GetClusters return clusters in store
func (e EtcdStore) GetClusters() ([]*Cluster, error) {
	rsp, err := e.cli.Get(e.clustersDir, true, false)

	if nil != err {
		return nil, err
	}

	l := rsp.Node.Nodes.Len()
	clusters := make([]*Cluster, l)

	for i := 0; i < l; i++ {
		clusters[i] = UnMarshalCluster([]byte(rsp.Node.Nodes[i].Value))
	}

	return clusters, nil
}

// GetCluster return cluster info
// if withBinded is true, return with binded servers
func (e EtcdStore) GetCluster(clusterName string, withBinded bool) (*Cluster, error) {
	key := fmt.Sprintf("%s/%s", e.clustersDir, clusterName)
	rsp, err := e.cli.Get(key, false, false)

	if nil != err {
		return nil, err
	}

	cluster := UnMarshalCluster([]byte(rsp.Node.Value))

	if withBinded {
		bindsValues, err := e.GetBindedServers(clusterName)

		if nil != err {
			return nil, err
		}

		cluster.BindServers = bindsValues
	}

	return cluster, nil
}

// GetBindedClusters return spec server binded clusters
func (e EtcdStore) GetBindedClusters(serverAddr string) ([]string, error) {
	return e.getBindedValues(serverAddr, 0, 1)
}

// SaveBind save bind to store
func (e EtcdStore) SaveBind(bind *Bind) error {
	key := fmt.Sprintf("%s/%s", e.bindsDir, bind.ToString())
	_, err := e.cli.Create(key, "", 0)

	return err
}

// UnBind delete bind from store
func (e EtcdStore) UnBind(bind *Bind) error {
	key := fmt.Sprintf("%s/%s", e.bindsDir, bind.ToString())

	_, err := e.cli.Delete(key, true)

	return err
}

// GetBinds return binds info
func (e EtcdStore) GetBinds() ([]*Bind, error) {
	rsp, err := e.cli.Get(e.bindsDir, false, false)

	if nil != err {
		return nil, err
	}

	l := rsp.Node.Nodes.Len()
	values := make([]*Bind, l)

	for i := 0; i < l; i++ {
		key := strings.Replace(rsp.Node.Nodes[i].Key, fmt.Sprintf("%s/", e.bindsDir), "", 1)
		infos := strings.SplitN(key, "-", 2)

		values[i] = &Bind{
			ServerAddr:  infos[0],
			ClusterName: infos[1],
		}
	}

	return values, nil
}

// SaveRouting save route to store
func (e EtcdStore) SaveRouting(routing *Routing) error {
	key := fmt.Sprintf("%s/%s", e.routingsDir, routing.ID)
	_, err := e.cli.Create(key, string(routing.Marshal()), uint64(routing.deadline))

	return err
}

// GetRoutings return routes in store
func (e EtcdStore) GetRoutings() ([]*Routing, error) {
	rsp, err := e.cli.Get(e.routingsDir, true, false)

	if nil != err {
		return nil, err
	}

	l := rsp.Node.Nodes.Len()
	routings := make([]*Routing, l)

	for i := 0; i < l; i++ {
		routings[i] = UnMarshalRouting([]byte(rsp.Node.Nodes[i].Value))
	}

	return routings, nil
}

// Clean clean data in store
func (e EtcdStore) Clean() error {
	_, err := e.cli.Delete(e.prefix, true)

	return err
}

// GC exec gc, delete some data
func (e EtcdStore) GC() error {
	// process not complete delete opts
	err := e.gcDir(e.deleteServersDir, e.DeleteServer)

	if nil != err {
		return err
	}

	err = e.gcDir(e.deleteClustersDir, e.DeleteCluster)

	if nil != err {
		return err
	}

	err = e.gcDir(e.deleteAggregationsDir, e.DeleteAggregation)

	if nil != err {
		return err
	}

	return nil
}

func (e EtcdStore) gcDir(dir string, fn func(value string) error) error {
	rsp, err := e.cli.Get(dir, false, true)
	if err != nil {
		return err
	}

	for _, node := range rsp.Node.Nodes {
		err = fn(node.Value)

		if err != nil {
			return err
		}
	}

	return nil
}

func (e EtcdStore) getBindedValues(target string, matchIndex, valueIndex int) ([]string, error) {
	rsp, err := e.cli.Get(e.bindsDir, false, false)

	if nil != err {
		return nil, err
	}

	values := list.New()
	l := rsp.Node.Nodes.Len()

	for i := 0; i < l; i++ {
		key := strings.Replace(rsp.Node.Nodes[i].Key, fmt.Sprintf("%s/", e.bindsDir), "", 1)
		infos := strings.SplitN(key, "-", 2)

		if len(infos) == 2 && target == infos[matchIndex] {
			values.PushBack(infos[valueIndex])
		}
	}

	return util.ToStringArray(values), nil
}

// Watch watch event from etcd
func (e EtcdStore) Watch(evtCh chan *Evt, stopCh chan bool) error {
	e.watchCh = make(chan *etcd.Response)
	e.evtCh = evtCh

	log.Infof("Etcd watch at: <%s>", e.prefix)

	go e.doWatch()

	_, err := e.cli.Watch(e.prefix, 0, true, e.watchCh, stopCh)
	return err
}

func (e EtcdStore) deleteKey(value, prefixKey, cacheKey string) error {
	deleteKey := fmt.Sprintf("%s/%s", cacheKey, value)
	_, err := e.cli.Set(deleteKey, value, 0)

	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s/%s", prefixKey, value)
	_, err = e.cli.Delete(key, true)

	if nil != err {
		return err
	}

	_, err = e.cli.Delete(deleteKey, true)

	return err
}

func (e EtcdStore) doWatch() {
	for {
		rsp := <-e.watchCh

		var evtSrc EvtSrc
		var evtType EvtType
		key := rsp.Node.Key

		if strings.HasPrefix(key, e.clustersDir) {
			evtSrc = EventSrcCluster
		} else if strings.HasPrefix(key, e.serversDir) {
			evtSrc = EventSrcServer
		} else if strings.HasPrefix(key, e.bindsDir) {
			evtSrc = EventSrcBind
		} else if strings.HasPrefix(key, e.aggregationsDir) {
			evtSrc = EventSrcAggregation
		} else if strings.HasPrefix(key, e.routingsDir) {
			evtSrc = EventSrcRouting
		} else {
			continue
		}

		log.Infof("Etcd changed: <%s, %s>", rsp.Node.Key, rsp.Action)

		if rsp.Action == "set" {
			if rsp.PrevNode == nil {
				evtType = EventTypeNew
			} else {
				evtType = EventTypeUpdate
			}
		} else if rsp.Action == "create" {
			evtType = EventTypeNew
		} else if rsp.Action == "delete" || rsp.Action == "expire" {
			evtType = EventTypeDelete
		} else {
			// unknow not support
			continue
		}

		e.evtCh <- e.watchMethodMapping[evtSrc](evtType, rsp)
	}
}

func (e EtcdStore) doWatchWithCluster(evtType EvtType, rsp *etcd.Response) *Evt {
	cluster := UnMarshalCluster([]byte(rsp.Node.Value))

	return &Evt{
		Src:   EventSrcCluster,
		Type:  evtType,
		Key:   strings.Replace(rsp.Node.Key, fmt.Sprintf("%s/", e.clustersDir), "", 1),
		Value: cluster,
	}
}

func (e EtcdStore) doWatchWithServer(evtType EvtType, rsp *etcd.Response) *Evt {
	server := UnMarshalServer([]byte(rsp.Node.Value))

	return &Evt{
		Src:   EventSrcServer,
		Type:  evtType,
		Key:   strings.Replace(rsp.Node.Key, fmt.Sprintf("%s/", e.serversDir), "", 1),
		Value: server,
	}
}

func (e EtcdStore) doWatchWithBind(evtType EvtType, rsp *etcd.Response) *Evt {
	key := strings.Replace(rsp.Node.Key, fmt.Sprintf("%s/", e.bindsDir), "", 1)
	infos := strings.SplitN(key, "-", 2)

	return &Evt{
		Src:  EventSrcBind,
		Type: evtType,
		Key:  rsp.Node.Key,
		Value: &Bind{
			ServerAddr:  infos[0],
			ClusterName: infos[1],
		},
	}
}

func (e EtcdStore) doWatchWithAggregation(evtType EvtType, rsp *etcd.Response) *Evt {
	ang := UnMarshalAggregation([]byte(rsp.Node.Value))
	value, _ := url.QueryUnescape(strings.Replace(rsp.Node.Key, fmt.Sprintf("%s/", e.aggregationsDir), "", 1))

	return &Evt{
		Src:   EventSrcAggregation,
		Type:  evtType,
		Key:   value,
		Value: ang,
	}
}

func (e EtcdStore) doWatchWithRouting(evtType EvtType, rsp *etcd.Response) *Evt {
	routing := UnMarshalRouting([]byte(rsp.Node.Value))

	return &Evt{
		Src:   EventSrcRouting,
		Type:  evtType,
		Key:   strings.Replace(rsp.Node.Key, fmt.Sprintf("%s/", e.routingsDir), "", 1),
		Value: routing,
	}
}

func (e EtcdStore) init() {
	e.watchMethodMapping[EventSrcBind] = e.doWatchWithBind
	e.watchMethodMapping[EventSrcServer] = e.doWatchWithServer
	e.watchMethodMapping[EventSrcCluster] = e.doWatchWithCluster
	e.watchMethodMapping[EventSrcAggregation] = e.doWatchWithAggregation
	e.watchMethodMapping[EventSrcRouting] = e.doWatchWithRouting
}
