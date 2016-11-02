package model

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
