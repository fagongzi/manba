package proxy

import (
	"strings"

	"github.com/valyala/fasthttp"
)

// GetRealClientIP get read client ip
func GetRealClientIP(ctx *fasthttp.RequestCtx) string {
	xforward := ctx.Request.Header.Peek("X-Forwarded-For")
	if nil == xforward {
		return strings.SplitN(ctx.RemoteAddr().String(), ":", 2)[0]
	}

	return strings.SplitN(string(xforward), ",", 2)[0]
}
