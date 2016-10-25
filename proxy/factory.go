package proxy

import (
	"errors"
	"strings"

	"github.com/fagongzi/gateway/conf"
)

var (
	// ErrKnownFilter known filter error
	ErrKnownFilter = errors.New("unknow filter")
)

const (
	// FilterHTTPAccess log filter
	FilterHTTPAccess = "HTTP-ACCESS"
	// FilterHeader header filter
	FilterHeader = "HEAD" // 处理head
	// FilterXForward xforward fiter
	FilterXForward = "XFORWARD"
	// FilterBlackList blacklist filter
	FilterBlackList = "BLACKLIST"
	// FilterAnalysis analysis filter
	FilterAnalysis = "ANALYSIS"
	// FilterRateLimiting limit filter
	FilterRateLimiting = "RATE-LIMITING"
	// FilterCircuitBreake circuit breake filter
	FilterCircuitBreake = "CIRCUIT-BREAKE"
)

func newFilter(name string, config *conf.Conf, proxy *Proxy) (Filter, error) {
	input := strings.ToUpper(name)

	switch input {
	case FilterHTTPAccess:
		return newAccessFilter(config, proxy), nil
	case FilterHeader:
		return newHeadersFilter(config, proxy), nil
	case FilterXForward:
		return newXForwardForFilter(config, proxy), nil
	case FilterAnalysis:
		return newAnalysisFilter(config, proxy), nil
	case FilterBlackList:
		return newBlackListFilter(config, proxy), nil
	case FilterRateLimiting:
		return newRateLimitingFilter(config, proxy), nil
	case FilterCircuitBreake:
		return newCircuitBreakeFilter(config, proxy), nil
	default:
		return nil, ErrKnownFilter
	}
}
