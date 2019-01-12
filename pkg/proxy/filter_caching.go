package proxy

import (
	"strings"
	"sync"
	"time"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/fagongzi/goetty"
	"github.com/fagongzi/util/hack"
	"github.com/valyala/fasthttp"
)

var (
	cachePool sync.Pool
)

// CachingFilter cache api result
type CachingFilter struct {
	filter.BaseFilter

	tw    *goetty.TimeoutWheel
	cache *util.Cache
}

func newCachingFilter(maxBytes uint64, tw *goetty.TimeoutWheel) filter.Filter {
	return &CachingFilter{
		tw:    tw,
		cache: util.NewLRUCache(maxBytes),
	}
}

// Name return name of this filter
func (f *CachingFilter) Name() string {
	return FilterCaching
}

// Pre execute before proxy
func (f *CachingFilter) Pre(c filter.Context) (statusCode int, err error) {
	if c.DispatchNode().Cache == nil {
		return f.BaseFilter.Post(c)
	}

	matches, id := getCachingID(c)
	if !matches {
		return f.BaseFilter.Post(c)
	}

	if value, ok := f.cache.Get(id); ok {
		c.SetAttr(filter.UsingCachingValue, value)
	}

	return f.BaseFilter.Post(c)
}

// Post execute after proxy
func (f *CachingFilter) Post(c filter.Context) (statusCode int, err error) {
	if c.DispatchNode().Cache == nil {
		return f.BaseFilter.Post(c)
	}

	matches, id := getCachingID(c)
	if !matches {
		return f.BaseFilter.Post(c)
	}

	f.cache.Add(id, genCachedValue(c))
	if c.DispatchNode().Cache.Deadline > 0 {
		f.tw.Schedule(time.Duration(c.DispatchNode().Cache.Deadline),
			f.removeCache, id)
	}

	return f.BaseFilter.Post(c)
}

func (f *CachingFilter) removeCache(id interface{}) {
	f.cache.Remove(id)
}

func getCachingID(c filter.Context) (bool, string) {
	req := c.ForwardRequest()
	if len(c.DispatchNode().Cache.Conditions) == 0 {
		return true, getID(req, c.DispatchNode().Cache.Keys)
	}

	matches := true
	for _, cond := range c.DispatchNode().Cache.Conditions {
		matches = conditionsMatches(&cond, req)
		if !matches {
			break
		}
	}

	if !matches {
		return false, ""
	}

	return matches, getID(req, c.DispatchNode().Cache.Keys)
}

func getID(req *fasthttp.Request, keys []metapb.Parameter) string {
	size := len(keys)
	if size == 0 {
		return hack.SliceToString(req.RequestURI())
	}

	ids := make([]string, size+1, size+1)
	ids[0] = hack.SliceToString(req.RequestURI())
	for idx, param := range keys {
		ids[idx+1] = paramValue(&param, req)
	}

	return strings.Join(ids, "-")
}

func genCachedValue(c filter.Context) []byte {
	contentType := c.Response().Header.ContentType()
	body := c.Response().Body()
	return filter.NewCachedValue(body, contentType)
}
