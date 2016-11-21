package proxy

import (
	"errors"

	"github.com/fagongzi/gateway/pkg/conf"
	"github.com/valyala/fasthttp"
)

var (
	// ErrWhitelist target ip not in in white list
	ErrWhitelist = errors.New("Err, target ip not in in white list")
)

// WhiteListFilter whitelist filter
type WhiteListFilter struct {
	baseFilter
	proxy *Proxy
}

func newWhiteListFilter(config *conf.Conf, proxy *Proxy) Filter {
	return WhiteListFilter{
		proxy: proxy,
	}
}

// Name return name of this filter
func (f WhiteListFilter) Name() string {
	return FilterWhiteList
}

// Pre execute before proxy
func (f WhiteListFilter) Pre(c *filterContext) (statusCode int, err error) {
	if !c.result.API.AccessCheckWhitelist(GetRealClientIP(c.ctx)) {
		return fasthttp.StatusForbidden, ErrWhitelist
	}

	return f.baseFilter.Pre(c)
}
