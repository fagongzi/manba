package proxy

import (
	"time"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/fagongzi/log"
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
	log.SetLevelByString(req.Level)

	rsp.Code = 0
	return nil
}

// AddAnalysisPoint add a point to analysis
func (m *Manager) AddAnalysisPoint(req model.AddAnalysisPointReq, rsp *model.AddAnalysisPointRsp) error {
	m.proxy.dispatcher.analysiser.AddRecentCount(req.Addr, time.Second*time.Duration(req.Secs))

	rsp.Code = 0
	return nil
}

// GetAnalysisPoint return analysis point data
func (m *Manager) GetAnalysisPoint(req model.GetAnalysisPointReq, rsp *model.GetAnalysisPointRsp) error {
	analysisor := m.proxy.dispatcher.analysiser

	rsp.Code = 0
	rsp.RequestCount = analysisor.GetRecentlyRequestCount(req.Addr, time.Second*time.Duration(req.Secs))
	rsp.RequestSuccessedCount = analysisor.GetRecentlyRequestSuccessedCount(req.Addr, time.Second*time.Duration(req.Secs))
	rsp.RejectCount = analysisor.GetRecentlyRejectCount(req.Addr, time.Second*time.Duration(req.Secs))
	rsp.QPS = analysisor.GetQPS(req.Addr, time.Second*time.Duration(req.Secs))

	rsp.RequestFailureCount = analysisor.GetRecentlyRequestFailureCount(req.Addr, time.Second*time.Duration(req.Secs))
	rsp.ContinuousFailureCount = analysisor.GetContinuousFailureCount(req.Addr)

	rsp.Max = analysisor.GetRecentlyMax(req.Addr, time.Second*time.Duration(req.Secs))
	rsp.Min = analysisor.GetRecentlyMin(req.Addr, time.Second*time.Duration(req.Secs))
	rsp.Avg = analysisor.GetRecentlyAvg(req.Addr, time.Second*time.Duration(req.Secs))

	return nil
}
