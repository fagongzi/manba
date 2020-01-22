package proxy

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/valyala/fasthttp"
)

var (
	options = []byte("OPTIONS")
)

// CrossCfg cross cfg
type CrossCfg struct {
	Headers []CrossHeader `json:"headers"`
}

// CrossHeader cross header
type CrossHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CrossDomainFilter cross domain
type CrossDomainFilter struct {
	filter.BaseFilter
	cfg CrossCfg
}

func newCrossDomainFilter(file string) (filter.Filter, error) {
	f := &CrossDomainFilter{}

	err := f.parseCfg(file)
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (f *CrossDomainFilter) parseCfg(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &f.cfg)
	if err != nil {
		return err
	}

	return nil
}

// Name return name of this filter
func (f *CrossDomainFilter) Name() string {
	return FilterCross
}

// Pre execute before proxy
func (f *CrossDomainFilter) Pre(c filter.Context) (statusCode int, err error) {
	if bytes.Compare(c.OriginRequest().Method(), options) != 0 {
		return f.BaseFilter.Pre(c)
	}

	resp := fasthttp.AcquireResponse()
	for _, h := range f.cfg.Headers {
		resp.Header.Add(h.Name, h.Value)
	}

	c.SetAttr(filter.AttrUsingResponse, resp)
	return filter.BreakFilterChainCode, nil
}
