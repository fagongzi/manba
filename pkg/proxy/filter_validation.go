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

// Name return name of this filter
func (v ValidationFilter) Name() string {
	return FilterValidation
}

// Pre pre filter, before proxy reuqest
func (v ValidationFilter) Pre(c filter.Context) (statusCode int, err error) {
	if c.ValidateRequest() {
		return v.BaseFilter.Pre(c)
	}

	return fasthttp.StatusBadRequest, ErrValidationFailure
}
