package proxy

import (
	"errors"
	"net/http"

	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/gateway/pkg/conf"
)

var (
	// ErrTraffixLimited traffic limit
	ErrTraffixLimited = errors.New("traffic limit")
)

// RateLimitingFilter RateLimitingFilter
type RateLimitingFilter struct {
	baseFilter
	config *conf.Conf
	proxy  *Proxy
}

func newRateLimitingFilter(config *conf.Conf, proxy *Proxy) Filter {
	return RateLimitingFilter{
		config: config,
		proxy:  proxy,
	}
}

// Name return name of this filter
func (f RateLimitingFilter) Name() string {
	return FilterRateLimiting
}

// Pre execute before proxy
func (f RateLimitingFilter) Pre(c *filterContext) (statusCode int, err error) {
	requestCounts := c.rb.GetAnalysis().GetRecentlyRequestCount(c.result.Svr.Addr, 1)

	if requestCounts >= c.result.Svr.MaxQPS {
		log.Warnf("qps: %d, last 1 secs: %d", c.result.Svr.MaxQPS, requestCounts)
		c.rb.GetAnalysis().Reject(c.result.Svr.Addr)
		return http.StatusServiceUnavailable, ErrTraffixLimited
	}

	return f.baseFilter.Pre(c)
}
