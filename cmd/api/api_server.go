package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/fagongzi/gateway/pkg/api"
	"github.com/fagongzi/log"
)

var (
	addr      = flag.String("addr", ":8080", "Addr: client entrypoint")
	addrStore = flag.String("addr-store", "etcd://127.0.0.1:2379", "Addr: store address")
	namespace = flag.String("namespace", "dev", "The namespace to isolation the environment.")
)

func main() {
	flag.Parse()

	log.InitLog()
	runtime.GOMAXPROCS(runtime.NumCPU())

	cfg := &api.Cfg{
		Addr:      *addr,
		AddrStore: *addrStore,
		Namespace: fmt.Sprintf("/%s", *namespace),
	}

	s, err := api.NewServer(cfg)
	if err != nil {
		log.Fatalf("start api server failed, errors:\n%+v", err)
	}

	go s.Start()

	waitStop(s)
}

func waitStop(s *api.Server) {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	sig := <-sc
	s.Stop()
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
