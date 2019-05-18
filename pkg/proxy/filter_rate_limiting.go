package proxy

import (
	"errors"
	"net/http"

	"github.com/fagongzi/gateway/pkg/filter"
)

var (
	errOverLimit = errors.New("too many requests")
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
	if !c.(*proxyContext).rateLimiter().do(1) {
		return http.StatusTooManyRequests, errOverLimit
	}

	return f.BaseFilter.Pre(c)
}
