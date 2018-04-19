package grpcx

import (
	"context"

	"github.com/fagongzi/log"
	"github.com/labstack/echo"
	md "github.com/labstack/echo/middleware"
)

type httpServer struct {
	addr   string
	server *echo.Echo
}

func newHTTPServer(addr string, httpSetup func(*echo.Echo)) *httpServer {
	server := echo.New()
	httpSetup(server)
	server.Use(md.Recover())

	return &httpServer{
		addr:   addr,
		server: server,
	}
}

func (s *httpServer) start() error {
	log.Infof("rpc: start a grpc http proxy server at %s", s.addr)
	return s.server.Start(s.addr)
}

func (s *httpServer) stop() error {
	return s.server.Shutdown(context.Background())
}
