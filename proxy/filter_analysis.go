package proxy

import (
	"github.com/fagongzi/gateway/conf"
)

type AnalysisFilter struct {
	baseFilter
	proxy  *Proxy
	config *conf.Conf
}

func newAnalysisFilter(config *conf.Conf, proxy *Proxy) Filter {
	return AnalysisFilter{
		config: config,
		proxy:  proxy,
	}
}

func (self AnalysisFilter) Name() string {
	return FILTER_ANALYSIS
}

func (self AnalysisFilter) Pre(c *filterContext) (statusCode int, err error) {
	c.rb.GetAnalysis().Request(c.result.Svr.Addr)
	return self.baseFilter.Pre(c)
}

func (self AnalysisFilter) Post(c *filterContext) (statusCode int, err error) {
	c.rb.GetAnalysis().Response(c.result.Svr.Addr, c.endAt-c.startAt)
	return self.baseFilter.Post(c)
}

func (self AnalysisFilter) PostErr(c *filterContext) {
	c.rb.GetAnalysis().Failure(c.result.Svr.Addr)
}
