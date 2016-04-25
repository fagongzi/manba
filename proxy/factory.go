package proxy

import (
	"errors"
	"github.com/fagongzi/gateway/conf"
	"strings"
)

var (
	ERR_KNOWN_FILTER = errors.New("unknow filter")
)

const (
	FILTER_HTTP_ACCESS = "HTTP-ACCESS" // 日志
	FILTER_HEAD        = "HEAD"        // 处理head
	FILTER_XFORWARD    = "XFORWARD"    // xforward
	FILTER_BLACKLIST   = "BLACKLIST"   // 黑名单
	// FILTER_WHITELIST      = "WHITELIST"      // 白名单
	FILTER_ANALYSIS       = "ANALYSIS"       // 分析数据
	FILTER_RATE_LIMITING  = "RATE-LIMITING"  // 限流
	FILTER_CIRCUIT_BREAKE = "CIRCUIT-BREAKE" // 断路保护
)

func newFilter(name string, config *conf.Conf, proxy *Proxy) (Filter, error) {
	input := strings.ToUpper(name)

	switch input {
	case FILTER_HTTP_ACCESS:
		return newAccessFilter(config, proxy), nil
	case FILTER_HEAD:
		return newHeadersFilter(config, proxy), nil
	case FILTER_XFORWARD:
		return newXForwardForFilter(config, proxy), nil
	case FILTER_ANALYSIS:
		return newAnalysisFilter(config, proxy), nil
	case FILTER_BLACKLIST:
		return newBlackListFilter(config, proxy), nil
	case FILTER_RATE_LIMITING:
		return newRateLimitingFilter(config, proxy), nil
	case FILTER_CIRCUIT_BREAKE:
		return newCircuitBreakeFilter(config, proxy), nil
	default:
		return nil, ERR_KNOWN_FILTER
	}
}
