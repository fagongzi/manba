package store

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/util/task"
	"github.com/toolkits/net"
)

var (
	// TICKER ticket
	TICKER = time.Second * 3
	// TTL timeout
	TTL = int64(5)
)

var (
	supportSchema = make(map[string]func(string, string, *task.Runner) (Store, error))
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
func GetStoreFrom(registryAddr, prefix string, runner *task.Runner) (Store, error) {
	u, err := url.Parse(registryAddr)
	if err != nil {
		panic(fmt.Sprintf("parse registry addr failed, errors:%+v", err))
	}

	schema := strings.ToLower(u.Scheme)
	fn, ok := supportSchema[schema]
	if ok {
		return fn(u.Host, prefix, runner)
	}

	return nil, fmt.Errorf("not support: %s", registryAddr)
}

func getEtcdStoreFrom(addr, prefix string, runner *task.Runner) (Store, error) {
	var addrs []string
	values := strings.Split(addr, ",")

	for _, value := range values {
		addrs = append(addrs, fmt.Sprintf("http://%s", value))
	}

	return NewEtcdStore(addrs, prefix, runner)
}

func convertIP(addr string) string {
	if strings.HasPrefix(addr, ":") {
		ips, err := net.IntranetIP()

		if err == nil {
			addr = strings.Replace(addr, ":", fmt.Sprintf("%s:", ips[0]), 1)
		}
	}

	return addr
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

	Watch(evtCh chan *Evt, stopCh chan bool) error

	Clean() error
}

func getKey(prefix string, id uint64) string {
	return fmt.Sprintf("%s/%020d", prefix, id)
}
