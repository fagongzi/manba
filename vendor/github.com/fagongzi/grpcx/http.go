package grpcx

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fagongzi/log"
	"github.com/labstack/echo"
	md "github.com/labstack/echo/middleware"
)

type httpServer struct {
	addr   string
	server *echo.Echo
}

func newHTTPServer(addr string) *httpServer {
	return &httpServer{
		addr:   addr,
		server: echo.New(),
	}
}

func (s *httpServer) start() error {
	s.server.Use(md.Recover())

	log.Infof("rpc: start a grpc http proxy server at %s", s.addr)
	return s.server.Start(s.addr)
}

func (s *httpServer) stop() error {
	return s.server.Shutdown(context.Background())
}

func (s *httpServer) addService(service Service) {
	for _, ep := range service.opts.httpEntrypoints {
		m := strings.ToUpper(ep.method)
		switch m {
		case echo.GET:
			s.server.GET(ep.path, func(c echo.Context) error {
				return s.handleHTTP(c, ep)
			})
		case echo.PUT:
			s.server.PUT(ep.path, func(c echo.Context) error {
				return s.handleHTTP(c, ep)
			})
		case echo.DELETE:
			s.server.DELETE(ep.path, func(c echo.Context) error {
				return s.handleHTTP(c, ep)
			})
		case echo.POST:
			s.server.POST(ep.path, func(c echo.Context) error {
				return s.handleHTTP(c, ep)
			})
		}
	}
}

func (s *httpServer) handleHTTP(c echo.Context, ep *httpEntrypoint) error {
	if ep.invoker == nil || ep.reqFactory == nil {
		return c.NoContent(http.StatusServiceUnavailable)
	}

	data, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	req := ep.reqFactory()
	if len(data) > 0 {
		err = json.Unmarshal(data, req)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
	}

	rsp, err := ep.invoker(req)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	result, err := json.Marshal(rsp)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSONBlob(http.StatusOK, result)
}
