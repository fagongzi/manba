package apiserver

import (
	"errors"
	"net"

	"github.com/fagongzi/gateway/pkg/pb/rpcpb"
	"github.com/fagongzi/gateway/pkg/store"
	"github.com/fagongzi/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
)

var (
	errRPCCancel = errors.New("rpc request is cancelled")
)

// GRPCAPIServer is a grpc api server
type GRPCAPIServer struct {
	addr        string
	server      *grpc.Server
	opts        *options
	metaService *metaService
}

// NewGRPCAPIServer returns a grpc server for api server
func NewGRPCAPIServer(addr string, db store.Store, opts ...Option) *GRPCAPIServer {
	serverOptions := &options{}
	for _, opt := range opts {
		opt(serverOptions)
	}

	return &GRPCAPIServer{
		addr:        addr,
		opts:        serverOptions,
		metaService: newMetaService(db),
	}
}

// Start start this api server
func (s *GRPCAPIServer) Start() error {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("grpc api server crash, errors:\n %+v", err)
		}
	}()

	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		log.Fatalf("start grpc api server failed at %s errors:\n %+v",
			s.addr,
			err)
		return err
	}

	s.server = grpc.NewServer()
	rpcpb.RegisterMetaServiceServer(s.server, s.metaService)
	s.publishServices()

	if err := s.server.Serve(lis); err != nil {
		return err
	}

	return nil
}

// GracefulStop stop the grpc server
func (s *GRPCAPIServer) GracefulStop() {
	s.server.GracefulStop()
}

func (s *GRPCAPIServer) publishServices() {
	if s.opts.publisher != nil {
		err := s.opts.publisher.Publish(rpcpb.ServiceMeta, naming.Update{
			Op:   naming.Add,
			Addr: s.addr,
		})
		if err != nil {
			log.Fatalf("rpc: publish service <%s> failed, error:\n%+v", rpcpb.ServiceMeta, err)
		}
	}
}
