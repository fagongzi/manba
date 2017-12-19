package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/fagongzi/gateway/pkg/pb/rpcpb"
	"github.com/fagongzi/gateway/pkg/service"
	"github.com/fagongzi/gateway/pkg/store"
	"github.com/fagongzi/grpcx"
	"github.com/fagongzi/log"
	"google.golang.org/grpc"
)

var (
	addr           = flag.String("addr", "127.0.0.1:9092", "Addr: client entrypoint")
	addrStore      = flag.String("addr-store", "etcd://127.0.0.1:2379", "Addr: store address")
	namespace      = flag.String("namespace", "dev", "The namespace to isolation the environment.")
	discovery      = flag.Bool("discovery", false, "Publish apiserver service via discovery.")
	servicePrefix  = flag.String("service-prefix", "/services", "The prefix for service name.")
	publishLease   = flag.Int64("publish-lease", 10, "Publish service lease seconds")
	publishTimeout = flag.Int("publish-timeout", 30, "Publish service timeout seconds")
)

func main() {
	flag.Parse()

	log.InitLog()
	runtime.GOMAXPROCS(runtime.NumCPU())

	db, err := store.GetStoreFrom(*addrStore, fmt.Sprintf("/%s", *namespace))
	if err != nil {
		log.Fatalf("init store failed for %s, errors:\n%+v",
			*addrStore,
			err)
	}

	var opts []grpcx.ServerOption
	if *discovery {
		opts = append(opts, grpcx.WithEtcdPublisher(db.Raw().(*clientv3.Client), *servicePrefix, *publishLease, time.Second*time.Duration(*publishTimeout)))
	}

	s := grpcx.NewGRPCServer(*addr, func(svr *grpc.Server) []grpcx.Service {
		var services []grpcx.Service

		rpcpb.RegisterMetaServiceServer(svr, service.NewMetaService(db))
		services = append(services, grpcx.Service{
			Name: rpcpb.ServiceMeta,
		})

		return services
	}, opts...)

	go s.Start()

	waitStop(s)
}

func waitStop(s *grpcx.GRPCServer) {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	sig := <-sc
	s.GracefulStop()
	log.Infof("exit: signal=<%d>.", sig)
	switch sig {
	case syscall.SIGTERM:
		log.Infof("exit: bye :-).")
		os.Exit(0)
	default:
		log.Infof("exit: bye :-(.")
		os.Exit(1)
	}
}
