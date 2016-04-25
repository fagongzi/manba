package proxy

import (
	"errors"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/gateway/conf"
	"net/http"
)

var (
	ERR_TRAFFIC_LIMITED = errors.New("traffic limit")
)

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

func (self RateLimitingFilter) Name() string {
	return FILTER_RATE_LIMITING
}

func (self RateLimitingFilter) Pre(c *filterContext) (statusCode int, err error) {
	requestCounts := c.rb.GetAnalysis().GetRecentlyRequestCount(c.result.Svr.Addr, 1)

	if requestCounts >= c.result.Svr.MaxQPS {
		log.Warnf("qps: %d, last 1 secs: %d", c.result.Svr.MaxQPS, requestCounts)
		c.rb.GetAnalysis().Reject(c.result.Svr.Addr)
		return http.StatusServiceUnavailable, ERR_TRAFFIC_LIMITED
	}

	return self.baseFilter.Pre(c)
}
