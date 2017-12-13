package proxy

import (
	"errors"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/valyala/fasthttp"
)

var (
	// ErrBlacklist target ip in black list
	ErrBlacklist = errors.New("Err, target ip in black list")
)

// BlackListFilter blacklist filter
type BlackListFilter struct {
	filter.BaseFilter
}

func newBlackListFilter() filter.Filter {
	return &BlackListFilter{}
}

// Name return name of this filter
func (f BlackListFilter) Name() string {
	return FilterBlackList
}

// Pre execute before proxy
func (f BlackListFilter) Pre(c filter.Context) (statusCode int, err error) {
	if c.AllowWithBlacklist(GetRealClientIP(c.OriginRequest())) {
		return fasthttp.StatusForbidden, ErrBlacklist
	}

	return f.BaseFilter.Pre(c)
}
