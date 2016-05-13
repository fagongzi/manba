package proxy

import (
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/gateway/conf"
	"time"
)

// record the http access log
// log format: $remoteip "$method $path HTTP/$proto" $code "$agent" $svr $cost

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

func (self AccessFilter) Name() string {
	return FILTER_HTTP_ACCESS
}

func (self AccessFilter) Post(c *filterContext) (statusCode int, err error) {
	cost := (c.endAt - c.startAt)

	log.Infof("%s \"%s %s %s\" %d \"%s\" %s %s",
		c.req.RemoteAddr,
		c.req.Method,
		c.outreq.RequestURI,
		c.req.Proto,
		c.result.Res.StatusCode,
		c.req.UserAgent(),
		c.result.Svr.Addr,
		time.Duration(cost))

	return self.baseFilter.Post(c)
}
