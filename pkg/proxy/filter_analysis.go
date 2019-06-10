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

// Init init filter
func (f *AnalysisFilter) Init(cfg string) error {
	return nil
}

// Name return name of this filter
func (f *AnalysisFilter) Name() string {
	return FilterAnalysis
}

// Pre execute before proxy
func (f *AnalysisFilter) Pre(c filter.Context) (statusCode int, err error) {
	// TODO: avoid lock overhead in every request
	c.Analysis().Request(c.(*proxyContext).circuitResourceID())
	return f.BaseFilter.Pre(c)
}

// Post execute after proxy
func (f *AnalysisFilter) Post(c filter.Context) (statusCode int, err error) {
	c.Analysis().Response(c.(*proxyContext).circuitResourceID(), c.EndAt().Sub(c.StartAt()).Nanoseconds())
	return f.BaseFilter.Post(c)
}

// PostErr execute proxy has errors
func (f *AnalysisFilter) PostErr(c filter.Context, code int, err error) {
	c.Analysis().Failure(c.(*proxyContext).circuitResourceID())
}
