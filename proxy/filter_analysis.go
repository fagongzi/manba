package proxy

import (
	"github.com/fagongzi/gateway/conf"
)

// AnalysisFilter analysis filter
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

// Name return name of this filter
func (f AnalysisFilter) Name() string {
	return FilterAnalysis
}

// Pre execute before proxy
func (f AnalysisFilter) Pre(c *filterContext) (statusCode int, err error) {
	c.rb.GetAnalysis().Request(c.result.Svr.Addr)
	return f.baseFilter.Pre(c)
}

// Post execute after proxy
func (f AnalysisFilter) Post(c *filterContext) (statusCode int, err error) {
	c.rb.GetAnalysis().Response(c.result.Svr.Addr, c.endAt-c.startAt)
	return f.baseFilter.Post(c)
}

// PostErr execute proxy has errors
func (f AnalysisFilter) PostErr(c *filterContext) {
	c.rb.GetAnalysis().Failure(c.result.Svr.Addr)
}
