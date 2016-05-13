// Copyright 2014 Wandoujia Inc. All Rights Reserved.
// Licensed under the MIT (MIT-LICENSE.txt) license.

package main

import (
	"flag"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/gateway/conf"
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/fagongzi/gateway/proxy"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strings"
)

var (
	cpus       = flag.Int("cpus", 1, "use cpu nums")
	addr       = flag.String("addr", ":80", "listen addr.(e.g. ip:port)")
	mgrAddr    = flag.String("mgr-addr", ":8081", "manager listen addr.(e.g. ip:port)")
	etcdAddr   = flag.String("etcd-addr", "http://127.0.0.1:2379", "etcd address, use ',' to splite.")
	etcdPrefix = flag.String("etcd-prefix", "/gateway", "etcd node prefix.")
	filters    = flag.String("filters", "analysis,rate-limiting,circuit-breake,http-access,head,xforward", "filters to apply, use ',' to splite.")
)

var (
	enableRuntimeVal = flag.Bool("runtime-val", false, "use var in strings to use.")
	enableHeadVal    = flag.Bool("head-val", false, "set per head like '$head_headname' to runtime val.")
	enableCookieVal  = flag.Bool("cookie-val", false, "set per cookie like '$cookie_cookiename' to runtime val.")

	enablePPROF = flag.Bool("pprof", false, "enable go pprof.")
	pprofAddr   = flag.String("pprof-addr", ":6060", "pprof listen addr.(e.g. ip:port)")
)

var (
	logFile  = flag.String("log-file", "", "which file to record log, if not set stdout to use.")
	logLevel = flag.String("log-level", "info", "log level.")
)

func main() {
	flag.Parse()

	if *enablePPROF {
		go func() {
			log.Println(http.ListenAndServe(*pprofAddr, nil))
		}()
	}

	runtime.GOMAXPROCS(*cpus)

	util.InitLog(*logFile)
	level := util.SetLogLevel(*logLevel)

	config := &conf.Conf{
		Addr:       *addr,
		MgrAddr:    *mgrAddr,
		LogLevel:   level,
		EtcdAddr:   *etcdAddr,
		EtcdPrefix: *etcdPrefix,

		ReqHeadStaticMapping: make(map[string]string),
		RspHeadStaticMapping: make(map[string]string),

		EnableRuntimeVal: *enableRuntimeVal,
		EnableCookieVal:  *enableCookieVal,
		EnableHeadVal:    *enableHeadVal,
	}

	log.Infof("conf:<%+v>", config)

	proxyInfo := &model.ProxyInfo{
		Conf: config,
	}

	store, err := model.NewEtcdStore(strings.Split(config.EtcdAddr, ","), config.EtcdPrefix)

	if err != nil {
		log.Panicf("init etcd store error:<%+v>", err)
	}

	register, _ := store.(model.Register)

	register.Registry(proxyInfo)

	rt := model.NewRouteTable(store)
	rt.Load()

	server := proxy.NewProxy(config, rt)

	fs := strings.Split(*filters, ",")
	for i := 0; i < len(fs); i++ {
		server.RegistryFilter(fs[i])
	}

	server.Start()
}
