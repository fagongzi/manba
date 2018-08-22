package proxy

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/valyala/fasthttp"
)

const (
	typeRequestAll     = "all"
	typeRequestFail    = "fail"
	typeRequestSucceed = "succeed"
	typeRequestLimit   = "limit"
	typeRequestReject  = "reject"
)

var (
	apiRequestCounterVec = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "gateway",
			Subsystem: "proxy",
			Name:      "api_request_total",
			Help:      "Total number of request made.",
		}, []string{"name", "type"})
)

func init() {
	prometheus.Register(apiRequestCounterVec)
}

func (p *Proxy) postRequest(api *apiRuntime, dispatches []*dispathNode) {
	doMetrics := true
	for _, dn := range dispatches {
		if doMetrics &&
			(dn.err == ErrCircuitClose || dn.err == ErrBlacklist || dn.err == ErrWhitelist) {
			incrRequestReject(api.meta.Name)
			doMetrics = false
		} else if doMetrics && dn.err == ErrCircuitHalfLimited {
			incrRequestLimit(api.meta.Name)
			doMetrics = false
		} else if doMetrics && dn.err != nil {
			incrRequestFailed(api.meta.Name)
			doMetrics = false
		} else if doMetrics && dn.code >= fasthttp.StatusBadRequest {
			incrRequestFailed(api.meta.Name)
			doMetrics = false
		}

		releaseDispathNode(dn)
	}

	if doMetrics {
		incrRequestSucceed(api.meta.Name)
	}
}

func incrRequest(name string) {
	apiRequestCounterVec.WithLabelValues(name, typeRequestAll).Inc()
}

func incrRequestFailed(name string) {
	apiRequestCounterVec.WithLabelValues(name, typeRequestFail).Inc()
}

func incrRequestSucceed(name string) {
	apiRequestCounterVec.WithLabelValues(name, typeRequestSucceed).Inc()
}

func incrRequestLimit(name string) {
	apiRequestCounterVec.WithLabelValues(name, typeRequestLimit).Inc()
}

func incrRequestReject(name string) {
	apiRequestCounterVec.WithLabelValues(name, typeRequestReject).Inc()
}
