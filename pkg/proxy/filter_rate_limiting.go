package proxy

import (
	"net/http"

	"github.com/fagongzi/gateway/pkg/filter"
	"golang.org/x/net/context"
)

// RateLimitingFilter RateLimitingFilter
type RateLimitingFilter struct {
	filter.BaseFilter
}

func newRateLimitingFilter() filter.Filter {
	return &RateLimitingFilter{}
}

// Init init filter
func (f *RateLimitingFilter) Init(cfg string) error {
	return nil
}

// Name return name of this filter
func (f *RateLimitingFilter) Name() string {
	return FilterRateLimiting
}

// Pre execute before proxy
func (f *RateLimitingFilter) Pre(c filter.Context) (statusCode int, err error) {
	err = c.(*proxyContext).result.dest.limiter.Wait(context.Background())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return f.BaseFilter.Pre(c)
}
