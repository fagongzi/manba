package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/fagongzi/gateway/pkg/proxy"
	"github.com/fagongzi/log"
)

type filterFlag []string

func (f *filterFlag) String() string {
	return "filter flag"
}

func (f *filterFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

var (
	defaultFilters = &filterFlag{}
	filters        = &filterFlag{}

	addr                          = flag.String("addr", "127.0.0.1:80", "Addr: http request entrypoint")
	addrRPC                       = flag.String("addr-rpc", "127.0.0.1:9091", "Addr: manager request entrypoint")
	addrStore                     = flag.String("addr-store", "etcd://127.0.0.1:2379", "Addr: store of meta data, support etcd or consul")
	addrPPROF                     = flag.String("addr-pprof", "", "Addr: pprof addr")
	namespace                     = flag.String("namespace", "dev", "The namespace to isolation the environment.")
	limitCountHeathCheckWorker    = flag.Int("limit-heathcheck", 1, "Limit: Count of heath check worker")
	limitIntervalHeathCheckSec    = flag.Int("limit-heathcheck-interval", 60, "Limit(sec): Interval for heath check")
	limitCountConn                = flag.Int("limit-conn", 64, "Limit(count): Count of connection per backend server")
	limitDurationConnKeepaliveSec = flag.Int("limit-conn-keepalive", 60, "Limit(sec): Keepalive for backend server connections")
	limitDurationConnIdleSec      = flag.Int("limit-conn-idle", 30, "Limit(sec): Idle for backend server connections")
	limitTimeoutWriteSec          = flag.Int("limit-timeout-write", 30, "Limit(sec): Timeout for write to backend servers")
	limitTimeoutReadSec           = flag.Int("limit-timeout-read", 30, "Limit(sec): Timeout for read from backend servers")
	limitBufferRead               = flag.Int("limit-buf-read", 2048, "Limit(bytes): Bytes for read buffer size")
	limitBufferWrite              = flag.Int("limit-buf-write", 1024, "Limit(bytes): Bytes for write buffer size")
	limitBytesBodyMB              = flag.Int("limit-body", 10, "Limit(MB): MB for body size")
	ttlProxy                      = flag.Int64("ttl-proxy", 10, "TTL(secs): proxy")
)

func init() {
	defaultFilters.Set(proxy.FilterWhiteList)
	defaultFilters.Set(proxy.FilterBlackList)
	defaultFilters.Set(proxy.FilterAnalysis)
	defaultFilters.Set(proxy.FilterRateLimiting)
	defaultFilters.Set(proxy.FilterCircuitBreake)
	defaultFilters.Set(proxy.FilterHTTPAccess)
	defaultFilters.Set(proxy.FilterHeader)
	defaultFilters.Set(proxy.FilterXForward)
	defaultFilters.Set(proxy.FilterValidation)
}

func main() {
	flag.Var(filters, "filter", "Plugin(Filter): format is <filter name>[:plugin file path]")
	flag.Parse()

	log.InitLog()
	runtime.GOMAXPROCS(runtime.NumCPU())

	if *addrPPROF != "" {
		go func() {
			log.Errorf("start pprof failed, errors:\n%+v",
				http.ListenAndServe(*addrPPROF, nil))
		}()
	}

	p := proxy.NewProxy(getCfg())
	go p.Start()

	waitStop(p)
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

func getCfg() *proxy.Cfg {
	cfg := &proxy.Cfg{
		Option: &proxy.Option{},
	}

	cfg.Addr = *addr
	cfg.AddrRPC = *addrRPC
	cfg.AddrPPROF = *addrPPROF
	cfg.AddrStore = *addrStore
	cfg.TTLProxy = *ttlProxy
	cfg.Namespace = fmt.Sprintf("/%s", *namespace)
	cfg.Option.LimitBytesBody = *limitBytesBodyMB * 1024 * 1024
	cfg.Option.LimitBufferRead = *limitBufferRead
	cfg.Option.LimitBufferWrite = *limitBufferWrite
	cfg.Option.LimitCountConn = *limitCountConn
	cfg.Option.LimitCountHeathCheckWorker = *limitCountHeathCheckWorker
	cfg.Option.LimitDurationConnIdle = time.Second * time.Duration(*limitDurationConnIdleSec)
	cfg.Option.LimitDurationConnKeepalive = time.Second * time.Duration(*limitDurationConnKeepaliveSec)
	cfg.Option.LimitTimeoutRead = time.Second * time.Duration(*limitTimeoutReadSec)
	cfg.Option.LimitTimeoutWrite = time.Second * time.Duration(*limitTimeoutWriteSec)
	cfg.Option.LimitIntervalHeathCheck = time.Second * time.Duration(*limitIntervalHeathCheckSec)

	specs := defaultFilters
	if len(*filters) > 0 {
		specs = filters
	}

	for _, spec := range *specs {
		filter, err := proxy.ParseFilter(spec)
		if err != nil {
			log.Fatalf("boostrap: parse filter failed: errors:\n%+v", err)
		}

		cfg.AddFilter(filter)
	}

	return cfg
}
