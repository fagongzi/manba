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

// Init init filter
func (f *AccessFilter) Init(cfg string) error {
	return nil
}

// Name return name of this filter
func (f *AccessFilter) Name() string {
	return FilterHTTPAccess
}

// Post execute after proxy
func (f *AccessFilter) Post(c filter.Context) (statusCode int, err error) {
	cost := c.EndAt().Sub(c.StartAt())

	if log.InfoEnabled() {
		log.Infof("filter: %s %s \"%s\" %d \"%s\" %s %s",
			filter.StringValue(filter.AttrClientRealIP, c),
			c.OriginRequest().Method(),
			c.ForwardRequest().RequestURI(),
			c.Response().StatusCode(),
			c.OriginRequest().UserAgent(),
			c.Server().Addr,
			cost)
	}

	return f.BaseFilter.Post(c)
}
