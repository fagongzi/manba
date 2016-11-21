package proxy

import (
	"errors"
	"strings"

	"github.com/fagongzi/gateway/pkg/conf"
)

var (
	// ErrKnownFilter known filter error
	ErrKnownFilter = errors.New("unknow filter")
)

const (
	// FilterHTTPAccess access log filter
	FilterHTTPAccess = "HTTP-ACCESS"
	// FilterHeader header filter
	FilterHeader = "HEAD" // process header fiter
	// FilterXForward xforward fiter
	FilterXForward = "XFORWARD"
	// FilterBlackList blacklist filter
	FilterBlackList = "BLACKLIST"
	// FilterWhiteList whitelist filter
	FilterWhiteList = "WHITELIST"
	// FilterAnalysis analysis filter
	FilterAnalysis = "ANALYSIS"
	// FilterRateLimiting limit filter
	FilterRateLimiting = "RATE-LIMITING"
	// FilterCircuitBreake circuit breake filter
	FilterCircuitBreake = "CIRCUIT-BREAKE"
	// FilterValidation validation request filter
	FilterValidation = "VALIDATION"
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
	case FilterWhiteList:
		return newWhiteListFilter(config, proxy), nil
	case FilterRateLimiting:
		return newRateLimitingFilter(config, proxy), nil
	case FilterCircuitBreake:
		return newCircuitBreakeFilter(config, proxy), nil
	case FilterValidation:
		return newValidationFilter(config, proxy), nil
	default:
		return nil, ErrKnownFilter
	}
}
