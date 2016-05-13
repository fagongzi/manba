package proxy

import (
	"fmt"
	"github.com/fagongzi/gateway/conf"
	"strings"
)

// 变量
// $remote_ip    客户端ip地址
// $remote_port  客户端port
// $proxy_ip     proxy地址
// $proxy_port   proxy端口

var (
	VAR_CLIENT_IP   = "$CLIENT_IP"
	VAR_CLIENT_PORT = "$CLIENT_PORT"

	VAR_PROXY_IP   = "$PROXY_IP"
	VAR_PROXY_PORT = "$PROXY_PORT"

	VAR_PREFIX_COOKIE = "$COOKIE_"
	VAR_PREFIX_HEAD   = "$HEAD_"
)

func (self HeadersFilter) setRuntimeVals(c *filterContext) {
	if self.config.EnableRuntimeVal {
		if self.config.EnableHeadVal {
			for k, vv := range c.outreq.Header {
				name := fmt.Sprintf("%s%s", VAR_PREFIX_HEAD, k)

				for _, v := range vv {
					c.runtimeVar[strings.ToUpper(name)] = v
				}
			}
		}

		if self.config.EnableCookieVal {
			cookies := c.outreq.Cookies()
			l := len(cookies)
			for i := 0; i < l; i++ {
				cookie := cookies[i]
				name := fmt.Sprintf("%s%s", VAR_PREFIX_COOKIE, cookie.Name)
				c.runtimeVar[strings.ToUpper(name)] = cookie.Value
			}
		}
	}
}

func setRuntimeVal(config *conf.Conf, c *filterContext) {
	if config.EnableRuntimeVal {
		client := strings.Split(c.outreq.RemoteAddr, ":")
		proxy := strings.Split(config.Addr, ":")

		c.runtimeVar[VAR_CLIENT_IP] = client[0]
		c.runtimeVar[VAR_CLIENT_PORT] = client[1]

		c.runtimeVar[VAR_PROXY_IP] = proxy[0]
		c.runtimeVar[VAR_PROXY_PORT] = proxy[1]
	}
}
