package proxy

import (
	"github.com/prometheus/client_golang/prometheus"
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
