package proxy

import (
	"github.com/fagongzi/gateway/conf"
)

// BlackListFilter blacklist filter
type BlackListFilter struct {
	baseFilter
	proxy  *Proxy
	config *conf.Conf
}

func newBlackListFilter(config *conf.Conf, proxy *Proxy) Filter {
	return BlackListFilter{
		config: config,
		proxy:  proxy,
	}
}

// Name return name of this filter
func (f BlackListFilter) Name() string {
	return FilterBlackList
}
