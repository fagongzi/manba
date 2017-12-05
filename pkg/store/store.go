package store

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/fagongzi/gateway/pkg/model"
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
	supportSchema["consul"] = getConsulStoreFrom
	supportSchema["etcd"] = getEtcdStoreFrom
}

// GetStoreFrom returns a store implemention, if not support returns error
func GetStoreFrom(registryAddr, prefix string, taskRunner *task.Runner) (Store, error) {
	u, err := url.Parse(registryAddr)
	if err != nil {
		panic(fmt.Sprintf("parse registry addr failed, errors:%+v", err))
	}

	schema := strings.ToLower(u.Scheme)
	fn, ok := supportSchema[schema]
	if ok {
		return fn(u.Host, prefix, taskRunner)
	}

	return nil, fmt.Errorf("not support: %s", registryAddr)
}

func getConsulStoreFrom(addr, prefix string, taskRunner *task.Runner) (Store, error) {
	return NewConsulStore(addr, prefix, taskRunner)
}

func getEtcdStoreFrom(addr, prefix string, taskRunner *task.Runner) (Store, error) {
	var addrs []string
	values := strings.Split(addr, ",")

	for _, value := range values {
		addrs = append(addrs, fmt.Sprintf("http://%s", value))
	}

	return NewEtcdStore(addrs, prefix, taskRunner)
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
	SaveBind(bind *model.Bind) error
	UnBind(id string) error
	GetBinds() ([]*model.Bind, error)
	GetBind(id string) (*model.Bind, error)

	SaveCluster(cluster *model.Cluster) error
	UpdateCluster(cluster *model.Cluster) error
	DeleteCluster(id string) error
	GetClusters() ([]*model.Cluster, error)
	GetCluster(id string) (*model.Cluster, error)

	SaveServer(svr *model.Server) error
	UpdateServer(svr *model.Server) error
	DeleteServer(id string) error
	GetServers() ([]*model.Server, error)
	GetServer(id string) (*model.Server, error)

	SaveAPI(api *model.API) error
	UpdateAPI(api *model.API) error
	DeleteAPI(id string) error
	GetAPIs() ([]*model.API, error)
	GetAPI(id string) (*model.API, error)

	SaveRouting(routing *model.Routing) error
	GetRoutings() ([]*model.Routing, error)

	Watch(evtCh chan *Evt, stopCh chan bool) error

	Clean() error
}
