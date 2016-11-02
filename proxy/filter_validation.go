package proxy

import (
	"errors"

	"github.com/fagongzi/gateway/conf"
	"github.com/valyala/fasthttp"
)

var (
	ErrValidationFailure = errors.New("request validation failure")
)

// ValidationFilter validation request
type ValidationFilter struct {
	baseFilter
	config *conf.Conf
	proxy  *Proxy
}

func newValidationFilter(config *conf.Conf, proxy *Proxy) Filter {
	return ValidationFilter{
		config: config,
		proxy:  proxy,
	}
}

// Name return name of this filter
func (v ValidationFilter) Name() string {
	return FilterValidation
}

func (v ValidationFilter) Pre(c *filterContext) (statusCode int, err error) {
	if c.result.Node.Validate(c.outreq) {
		return v.baseFilter.Pre(c)
	}

	return fasthttp.StatusBadRequest, ErrValidationFailure
}
