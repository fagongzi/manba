package proxy

import (
	"errors"

	"github.com/fagongzi/gateway/pkg/conf"
	"github.com/valyala/fasthttp"
)

var (
	// ErrBlacklist target ip in black list
	ErrBlacklist = errors.New("Err, target ip in black list")
)

// BlackListFilter blacklist filter
type BlackListFilter struct {
	baseFilter
	proxy *Proxy
}

func newBlackListFilter(config *conf.Conf, proxy *Proxy) Filter {
	return BlackListFilter{
		proxy: proxy,
	}
}

// Name return name of this filter
func (f BlackListFilter) Name() string {
	return FilterBlackList
}

// Pre execute before proxy
func (f BlackListFilter) Pre(c *filterContext) (statusCode int, err error) {
	if c.result.API.AccessCheckBlacklist(GetRealClientIP(c.ctx)) {
		return fasthttp.StatusForbidden, ErrBlacklist
	}

	return f.baseFilter.Pre(c)
}
