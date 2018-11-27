package proxy

import (
	"time"

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

	apiResponseHistogramVec = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "gateway",
			Subsystem: "proxy",
			Name:      "api_response_duration_seconds",
			Help:      "Bucketed histogram of api response time duration",
			Buckets:   prometheus.ExponentialBuckets(0.0005, 2.0, 20),
		}, []string{"name"})
)

func init() {
	prometheus.Register(apiRequestCounterVec)
	prometheus.Register(apiResponseHistogramVec)
}

func (p *Proxy) postRequest(api *apiRuntime, dispatches []*dispathNode, startAt time.Time) {
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
		observeAPIResponse(api.meta.Name, startAt)
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

func observeAPIResponse(name string, startAt time.Time) {
	now := time.Now()
	apiResponseHistogramVec.WithLabelValues(name).Observe(now.Sub(startAt).Seconds())
}
