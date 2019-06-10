package proxy

import (
	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/gateway/pkg/util"
)

// PrepareFilter Must be in the first of the filter chain,
// used to get some public information into the context,
// to avoid subsequent filters to do duplicate things.
type PrepareFilter struct {
	filter.BaseFilter
}

func newPrepareFilter() filter.Filter {
	return &PrepareFilter{}
}

// Init init filter
func (f *PrepareFilter) Init(cfg string) error {
	return nil
}

// Name return name of this filter
func (f *PrepareFilter) Name() string {
	return FilterPrepare
}

// Pre execute before proxy
func (f *PrepareFilter) Pre(c filter.Context) (statusCode int, err error) {
	c.SetAttr(filter.AttrClientRealIP, util.ClientIP(c.OriginRequest()))
	return f.BaseFilter.Pre(c)
}
