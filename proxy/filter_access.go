package proxy

import (
	"time"

	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/gateway/conf"
)

// AccessFilter record the http access log
// log format: $remoteip "$method $path" $code "$agent" $svr $cost
type AccessFilter struct {
	baseFilter
	config *conf.Conf
	proxy  *Proxy
}

func newAccessFilter(config *conf.Conf, proxy *Proxy) Filter {
	return AccessFilter{
		config: config,
		proxy:  proxy,
	}
}

// Name return name of this filter
func (f AccessFilter) Name() string {
	return FilterHTTPAccess
}

// Post execute after proxy
func (f AccessFilter) Post(c *filterContext) (statusCode int, err error) {
	cost := (c.endAt - c.startAt)

	log.Infof("%s %s \"%s\" %d \"%s\" %s %s",
		c.ctx.RemoteIP().String(),
		c.ctx.Method(),
		c.outreq.RequestURI(),
		c.result.Res.StatusCode(),
		c.ctx.UserAgent(),
		c.result.Svr.Addr,
		time.Duration(cost))

	return f.baseFilter.Post(c)
}
