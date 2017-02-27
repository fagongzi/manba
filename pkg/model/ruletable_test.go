package model

import (
	"net/http"
	"testing"
	"time"

	"github.com/labstack/echo"
	sd "github.com/labstack/echo/engine/standard"
)

const (
	etcdAddr   = "http://192.168.70.13:2379"
	etcdPrefix = "/gateway2"
)

var (
	serverAddr    = "127.0.0.1:12345"
	apiURL        = "/api/test"
	apiMethod     = "GET"
	checkDuration = 3
	checkTimeout  = 2
	clusterName   = "app"
	lbName        = "ROUNDROBIN"
	sleep         = false
)

var rt *RouteTable

func createRouteTable(t *testing.T) {
	store, err := NewEtcdStore([]string{etcdAddr}, etcdPrefix)

	if nil != err {
		t.Fatalf("create etcd store err.addr:<%s>", err)
	}

	store.Clean()

	rt = NewRouteTable(nil, store)
	time.Sleep(time.Second * 1)
}

func createLocalServer() {
	e := echo.New()

	e.Get("/check", func() echo.HandlerFunc {
		return func(c echo.Context) error {
			if sleep {
				time.Sleep(time.Second * time.Duration(checkTimeout+1))
			}

			return c.String(http.StatusOK, "OK")
		}
	}())

	e.Run(sd.New(serverAddr))
}

func waitNotify() {
	time.Sleep(time.Second * 1)
}

func TestCreateRouteTable(t *testing.T) {
	createRouteTable(t)
}

func TestEtcdWatchNewServer(t *testing.T) {
	go createLocalServer()

	server := &Server{
		Schema:                    "http",
		Addr:                      serverAddr,
		CheckPath:                 "/check",
		CheckDuration:             checkDuration,
		CheckTimeout:              checkTimeout,
		MaxQPS:                    1500,
		HalfToOpenSeconds:         10,
		HalfTrafficRate:           10,
		OpenToCloseCollectSeconds: 1,
		OpenToCloseFailureRate:    100,
	}

	err := rt.store.SaveServer(server)

	if nil != err {
		t.Error("add server err.")
		return
	}

	waitNotify()

	if len(rt.svrs) != 1 {
		t.Errorf("expect:<1>, acture:<%d>", len(rt.svrs))
		return
	}

	if rt.svrs[serverAddr].lock == nil {
		t.Error("server init error.")
		return
	}
}

func TestServerCheckOk(t *testing.T) {
	time.Sleep(time.Second * time.Duration(checkDuration))

	if rt.svrs[serverAddr].Status == Down {
		t.Errorf("status check ok err.expect:<UP>, acture:<%v>", Down)
	}
}

func TestServerCheckTimeout(t *testing.T) {
	defer func() {
		sleep = false
	}()

	sleep = true
	time.Sleep(time.Second * time.Duration(checkDuration*2+1)) // 等待两个周期

	if rt.svrs[serverAddr].Status == Up {
		t.Errorf("status check timeout err.expect:<DOWN>, acture:<%v>", Up)
		return
	}
}

func TestServerCheckTimeoutRecovery(t *testing.T) {
	time.Sleep(time.Second * time.Duration(checkDuration*2+1)) // 等待两个周期

	if rt.svrs[serverAddr].Status == Down {
		t.Errorf("status check timeout recovery err.expect:<UP>, acture:<%v>", Up)
		return
	}
}

func TestEtcdWatchNewCluster(t *testing.T) {
	cluster := &Cluster{
		Name:   clusterName,
		LbName: lbName,
	}

	err := rt.store.SaveCluster(cluster)

	if nil != err {
		t.Error("add cluster err.")
		return
	}

	waitNotify()

	if len(rt.clusters) == 1 {
		return
	}

	t.Errorf("expect:<1>, acture:<%d>", len(rt.clusters))
}

func TestEtcdWatchNewBind(t *testing.T) {
	bind := &Bind{
		ClusterName: clusterName,
		ServerAddr:  serverAddr,
	}

	err := rt.store.SaveBind(bind)

	if nil != err {
		t.Error("add cluster err.")
		return
	}

	waitNotify()

	if len(rt.mapping) == 1 {
		return
	}

	t.Errorf("expect:<1>, acture:<%d>. %+v", len(rt.mapping), rt.mapping)
}

func TestEtcdWatchNewAPI(t *testing.T) {
	n := &Node{
		AttrName:    "test",
		ClusterName: clusterName,
	}

	err := rt.store.SaveAPI(&API{
		URL:    apiURL,
		Method: apiMethod,
		Nodes:  []*Node{n},
	})

	if nil != err {
		t.Error("add api err.")
		return
	}

	waitNotify()

	if len(rt.apis) == 1 {
		return
	}

	t.Errorf("expect:<1>, acture:<%d>", len(rt.apis))
}

