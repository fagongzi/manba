package grpcx

import (
	"net"

	"github.com/fagongzi/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
)

// Service is a service define
type Service struct {
	Name     string
	Metadata interface{}
}

// ServiceRegister registry grpc services
type ServiceRegister func(*grpc.Server) []Service

// GRPCServer is a grpc server
type GRPCServer struct {
	addr     string
	server   *grpc.Server
	opts     *serverOptions
	register ServiceRegister
	services []Service
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

	if err := s.server.Serve(lis); err != nil {
		return err
	}

	return nil
}

// GracefulStop stop the grpc server
func (s *GRPCServer) GracefulStop() {
	s.server.GracefulStop()
}

func (s *GRPCServer) publishServices() {
	if s.opts.publisher != nil {
		for _, service := range s.services {
			err := s.opts.publisher.Publish(service.Name, naming.Update{
				Op:       naming.Add,
				Addr:     s.addr,
				Metadata: service.Metadata,
			})
			if err != nil {
				log.Fatalf("rpc: publish service <%s> failed, error:\n%+v", service.Name, err)
			}

			log.Infof("rpc: service <%s> already published", service.Name)
		}
	}
}
