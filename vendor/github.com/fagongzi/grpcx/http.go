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

type httpEntrypoint struct {
	path       string
	method     string
	reqFactory func() interface{}
	invoker    func(interface{}, echo.Context) (interface{}, error)
	handler    func(echo.Context) error
}

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

func (s *httpServer) addHTTPEntrypoints(httpEntrypoints ...*httpEntrypoint) {
	for _, ep := range httpEntrypoints {
		invoker := ep
		m := strings.ToUpper(ep.method)
		switch m {
		case echo.GET:
			s.server.GET(ep.path, func(c echo.Context) error {
				return s.handleHTTP(c, invoker)
			})
		case echo.PUT:
			s.server.PUT(ep.path, func(c echo.Context) error {
				return s.handleHTTP(c, invoker)
			})
		case echo.DELETE:
			s.server.DELETE(ep.path, func(c echo.Context) error {
				return s.handleHTTP(c, invoker)
			})
		case echo.POST:
			s.server.POST(ep.path, func(c echo.Context) error {
				return s.handleHTTP(c, invoker)
			})
		}
	}
}

func (s *httpServer) handleHTTP(c echo.Context, ep *httpEntrypoint) error {
	if ep.handler == nil &&
		ep.invoker == nil &&
		ep.reqFactory == nil {
		return c.NoContent(http.StatusServiceUnavailable)
	}

	if ep.handler != nil {
		return ep.handler(c)
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

	rsp, err := ep.invoker(req, c)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	result, err := json.Marshal(rsp)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSONBlob(http.StatusOK, result)
}
