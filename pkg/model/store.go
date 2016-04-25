package model

type EvtType int
type EvtSrc int

const (
	EVT_TYPE_NEW    = EvtType(0)
	EVT_TYPE_UPDATE = EvtType(1)
	EVT_TYPE_DELETE = EvtType(2)
)

const (
	EVT_SRC_CLUSTER     = EvtSrc(0)
	EVT_SRC_SERVER      = EvtSrc(1)
	EVT_SRC_BIND        = EvtSrc(2)
	EVT_STC_AGGREGATION = EvtSrc(3)
)

type Evt struct {
	Src   EvtSrc
	Type  EvtType
	Key   string
	Value interface{}
}

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

	Watch(evtCh chan *Evt, stopCh chan bool) error

	Clean() error
	GC() error
}
