package proxy

import "github.com/fagongzi/gateway/conf"

// XForwardForFilter XForwardForFilter
type XForwardForFilter struct {
	baseFilter
	config *conf.Conf
	proxy  *Proxy
}

func newXForwardForFilter(config *conf.Conf, proxy *Proxy) Filter {
	return XForwardForFilter{
		config: config,
		proxy:  proxy,
	}
}

// Name return name of this filter
func (f XForwardForFilter) Name() string {
	return FilterXForward
}

// Pre execute before proxy
func (f XForwardForFilter) Pre(c *filterContext) (statusCode int, err error) {
	c.outreq.Header.Add("X-Forwarded-For", c.ctx.RemoteIP().String())
	return f.baseFilter.Pre(c)
}
