package proxy

import (
	_ "github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/gateway/conf"
)

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

func (self BlackListFilter) Name() string {
	return FILTER_BLACKLIST
}
