package main

import (
	"flag"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/fagongzi/gateway/pkg/api"
	"github.com/fagongzi/log"
)

var (
	addr         = flag.String("addr", ":8080", "listen addr.(e.g. ip:port)")
	registryAddr = flag.String("registry-addr", "etcd://127.0.0.1:2379", "registry address")
	prefix       = flag.String("prefix", "/gateway", "node prefix.")
)

var (
	userName = flag.String("user", "admin", "admin user name")
	pwd      = flag.String("pwd", "admin", "admin user pwd")
)

func main() {
	flag.Parse()

	log.InitLog()
	runtime.GOMAXPROCS(runtime.NumCPU())

	s, err := api.NewServer(api.ParseCfg())
	if err != nil {
		log.Fatalf("bootstrap: start api server failed, errors:\n%+v", err)
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
