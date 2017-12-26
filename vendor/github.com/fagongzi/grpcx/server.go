package grpcx

import (
	"fmt"
	"net"

	"github.com/fagongzi/log"
	"github.com/labstack/echo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
)

// ServiceRegister registry grpc services
type ServiceRegister func(*grpc.Server) []Service

// GRPCServer is a grpc server
type GRPCServer struct {
	addr         string
	httpServer   *httpServer
	server       *grpc.Server
	opts         *serverOptions
	register     ServiceRegister
	services     []Service
	httpHandlers []*httpEntrypoint
}

// NewGRPCServer returns a grpc server
func NewGRPCServer(addr string, register ServiceRegister, opts ...ServerOption) *GRPCServer {
	sopts := &serverOptions{}
	for _, opt := range opts {
		opt(sopts)
	}

	return &GRPCServer{
		addr:     addr,
		opts:     sopts,
		register: register,
	}
}

// AddGetHTTPHandler add get http handler
func (s *GRPCServer) AddGetHTTPHandler(path string, handler func(echo.Context) error) {
	s.addHTTPHandler(path, echo.GET, handler)
}

// AddPostHTTPHandler add post http handler
func (s *GRPCServer) AddPostHTTPHandler(path string, handler func(echo.Context) error) {
	s.addHTTPHandler(path, echo.POST, handler)
}

// AddPutHTTPHandler add put http handler
func (s *GRPCServer) AddPutHTTPHandler(path string, handler func(echo.Context) error) {
	s.addHTTPHandler(path, echo.PUT, handler)
}

// AddDeleteHTTPHandler add delete http handler
func (s *GRPCServer) AddDeleteHTTPHandler(path string, handler func(echo.Context) error) {
	s.addHTTPHandler(path, echo.DELETE, handler)
}

// Start start this api server
func (s *GRPCServer) Start() error {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("rpc: grpc server crash, errors:\n %+v", err)
		}
	}()

	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatalf("rpc: start grpc server failed at %s errors:\n %+v",
			s.addr,
			err)
		return err
	}

	s.server = grpc.NewServer()
	s.services = s.register(s.server)
	s.publishServices()

	if s.opts.httpServer != "" {
		s.createHTTPServer()
		for _, service := range s.services {
			if len(service.opts.httpEntrypoints) > 0 {
				s.httpServer.addHTTPEntrypoints(service.opts.httpEntrypoints...)
				log.Infof("rpc: service %s added to http proxy", service.Name)
			}
		}
	}

	if len(s.httpHandlers) > 0 {
		s.createHTTPServer()
		s.httpServer.addHTTPEntrypoints(s.httpHandlers...)
	}

	if s.httpServer != nil {
		go func() {
			err := s.httpServer.start()
			if err != nil {
				log.Fatalf("rpc: start http proxy failed, errors:\n%+v", err)
			}
		}()
	}

	if err := s.server.Serve(lis); err != nil {
		return err
	}

	return nil
}

// GracefulStop stop the grpc server
func (s *GRPCServer) GracefulStop() {
	if s.httpServer != nil {
		s.httpServer.stop()
	}
	s.server.GracefulStop()
}

func (s *GRPCServer) addHTTPHandler(path, method string, handler func(echo.Context) error) {
	s.httpHandlers = append(s.httpHandlers, &httpEntrypoint{
		path:    path,
		method:  method,
		handler: handler,
	})
}

func (s *GRPCServer) createHTTPServer() {
	if s.httpServer == nil {
		s.httpServer = newHTTPServer(s.opts.httpServer)
	}
}

func (s *GRPCServer) publishServices() {
	if s.opts.publisher != nil {
		for _, service := range s.services {
			err := s.opts.publisher.Publish(service.Name, naming.Update{
				Op:       naming.Add,
				Addr:     adjustAddr(s.addr),
				Metadata: service.Metadata,
			})
			if err != nil {
				log.Fatalf("rpc: publish service <%s> failed, error:\n%+v", service.Name, err)
			}

			log.Infof("rpc: service <%s> already published", service.Name)
		}
	}
}

func adjustAddr(addr string) string {
	if addr[0] == ':' {
		ips, err := intranetIP()
		if err != nil {
			log.Fatalf("get intranet ip failed, error:\n%+v", err)
		}

		return fmt.Sprintf("%s%s", ips[0], addr)
	}

	return addr
}
