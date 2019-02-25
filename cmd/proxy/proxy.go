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
	"github.com/fagongzi/gateway/pkg/util"
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
	addrStore                     = flag.String("addr-store", "etcd://127.0.0.1:2379", "Addr: store of meta data, support etcd")
	addrStoreUser                 = flag.String("addr-store-user", "", "addr Store UserName")
	addrStorePwd                  = flag.String("addr-store-pwd", "", "addr Store Password")
	addrPPROF                     = flag.String("addr-pprof", "", "Addr: pprof addr")
	namespace                     = flag.String("namespace", "dev", "The namespace to isolation the environment.")
	limitCpus                     = flag.Int("limit-cpus", 0, "Limit: schedule threads count")
	limitCountDispatchWorker      = flag.Int("limit-dispatch", 64, "Limit: Count of dispatch worker")
	limitCountCopyWorker          = flag.Int("limit-copy", 4, "Limit: Count of copy worker")
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
	limitBytesCachingMB           = flag.Uint64("limit-caching", 64, "Limit(MB): MB for caching size")
	ttlProxy                      = flag.Int64("ttl-proxy", 10, "TTL(secs): proxy")
	version                       = flag.Bool("version", false, "Show version info")

	// internal plugin configuration file
	jwtCfg = flag.String("jwt", "", "PLugin(JWT): jwt plugin configuration file, json format")

	// metric
	metricJob          = flag.String("metric-job", "", "prometheus job name")
	metricInstance     = flag.String("metric-instance", "", "prometheus instance name")
	metricAddress      = flag.String("metric-address", "", "prometheus proxy address")
	metricIntervalSync = flag.Uint64("interval-metric-sync", 0, "Interval(sec): metric sync")

	// enable features
	enableWebSocket = flag.Bool("websocket", false, "enable websocket")
)

func init() {
	defaultFilters.Set(proxy.FilterWhiteList)
	defaultFilters.Set(proxy.FilterBlackList)
	defaultFilters.Set(proxy.FilterCaching)
	defaultFilters.Set(proxy.FilterAnalysis)
	defaultFilters.Set(proxy.FilterRateLimiting)
	defaultFilters.Set(proxy.FilterCircuitBreake)
	defaultFilters.Set(proxy.FilterHTTPAccess)
	defaultFilters.Set(proxy.FilterHeader)
	defaultFilters.Set(proxy.FilterXForward)
	defaultFilters.Set(proxy.FilterValidation)
	defaultFilters.Set(proxy.FilterJSPlugin)
}

func main() {
	flag.Var(filters, "filter", "Plugin(Filter): format is <filter name>[:plugin file path][:plugin config file path]")
	flag.Parse()

	if *version && util.PrintVersion() {
		os.Exit(0)
	}

	log.InitLog()

	if *limitCpus == 0 {
		runtime.GOMAXPROCS(runtime.NumCPU())
	} else {
		runtime.GOMAXPROCS(*limitCpus)
	}

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
		Metric: util.NewMetricCfg(*metricJob, *metricInstance, *metricAddress, time.Second*time.Duration(*metricIntervalSync)),
	}

	cfg.Addr = *addr
	cfg.AddrRPC = *addrRPC
	cfg.AddrPPROF = *addrPPROF
	cfg.AddrStore = *addrStore
	cfg.AddrStoreUserName = *addrStoreUser
	cfg.AddrStorePwd = *addrStorePwd
	cfg.TTLProxy = *ttlProxy
	cfg.Namespace = fmt.Sprintf("/%s", *namespace)
	cfg.Option.LimitBytesBody = *limitBytesBodyMB * 1024 * 1024
	cfg.Option.LimitBytesCaching = *limitBytesCachingMB * 1024 * 1024
	cfg.Option.LimitBufferRead = *limitBufferRead
	cfg.Option.LimitBufferWrite = *limitBufferWrite
	cfg.Option.LimitCountConn = *limitCountConn
	cfg.Option.LimitCountDispatchWorker = uint64(*limitCountDispatchWorker)
	cfg.Option.LimitCountCopyWorker = uint64(*limitCountCopyWorker)
	cfg.Option.LimitCountHeathCheckWorker = *limitCountHeathCheckWorker
	cfg.Option.LimitDurationConnIdle = time.Second * time.Duration(*limitDurationConnIdleSec)
	cfg.Option.LimitDurationConnKeepalive = time.Second * time.Duration(*limitDurationConnKeepaliveSec)
	cfg.Option.LimitTimeoutRead = time.Second * time.Duration(*limitTimeoutReadSec)
	cfg.Option.LimitTimeoutWrite = time.Second * time.Duration(*limitTimeoutWriteSec)
	cfg.Option.LimitIntervalHeathCheck = time.Second * time.Duration(*limitIntervalHeathCheckSec)
	cfg.Option.JWTCfgFile = *jwtCfg
	cfg.Option.EnableWebSocket = *enableWebSocket

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
