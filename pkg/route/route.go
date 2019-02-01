package route

import (
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/util/hack"
)

// Route route for api match
// url define: /conststring/(number|string|enum:m1|m2|m3)[:argname]
type Route struct {
	root node
}

// NewRoute create a api match route
func NewRoute() *Route {
	return &Route{
		root: node{
			nt:    slashType,
			value: slashValue,
		},
	}
}

// Add add a url to this route
func (r *Route) Add(api metapb.API) error {
	p := newParser(hack.StringToSlice(api.URLPattern))
	_, err := p.parse()
	if err != nil {
		return err
	}

	return nil
}
