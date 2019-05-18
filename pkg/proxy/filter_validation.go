package proxy

import (
	"errors"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/valyala/fasthttp"
)

var (
	// ErrValidationFailure validation failure
	ErrValidationFailure = errors.New("request validation failure")
)

// ValidationFilter validation request
type ValidationFilter struct {
	filter.BaseFilter
}

func newValidationFilter() filter.Filter {
	return &ValidationFilter{}
}

// Init init filter
func (f *ValidationFilter) Init(cfg string) error {
	return nil
}

// Name return name of this filter
func (f *ValidationFilter) Name() string {
	return FilterValidation
}

// Pre pre filter, before proxy reuqest
func (f *ValidationFilter) Pre(c filter.Context) (statusCode int, err error) {
	if c.(*proxyContext).validateRequest() {
		return f.BaseFilter.Pre(c)
	}

	return fasthttp.StatusBadRequest, ErrValidationFailure
}
