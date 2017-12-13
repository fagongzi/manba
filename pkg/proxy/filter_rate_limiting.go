package proxy

import (
	"errors"
	"net/http"

	"github.com/fagongzi/gateway/pkg/filter"
	"golang.org/x/net/context"
)

var (
	// ErrTraffixLimited traffic limit
	ErrTraffixLimited = errors.New("traffic limit")
)

// RateLimitingFilter RateLimitingFilter
type RateLimitingFilter struct {
	filter.BaseFilter
}

func newRateLimitingFilter() filter.Filter {
	return &RateLimitingFilter{}
}

// Name return name of this filter
func (f RateLimitingFilter) Name() string {
	return FilterRateLimiting
}

// Pre execute before proxy
func (f RateLimitingFilter) Pre(c filter.Context) (statusCode int, err error) {
	err = c.(*proxyContext).result.api.limiter.Wait(context.Background())
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return f.BaseFilter.Pre(c)
}
