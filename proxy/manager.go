package proxy

import (
	"net"
	"net/rpc"

	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/fagongzi/gateway/pkg/util"
)

// Manager support runtime remote interface
type Manager struct {
	proxy *Proxy
}

func newManager(proxy *Proxy) *Manager {
	return &Manager{proxy: proxy}
}

// SetLogLevel set log level
func (m *Manager) SetLogLevel(req model.SetLogReq, rsp *model.SetLogRsp) error {
	level := util.SetLogLevel(req.Level)
	m.proxy.config.LogLevel = level

	rsp.Code = 0
	return nil
}

// AddAnalysisPoint add a point to analysis
func (m *Manager) AddAnalysisPoint(req model.AddAnalysisPointReq, rsp *model.AddAnalysisPointRsp) error {
	m.proxy.routeTable.GetAnalysis().AddRecentCount(req.Addr, req.Secs)

	rsp.Code = 0
	return nil
}

// GetAnalysisPoint return analysis point data
func (m *Manager) GetAnalysisPoint(req model.GetAnalysisPointReq, rsp *model.GetAnalysisPointRsp) error {
	analysisor := m.proxy.routeTable.GetAnalysis()

	rsp.Code = 0
	rsp.RequestCount = analysisor.GetRecentlyRequestCount(req.Addr, req.Secs)
	rsp.RequestSuccessedCount = analysisor.GetRecentlyRequestSuccessedCount(req.Addr, req.Secs)
	rsp.RejectCount = analysisor.GetRecentlyRejectCount(req.Addr, req.Secs)
	rsp.QPS = analysisor.GetQPS(req.Addr, req.Secs)

	rsp.RequestFailureCount = analysisor.GetRecentlyRequestFailureCount(req.Addr, req.Secs)
	rsp.ContinuousFailureCount = analysisor.GetContinuousFailureCount(req.Addr)

	rsp.Max = analysisor.GetRecentlyMax(req.Addr, req.Secs)
	rsp.Min = analysisor.GetRecentlyMin(req.Addr, req.Secs)
	rsp.Avg = analysisor.GetRecentlyAvg(req.Addr, req.Secs)

	return nil
}

func (p *Proxy) startRPCServer() error {
	tcpAddr, err := net.ResolveTCPAddr("tcp", p.config.MgrAddr)

	if err != nil {
		return err
	}

	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}

	log.Infof("Mgr listen at %s.", p.config.MgrAddr)

	server := rpc.NewServer()

	mgrService := newManager(p)
	server.Register(mgrService)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.ErrorError(err, "Mgr error.")
				continue
			}

			go server.ServeConn(conn)
		}
	}()

	return nil
}
