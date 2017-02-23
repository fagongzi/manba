package proxy

import (
	"errors"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/valyala/fasthttp"
)

var (
	// ErrWhitelist target ip not in in white list
	ErrWhitelist = errors.New("Err, target ip not in in white list")
)

// WhiteListFilter whitelist filter
type WhiteListFilter struct {
	filter.BaseFilter
}

func newWhiteListFilter() filter.Filter {
	return &WhiteListFilter{}
}

// Name return name of this filter
func (f WhiteListFilter) Name() string {
	return FilterWhiteList
}

// Pre execute before proxy
func (f WhiteListFilter) Pre(c filter.Context) (statusCode int, err error) {
	if !c.InWhitelist(GetRealClientIP(c.GetOriginRequestCtx())) {
		return fasthttp.StatusForbidden, ErrWhitelist
	}

	return f.BaseFilter.Pre(c)
}
