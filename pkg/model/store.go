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
	// EventSrcAggregation aggregation event
	EventSrcAggregation = EvtSrc(3)
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
	GetCluster(clusterName string, withBinded bool) (*Cluster, error)
	GetBindedClusters(serverAddr string) ([]string, error)

	SaveServer(svr *Server) error
	UpdateServer(svr *Server) error
	DeleteServer(addr string) error
	GetServers() ([]*Server, error)
	GetServer(serverAddr string, withBinded bool) (*Server, error)
	GetBindedServers(clusterName string) ([]string, error)

	SaveAggregation(agn *Aggregation) error
	UpdateAggregation(agn *Aggregation) error
	DeleteAggregation(url string) error
	GetAggregations() ([]*Aggregation, error)
	GetAggregation(url string) (*Aggregation, error)

	SaveRouting(routing *Routing) error
	GetRoutings() ([]*Routing, error)

	Watch(evtCh chan *Evt, stopCh chan bool) error

	Clean() error
	GC() error
}
