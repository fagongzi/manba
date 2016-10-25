package proxy

import (
	"net/http"

	"github.com/fagongzi/gateway/pkg/model"
	"github.com/valyala/fasthttp"
)

type filterContext struct {
	rw         http.ResponseWriter
	ctx        *fasthttp.RequestCtx
	outreq     *fasthttp.Request
	result     *model.RouteResult
	rb         *model.RouteTable
	startAt    int64
	endAt      int64
	runtimeVar map[string]string
}

// Filter filter interface
type Filter interface {
	Name() string

	Pre(c *filterContext) (statusCode int, err error)
	Post(c *filterContext) (statusCode int, err error)
	PostErr(c *filterContext)
}

type baseFilter struct{}

// Pre execute before proxy
func (f baseFilter) Pre(c *filterContext) (statusCode int, err error) {
	return http.StatusOK, nil
}

// Post execute after proxy
func (f baseFilter) Post(c *filterContext) (statusCode int, err error) {
	return http.StatusOK, nil
}

// PostErr execute proxy has errors
func (f baseFilter) PostErr(c *filterContext) {

}

func (f *Proxy) doPreFilters(c *filterContext) (filterName string, statusCode int, err error) {
	for iter := f.filters.Front(); iter != nil; iter = iter.Next() {
		f, _ := iter.Value.(Filter)
		filterName = f.Name()

		statusCode, err = f.Pre(c)
		if nil != err {
			return filterName, statusCode, err
		}
	}

	return "", http.StatusOK, nil
}

func (f *Proxy) doPostFilters(c *filterContext) (filterName string, statusCode int, err error) {
	for iter := f.filters.Back(); iter != nil; iter = iter.Prev() {
		f, _ := iter.Value.(Filter)

		statusCode, err = f.Post(c)
		if nil != err {
			return filterName, statusCode, err
		}
	}

	return "", http.StatusOK, nil
}

func (f *Proxy) doPostErrFilters(c *filterContext) {
	for iter := f.filters.Back(); iter != nil; iter = iter.Prev() {
		f, _ := iter.Value.(Filter)

		f.PostErr(c)
	}
}
