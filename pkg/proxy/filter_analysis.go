package proxy

import (
	"github.com/fagongzi/gateway/pkg/filter"
)

// AnalysisFilter analysis filter
type AnalysisFilter struct {
	filter.BaseFilter
}

func newAnalysisFilter() filter.Filter {
	return &AnalysisFilter{}
}

// Name return name of this filter
func (f AnalysisFilter) Name() string {
	return FilterAnalysis
}

// Pre execute before proxy
func (f AnalysisFilter) Pre(c filter.Context) (statusCode int, err error) {
	c.GetAnalysis().Request(c.GetProxyServerAddr())
	return f.BaseFilter.Pre(c)
}

// Post execute after proxy
func (f AnalysisFilter) Post(c filter.Context) (statusCode int, err error) {
	c.GetAnalysis().Response(c.GetProxyServerAddr(), c.GetEndAt().Sub(c.GetStartAt()).Nanoseconds())
	return f.BaseFilter.Post(c)
}

// PostErr execute proxy has errors
func (f AnalysisFilter) PostErr(c filter.Context) {
	c.GetAnalysis().Failure(c.GetProxyServerAddr())
}
