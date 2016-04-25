package model

import (
	"container/list"
	"fmt"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/coreos/go-etcd/etcd"
	"github.com/fagongzi/gateway/pkg/util"
	"net/url"
	"strings"
)

type EtcdStore struct {
	prefix                string
	clustersDir           string
	serversDir            string
	bindsDir              string
	aggregationsDir       string
	proxiesDir            string
	deleteServersDir      string
	deleteClustersDir     string
	deleteAggregationsDir string

	cli *etcd.Client

	watchCh chan *etcd.Response
	evtCh   chan *Evt

	watchMethodMapping map[EvtSrc]func(EvtType, *etcd.Response) *Evt
}

func NewEtcdStore(etcdAddrs []string, prefix string) (Store, error) {
	store := EtcdStore{
		prefix:                prefix,
		clustersDir:           fmt.Sprintf("%s/clusters", prefix),
		serversDir:            fmt.Sprintf("%s/servers", prefix),
		bindsDir:              fmt.Sprintf("%s/binds", prefix),
		aggregationsDir:       fmt.Sprintf("%s/aggregations", prefix),
		proxiesDir:            fmt.Sprintf("%s/proxy", prefix),
		deleteServersDir:      fmt.Sprintf("%s/delete/servers", prefix),
		deleteClustersDir:     fmt.Sprintf("%s/delete/clusters", prefix),
		deleteAggregationsDir: fmt.Sprintf("%s/delete/aggregations", prefix),

		cli:                etcd.NewClient(etcdAddrs),
		watchMethodMapping: make(map[EvtSrc]func(EvtType, *etcd.Response) *Evt),
	}

	store.init()
	return store, nil
}

func (self EtcdStore) SaveAggregation(agn *Aggregation) error {
	key := fmt.Sprintf("%s/%s", self.aggregationsDir, url.QueryEscape(agn.Url))
	_, err := self.cli.Create(key, string(agn.Marshal()), 0)

	return err
}

func (self EtcdStore) UpdateAggregation(agn *Aggregation) error {
	key := fmt.Sprintf("%s/%s", self.aggregationsDir, url.QueryEscape(agn.Url))
	_, err := self.cli.Set(key, string(agn.Marshal()), 0)

	return err
}

func (self EtcdStore) DeleteAggregation(aggregationUrl string) error {
	return self.deleteKey(url.QueryEscape(aggregationUrl), self.aggregationsDir, self.deleteAggregationsDir)
}

