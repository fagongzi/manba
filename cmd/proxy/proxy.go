package main

import (
	"flag"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/fagongzi/gateway/pkg/conf"
	"github.com/fagongzi/gateway/pkg/proxy"
	"github.com/fagongzi/log"
)

var (
	configFile = flag.String("cfg", "", "config file")
)

func main() {
	flag.Parse()

	log.InitLog()
	runtime.GOMAXPROCS(runtime.NumCPU())

	cnf := conf.GetCfg(*configFile)

	enablePPROF(cnf)

	log.Infof("bootstrap: gateway proxy start with conf:<%+v>",
		cnf)

	p := proxy.NewProxy(cnf)
	go p.Start()

	waitStop(p)
}

func enablePPROF(cnf *conf.Conf) {
	if cnf.EnablePPROF {
		go func() {
			log.Errorf("bootstrap: start pprof failed, errors:\n%+v",
				http.ListenAndServe(cnf.PPROFAddr, nil))
		}()
	}
}

func waitStop(p *proxy.Proxy) {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	sig := <-sc
	p.Stop()
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
