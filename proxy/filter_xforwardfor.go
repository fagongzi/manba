package proxy

import "github.com/fagongzi/gateway/pkg/filter"

// XForwardForFilter XForwardForFilter
type XForwardForFilter struct {
	filter.BaseFilter
}

func newXForwardForFilter() filter.Filter {
	return &XForwardForFilter{}
}

// Name return name of this filter
func (f XForwardForFilter) Name() string {
	return FilterXForward
}

// Pre execute before proxy
func (f XForwardForFilter) Pre(c filter.Context) (statusCode int, err error) {
	c.GetProxyOuterRequest().Header.Add("X-Forwarded-For", c.GetOriginRequestCtx().RemoteIP().String())
	return f.BaseFilter.Pre(c)
}
