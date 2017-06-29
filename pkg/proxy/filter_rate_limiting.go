package proxy

import (
	"errors"
	"net/http"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/log"
)

var (
	// ErrTraffixLimited traffic limit
	ErrTraffixLimited = errors.New("traffic limit")
)

// RateLimitingFilter RateLimitingFilter
type RateLimitingFilter struct {
	filter.BaseFilter
}

func newRateLimitingFilter() filter.Filter {
	return &RateLimitingFilter{}
}

// Name return name of this filter
func (f RateLimitingFilter) Name() string {
	return FilterRateLimiting
}

// Pre execute before proxy
func (f RateLimitingFilter) Pre(c filter.Context) (statusCode int, err error) {
	requestCounts := c.GetRecentlyRequestCount(1)

	if requestCounts >= c.GetMaxQPS() {
		log.Warnf("filter: qps: %d, last 1 secs: %d", c.GetMaxQPS(), requestCounts)
		c.RecordMetricsForReject()
		return http.StatusServiceUnavailable, ErrTraffixLimited
	}

	return f.BaseFilter.Pre(c)
}
