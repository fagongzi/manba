package api

import "flag"

var (
	addr         = flag.String("addr", ":8080", "listen addr of api server")
	registryAddr = flag.String("registry-addr", "etcd://127.0.0.1:2379", "registry address")
	prefix       = flag.String("prefix", "/gateway", "node prefix.")
)

// Cfg is cfg of api server
type Cfg struct {
	Addr         string
	RegistryAddr string
	Prefix       string
}

// ParseCfg parse cfg from command args
func ParseCfg() *Cfg {
	if !flag.Parsed() {
		flag.Parse()
	}

	return &Cfg{
		Addr:         *addr,
		RegistryAddr: *registryAddr,
		Prefix:       *prefix,
	}
}
