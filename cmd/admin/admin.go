package main

import (
	"flag"
	"runtime"

	"github.com/fagongzi/gateway/cmd/admin/pkg/server"
)

var (
	cpus         = flag.Int("cpus", 1, "use cpu nums")
	addr         = flag.String("addr", ":8080", "listen addr.(e.g. ip:port)")
	registryAddr = flag.String("registry-addr", "[ectd|consul]://127.0.0.1:8500", "registry address")
	prefix       = flag.String("prefix", "/gateway", "node prefix.")
)

var (
	userName = flag.String("user", "admin", "admin user name")
	pwd      = flag.String("pwd", "admin", "admin user pwd")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(*cpus)
	s := server.NewAdminServer(*addr, *registryAddr, *prefix, *userName, *pwd)
	s.Start()
}