func (self EtcdStore) GetAggregations() ([]*Aggregation, error) {
	rsp, err := self.cli.Get(self.aggregationsDir, true, false)

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

func (self EtcdStore) SaveServer(svr *Server) error {
	key := fmt.Sprintf("%s/%s", self.serversDir, svr.Addr)
	_, err := self.cli.Create(key, string(svr.Marshal()), 0)

	return err
}

func (self EtcdStore) UpdateServer(svr *Server) error {
	old, err := self.GetServer(svr.Addr, false)

	if nil != err {
		return err
	}

	old.updateFrom(svr)

	key := fmt.Sprintf("%s/%s", self.serversDir, old.Addr)
	_, err = self.cli.Set(key, string(old.Marshal()), 0)

	return err
}

func (self EtcdStore) DeleteServer(addr string) error {
	err := self.deleteKey(addr, self.serversDir, self.deleteServersDir)

	if err != nil {
		return err
	}

	// TODO: delete bind

	return nil
}

func (self EtcdStore) GetServers() ([]*Server, error) {
	rsp, err := self.cli.Get(self.serversDir, true, false)

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

func (self EtcdStore) GetServer(serverAddr string, withBinded bool) (*Server, error) {
	key := fmt.Sprintf("%s/%s", self.serversDir, serverAddr)
	rsp, err := self.cli.Get(key, false, false)

	if nil != err {
		return nil, err
	}

	server := UnMarshalServer([]byte(rsp.Node.Value))

	if withBinded {
		bindsValues, err := self.GetBindedClusters(serverAddr)

		if nil != err {
			return nil, err
		}

		server.BindClusters = bindsValues
	}

	return server, nil
}

func (self EtcdStore) GetBindedServers(clusterName string) ([]string, error) {
	return self.getBindedValues(clusterName, 1, 0)
}

func (self EtcdStore) SaveCluster(cluster *Cluster) error {
	key := fmt.Sprintf("%s/%s", self.clustersDir, cluster.Name)
	_, err := self.cli.Create(key, string(cluster.Marshal()), 0)

	return err
}

func (self EtcdStore) UpdateCluster(cluster *Cluster) error {
	old, err := self.GetCluster(cluster.Name, false)

	if nil != err {
		return err
	}

	old.updateFrom(cluster)

	key := fmt.Sprintf("%s/%s", self.clustersDir, old.Name)
	_, err = self.cli.Set(key, string(old.Marshal()), 0)

	return err
}

func (self EtcdStore) DeleteCluster(name string) error {
	err := self.deleteKey(name, self.clustersDir, self.deleteClustersDir)

	if err != nil {
		return err
	}

	// TODO: delete bind

	return nil
}

func (self EtcdStore) GetClusters() ([]*Cluster, error) {
	rsp, err := self.cli.Get(self.clustersDir, true, false)

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

func (self EtcdStore) GetCluster(clusterName string, withBinded bool) (*Cluster, error) {
	key := fmt.Sprintf("%s/%s", self.clustersDir, clusterName)
	rsp, err := self.cli.Get(key, false, false)

	if nil != err {
		return nil, err
	}

	cluster := UnMarshalCluster([]byte(rsp.Node.Value))

	if withBinded {
		bindsValues, err := self.GetBindedServers(clusterName)

		if nil != err {
			return nil, err
		}

		cluster.BindServers = bindsValues
	}

	return cluster, nil
}

func (self EtcdStore) GetBindedClusters(serverAddr string) ([]string, error) {
	return self.getBindedValues(serverAddr, 0, 1)
}

func (self EtcdStore) SaveBind(bind *Bind) error {
	key := fmt.Sprintf("%s/%s", self.bindsDir, bind.ToString())
	_, err := self.cli.Create(key, "", 0)

	return err
}

func (self EtcdStore) UnBind(bind *Bind) error {
	key := fmt.Sprintf("%s/%s", self.bindsDir, bind.ToString())

	_, err := self.cli.Delete(key, true)

	return err
}

func (self EtcdStore) GetBinds() ([]*Bind, error) {
	rsp, err := self.cli.Get(self.bindsDir, false, false)

	if nil != err {
		return nil, err
	}

	l := rsp.Node.Nodes.Len()
	values := make([]*Bind, l)

	for i := 0; i < l; i++ {
		key := strings.Replace(rsp.Node.Nodes[i].Key, fmt.Sprintf("%s/", self.bindsDir), "", 1)
		infos := strings.SplitN(key, "-", 2)

		values[i] = &Bind{
			ServerAddr:  infos[0],
			ClusterName: infos[1],
		}
	}

	return values, nil
}

func (self EtcdStore) Clean() error {
	_, err := self.cli.Delete(self.prefix, true)

	return err
}

func (self EtcdStore) GC() error {
	// process not complete delete opts
	err := self.gcDir(self.deleteServersDir, self.DeleteServer)

	if nil != err {
		return err
	}

	err = self.gcDir(self.deleteClustersDir, self.DeleteCluster)

	if nil != err {
		return err
	}

	err = self.gcDir(self.deleteAggregationsDir, self.DeleteAggregation)

	if nil != err {
		return err
	}

	return nil
}

func (self EtcdStore) gcDir(dir string, fn func(value string) error) error {
	rsp, err := self.cli.Get(dir, false, true)
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

func (self EtcdStore) getBindedValues(target string, matchIndex, valueIndex int) ([]string, error) {
	rsp, err := self.cli.Get(self.bindsDir, false, false)

	if nil != err {
		return nil, err
	}

	values := list.New()
	l := rsp.Node.Nodes.Len()

	for i := 0; i < l; i++ {
		key := strings.Replace(rsp.Node.Nodes[i].Key, fmt.Sprintf("%s/", self.bindsDir), "", 1)
		infos := strings.SplitN(key, "-", 2)

		if len(infos) == 2 && target == infos[matchIndex] {
			values.PushBack(infos[valueIndex])
		}
	}

	return util.ToStringArray(values), nil
}

func (self EtcdStore) Watch(evtCh chan *Evt, stopCh chan bool) error {
	self.watchCh = make(chan *etcd.Response)
	self.evtCh = evtCh

	log.Infof("Etcd watch at: <%s>", self.prefix)

	go self.doWatch()

	_, err := self.cli.Watch(self.prefix, 0, true, self.watchCh, stopCh)
	return err
}

func (self EtcdStore) deleteKey(value, prefixKey, cacheKey string) error {
	deleteKey := fmt.Sprintf("%s/%s", cacheKey, value)
	_, err := self.cli.Set(deleteKey, value, 0)

	if err != nil {
		return err
	}

	key := fmt.Sprintf("%s/%s", prefixKey, value)
	_, err = self.cli.Delete(key, true)

	if nil != err {
		return err
	}

	_, err = self.cli.Delete(deleteKey, true)

	return err
}

func (self EtcdStore) doWatch() {
	for {
		rsp := <-self.watchCh

		var evtSrc EvtSrc
		var evtType EvtType
		key := rsp.Node.Key

		if strings.HasPrefix(key, self.clustersDir) {
			evtSrc = EVT_SRC_CLUSTER
		} else if strings.HasPrefix(key, self.serversDir) {
			evtSrc = EVT_SRC_SERVER
		} else if strings.HasPrefix(key, self.bindsDir) {
			evtSrc = EVT_SRC_BIND
		} else if strings.HasPrefix(key, self.aggregationsDir) {
			evtSrc = EVT_STC_AGGREGATION
		} else {
			continue
		}

		log.Infof("Etcd changed: <%s, %s>", rsp.Node.Key, rsp.Action)

		if rsp.Action == "set" {
			if rsp.PrevNode == nil {
				evtType = EVT_TYPE_NEW
			} else {
				evtType = EVT_TYPE_UPDATE
			}
		} else if rsp.Action == "create" {
			evtType = EVT_TYPE_NEW
		} else if rsp.Action == "delete" {
			evtType = EVT_TYPE_DELETE
		} else {
			// unknow not support
			continue
		}

		self.evtCh <- self.watchMethodMapping[evtSrc](evtType, rsp)
	}
}

func (self EtcdStore) doWatchWithCluster(evtType EvtType, rsp *etcd.Response) *Evt {
	cluster := UnMarshalCluster([]byte(rsp.Node.Value))

	return &Evt{
		Src:   EVT_SRC_CLUSTER,
		Type:  evtType,
		Key:   strings.Replace(rsp.Node.Key, fmt.Sprintf("%s/", self.clustersDir), "", 1),
		Value: cluster,
	}
}

func (self EtcdStore) doWatchWithServer(evtType EvtType, rsp *etcd.Response) *Evt {
	server := UnMarshalServer([]byte(rsp.Node.Value))

	return &Evt{
		Src:   EVT_SRC_SERVER,
		Type:  evtType,
		Key:   strings.Replace(rsp.Node.Key, fmt.Sprintf("%s/", self.serversDir), "", 1),
		Value: server,
	}
}

func (self EtcdStore) doWatchWithBind(evtType EvtType, rsp *etcd.Response) *Evt {
	key := strings.Replace(rsp.Node.Key, fmt.Sprintf("%s/", self.bindsDir), "", 1)
	infos := strings.SplitN(key, "-", 2)

	return &Evt{
		Src:  EVT_SRC_BIND,
		Type: evtType,
		Key:  rsp.Node.Key,
		Value: &Bind{
			ServerAddr:  infos[0],
			ClusterName: infos[1],
		},
	}
}

func (self EtcdStore) doWatchWithAggregation(evtType EvtType, rsp *etcd.Response) *Evt {
	ang := UnMarshalAggregation([]byte(rsp.Node.Value))
	value, _ := url.QueryUnescape(strings.Replace(rsp.Node.Key, fmt.Sprintf("%s/", self.aggregationsDir), "", 1))

	return &Evt{
		Src:   EVT_STC_AGGREGATION,
		Type:  evtType,
		Key:   value,
		Value: ang,
	}
}

func (self EtcdStore) init() error {
	self.watchMethodMapping[EVT_SRC_BIND] = self.doWatchWithBind
	self.watchMethodMapping[EVT_SRC_SERVER] = self.doWatchWithServer
	self.watchMethodMapping[EVT_SRC_CLUSTER] = self.doWatchWithCluster
	self.watchMethodMapping[EVT_STC_AGGREGATION] = self.doWatchWithAggregation

	var err error
	_, err = self.cli.SetDir(self.clustersDir, 0)
	if nil != err {
		return err
	}

	_, err = self.cli.SetDir(self.serversDir, 0)
	if nil != err {
		return err
	}

	_, err = self.cli.SetDir(self.bindsDir, 0)
	if nil != err {
		return err
	}

	_, err = self.cli.SetDir(self.aggregationsDir, 0)
	if nil != err {
		return err
	}

	_, err = self.cli.SetDir(self.proxiesDir, 0)
	if nil != err {
		return err
	}

	return nil
}
