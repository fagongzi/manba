package store

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/pb/rpcpb"
	"github.com/fagongzi/gateway/pkg/util"
)

var (
	// TICKER ticket
	TICKER = time.Second * 3
	// TTL timeout
	TTL = int64(5)
)

var (
	supportSchema = make(map[string]func(string, string) (Store, error))
)

// EvtType event type
type EvtType int

// EvtSrc event src
type EvtSrc int

const (
	// EventTypeNew event type new
	EventTypeNew = EvtType(0)
	// EventTypeUpdate event type update
	EventTypeUpdate = EvtType(1)
	// EventTypeDelete event type delete
	EventTypeDelete = EvtType(2)
)

const (
	// EventSrcCluster cluster event
	EventSrcCluster = EvtSrc(0)
	// EventSrcServer server event
	EventSrcServer = EvtSrc(1)
	// EventSrcBind bind event
	EventSrcBind = EvtSrc(2)
	// EventSrcAPI api event
	EventSrcAPI = EvtSrc(3)
	// EventSrcRouting routing event
	EventSrcRouting = EvtSrc(4)
	// EventSrcProxy routing event
	EventSrcProxy = EvtSrc(5)
	// EventSrcPlugin plugin event
	EventSrcPlugin = EvtSrc(6)
	// EventSrcApplyPlugin apply plugin event
	EventSrcApplyPlugin = EvtSrc(7)
)

// Evt event
type Evt struct {
	Src   EvtSrc
	Type  EvtType
	Key   string
	Value interface{}
}

func init() {
	supportSchema["etcd"] = getEtcdStoreFrom
}

// GetStoreFrom returns a store implemention, if not support returns error
func GetStoreFrom(registryAddr, prefix string) (Store, error) {
	u, err := url.Parse(registryAddr)
	if err != nil {
		panic(fmt.Sprintf("parse registry addr failed, errors:%+v", err))
	}

	schema := strings.ToLower(u.Scheme)
	fn, ok := supportSchema[schema]
	if ok {
		return fn(u.Host, prefix)
	}

	return nil, fmt.Errorf("not support: %s", registryAddr)
}

func getEtcdStoreFrom(addr, prefix string) (Store, error) {
	var addrs []string
	values := strings.Split(addr, ",")

	for _, value := range values {
		addrs = append(addrs, fmt.Sprintf("http://%s", value))
	}

	return NewEtcdStore(addrs, prefix)
}

// Store store interface
type Store interface {
	Raw() interface{}

	AddBind(bind *metapb.Bind) error
	RemoveBind(bind *metapb.Bind) error
	RemoveClusterBind(id uint64) error
	GetBindServers(id uint64) ([]uint64, error)

	PutCluster(cluster *metapb.Cluster) (uint64, error)
	RemoveCluster(id uint64) error
	GetClusters(limit int64, fn func(interface{}) error) error
	GetCluster(id uint64) (*metapb.Cluster, error)

	PutServer(svr *metapb.Server) (uint64, error)
	RemoveServer(id uint64) error
	GetServers(limit int64, fn func(interface{}) error) error
	GetServer(id uint64) (*metapb.Server, error)

	PutAPI(api *metapb.API) (uint64, error)
	RemoveAPI(id uint64) error
	GetAPIs(limit int64, fn func(interface{}) error) error
	GetAPI(id uint64) (*metapb.API, error)

	PutRouting(routing *metapb.Routing) (uint64, error)
	RemoveRouting(id uint64) error
	GetRoutings(limit int64, fn func(interface{}) error) error
	GetRouting(id uint64) (*metapb.Routing, error)

	PutPlugin(plugin *metapb.Plugin) (uint64, error)
	RemovePlugin(id uint64) error
	GetPlugins(limit int64, fn func(interface{}) error) error
	GetPlugin(id uint64) (*metapb.Plugin, error)
	ApplyPlugins(applied *metapb.AppliedPlugins) error
	GetAppliedPlugins() (*metapb.AppliedPlugins, error)

	RegistryProxy(proxy *metapb.Proxy, ttl int64) error
	GetProxies(limit int64, fn func(*metapb.Proxy) error) error

	Watch(evtCh chan *Evt, stopCh chan bool) error

	Clean() error
	SetID(id uint64) error
	BackupTo(to string) error
	Batch(batch *rpcpb.BatchReq) (*rpcpb.BatchRsp, error)
	System() (*metapb.System, error)
}

func getKey(prefix string, id uint64) string {
	return fmt.Sprintf("%s/%020d", prefix, id)
}

func getAddrKey(prefix string, addr string) string {
	return fmt.Sprintf("%s/%s", prefix, util.GetAddrFormat(addr))
}
