package proxy

import (
	"bytes"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/util/hack"
)

var (
	headerName = []byte("X-Forwarded-For")
)

// XForwardForFilter XForwardForFilter
type XForwardForFilter struct {
	filter.BaseFilter
}

func newXForwardForFilter() filter.Filter {
	return &XForwardForFilter{}
}

// Init init filter
func (f *XForwardForFilter) Init(cfg string) error {
	return nil
}

// Name return name of this filter
func (f *XForwardForFilter) Name() string {
	return FilterXForward
}

// Pre execute before proxy
func (f *XForwardForFilter) Pre(c filter.Context) (statusCode int, err error) {
	prevForward := c.OriginRequest().Request.Header.PeekBytes(headerName)
	if len(prevForward) == 0 {
		c.ForwardRequest().Header.SetBytesKV(headerName, hack.StringToSlice(c.OriginRequest().RemoteIP().String()))
	} else {
		var buf bytes.Buffer
		buf.Write(prevForward)
		buf.WriteByte(',')
		buf.WriteByte(' ')
		buf.WriteString(c.OriginRequest().RemoteIP().String())
		c.ForwardRequest().Header.SetBytesKV(headerName, buf.Bytes())
	}

	return f.BaseFilter.Pre(c)
}
