package model

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/toolkits/net"
)

var (
	// TICKER ticket
	TICKER = time.Second * 3
	// TTL timeout
	TTL = uint64(5)
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

func getConsulStoreFrom(addr, prefix string) (Store, error) {
	return NewConsulStore(addr, prefix)
}

func getEtcdStoreFrom(addr, prefix string) (Store, error) {
	var addrs []string
	values := strings.Split(addr, ",")

	for _, value := range values {
		addrs = append(addrs, fmt.Sprintf("http://%s", value))
	}

	return NewEtcdStore(addrs, prefix)
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
	SaveBind(bind *Bind) error
	UnBind(bind *Bind) error
	GetBinds() ([]*Bind, error)

	SaveCluster(cluster *Cluster) error
	UpdateCluster(cluster *Cluster) error
	DeleteCluster(name string) error
	GetClusters() ([]*Cluster, error)
	GetCluster(clusterName string) (*Cluster, error)

	SaveServer(svr *Server) error
	UpdateServer(svr *Server) error
	DeleteServer(addr string) error
	GetServers() ([]*Server, error)
	GetServer(serverAddr string) (*Server, error)

	SaveAPI(api *API) error
	UpdateAPI(api *API) error
	DeleteAPI(url string, method string) error
	GetAPIs() ([]*API, error)
	GetAPI(url string, method string) (*API, error)

	SaveRouting(routing *Routing) error
	GetRoutings() ([]*Routing, error)

	Watch(evtCh chan *Evt, stopCh chan bool) error

	Clean() error
	GC() error
}
