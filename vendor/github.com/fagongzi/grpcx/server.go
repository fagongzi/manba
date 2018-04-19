package grpcx

import (
	"fmt"
	"net"

	"github.com/fagongzi/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
)

// ServiceRegister registry grpc services
type ServiceRegister func(*grpc.Server) []Service

// GRPCServer is a grpc server
type GRPCServer struct {
	addr       string
	httpServer *httpServer
	server     *grpc.Server
	opts       *serverOptions
	register   ServiceRegister
	services   []Service
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

	if s.opts.httpAddr != "" {
		s.createHTTPServer()
	}

	if s.httpServer != nil {
		go func() {
			err := s.httpServer.start()
			if err != nil {
				log.Fatalf("rpc: start http proxy failed, errors:\n%+v", err)
			}
		}()
	}

	return s.server.Serve(lis)
}

// GracefulStop stop the grpc server
func (s *GRPCServer) GracefulStop() {
	if s.httpServer != nil {
		s.httpServer.stop()
	}
	s.server.GracefulStop()
}

func (s *GRPCServer) createHTTPServer() {
	if s.httpServer == nil {
		s.httpServer = newHTTPServer(s.opts.httpAddr, s.opts.httpSetup)
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