func TestEtcdWatchUpdateServer(t *testing.T) {
	server := &Server{
		Schema:                    "http",
		Addr:                      serverAddr,
		CheckPath:                 "/check",
		CheckDuration:             checkDuration,
		CheckTimeout:              checkTimeout * 2,
		MaxQPS:                    3000,
		HalfToOpenSeconds:         100,
		HalfTrafficRate:           30,
		OpenToCloseCollectSeconds: 1,
		OpenToCloseFailureRate:    100,
	}

	err := rt.store.UpdateServer(server)

	if nil != err {
		t.Error("update server err.")
		return
	}

	waitNotify()

	svr := rt.svrs[serverAddr]

	if svr.MaxQPS != server.MaxQPS {
		t.Errorf("MaxQPS expect:<%d>, acture:<%d>. ", server.MaxQPS, svr.MaxQPS)
		return
	}

	if svr.HalfToOpenSeconds != server.HalfToOpenSeconds {
		t.Errorf("HalfToOpen expect:<%d>, acture:<%d>. ", server.HalfToOpenSeconds, svr.HalfToOpenSeconds)
		return
	}

	if svr.HalfTrafficRate != server.HalfTrafficRate {
		t.Errorf("HalfTrafficRate expect:<%d>, acture:<%d>. ", server.HalfTrafficRate, svr.HalfTrafficRate)
		return
	}

	if svr.OpenToCloseCollectSeconds != server.OpenToCloseCollectSeconds {
		t.Errorf("OpenToCloseCollectSeconds expect:<%d>, acture:<%d>. ", server.OpenToCloseCollectSeconds, svr.OpenToCloseCollectSeconds)
		return
	}

	if svr.OpenToCloseFailureRate != server.OpenToCloseFailureRate {
		t.Errorf("OpenToCloseFailureRate expect:<%d>, acture:<%d>. ", server.OpenToCloseFailureRate, svr.OpenToCloseFailureRate)
		return
	}

	if svr.CheckTimeout == server.CheckTimeout {
		t.Errorf("CheckTimeout expect:<%d>, acture:<%d>. ", svr.CheckTimeout, server.CheckTimeout)
		return
	}
}

func TestEtcdWatchUpdateCluster(t *testing.T) {
	cluster := &Cluster{
		Name:   clusterName,
		LbName: lbName,
	}

	err := rt.store.UpdateCluster(cluster)

	if nil != err {
		t.Error("update cluster err.")
		return
	}

	waitNotify()

	existCluster := rt.clusters[clusterName]

	if existCluster.LbName != cluster.LbName {
		t.Errorf("LbName expect:<%s>, acture:<%s>. ", cluster.LbName, existCluster.LbName)
		return
	}
}

func TestEtcdWatchUpdateAPI(t *testing.T) {
	n := &Node{
		AttrName:    "test",
		ClusterName: clusterName,
	}

	n2 := &Node{
		AttrName:    "tes2t",
		ClusterName: clusterName,
	}

	api := &API{
		URL:    apiURL,
		Method: apiMethod,
		Nodes:  []*Node{n, n2},
	}

	err := rt.store.UpdateAPI(api)

	if nil != err {
		t.Error("update api err.")
		return
	}

	waitNotify()

	existAPI, _ := rt.apis[getAPIKey(api.URL, api.Method)]

	if len(existAPI.Nodes) != len(api.Nodes) {
		t.Errorf("Nodes expect:<%d>, acture:<%d>. ", len(existAPI.Nodes), len(api.Nodes))
		return
	}
}

func TestEtcdWatchDeleteCluster(t *testing.T) {
	rt.store.UnBind(&Bind{
		ClusterName: clusterName,
		ServerAddr:  serverAddr,
	})

	err := rt.store.DeleteCluster(clusterName)

	if nil != err {
		t.Error("delete cluster err.", err)
		return
	}

	waitNotify()

	if len(rt.clusters) != 0 {
		t.Errorf("clusters expect:<0>, acture:<%d>", len(rt.clusters))
		return
	}

	banded, _ := rt.mapping[serverAddr]

	if len(banded) != 0 {
		t.Errorf("banded expect:<0>, acture:<%d>", len(banded))
		return
	}
}

func TestEtcdWatchDeleteServer(t *testing.T) {
	rt.store.UnBind(&Bind{
		ClusterName: clusterName,
		ServerAddr:  serverAddr,
	})

	err := rt.store.DeleteServer(serverAddr)

	if nil != err {
		t.Error("delete server err.", err)
		return
	}

	waitNotify()

	if len(rt.svrs) != 0 {
		t.Errorf("svrs expect:<0>, acture:<%d>", len(rt.svrs))
		return
	}

	if len(rt.mapping) != 0 {
		t.Errorf("mapping expect:<0>, acture:<%d>", len(rt.mapping))
		return
	}
}

func TestEtcdWatchDeleteAPI(t *testing.T) {
	err := rt.store.DeleteAPI(apiURL, apiMethod)

	if nil != err {
		t.Error("delete api err.")
		return
	}

	waitNotify()

	if len(rt.apis) != 0 {
		t.Errorf("apis expect:<0>, acture:<%d>", len(rt.apis))
		return
	}
}

func TestEtcdWatchNewRouting(t *testing.T) {
	r, err := NewRouting(`desc = "test"; deadline = 100; rule = ["$query_abc == 10", "$query_123 == 20"];`, clusterName, "")

	if nil != err {
		t.Error("add routing err.")
		return
	}

	err = rt.store.SaveRouting(r)

	if nil != err {
		t.Error("add routing err.")
		return
	}

	waitNotify()

	if len(rt.routings) == 1 {
		delete(rt.routings, r.ID)
		return
	}

	t.Errorf("expect:<1>, acture:<%d>", len(rt.routings))
}

func TestEtcdWatchDeleteRouting(t *testing.T) {
	r, err := NewRouting(`desc = "test"; deadline = 3; rule = ["$query_abc == 10", "$query_123 == 20"];`, clusterName, "")

	if nil != err {
		t.Error("add routing err.")
		return
	}

	err = rt.store.SaveRouting(r)

	if nil != err {
		t.Error("add routing err.")
		return
	}

	time.Sleep(time.Second * 30)

	if len(rt.routings) == 0 {
		return
	}

	t.Errorf("expect:<0>, acture:<%d>", len(rt.routings))
}
