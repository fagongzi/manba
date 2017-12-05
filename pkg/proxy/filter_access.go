package proxy

import (
	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/log"
)

// AccessFilter record the http access log
// log format: $remoteip "$method $path" $code "$agent" $svr $cost
type AccessFilter struct {
	filter.BaseFilter
}

func newAccessFilter() filter.Filter {
	return &AccessFilter{}
}

// Name return name of this filter
func (f AccessFilter) Name() string {
	return FilterHTTPAccess
}

// Post execute after proxy
func (f AccessFilter) Post(c filter.Context) (statusCode int, err error) {
	cost := c.GetEndAt().Sub(c.GetStartAt())

	log.Infof("filter: %s %s \"%s\" %d \"%s\" %s %s",
		GetRealClientIP(c.GetOriginRequestCtx()),
		c.GetOriginRequestCtx().Method(),
		c.GetProxyOuterRequest().RequestURI(),
		c.GetProxyResponse().StatusCode(),
		c.GetOriginRequestCtx().UserAgent(),
		c.GetProxyServerAddr(),
		cost)

	return f.BaseFilter.Post(c)
}
