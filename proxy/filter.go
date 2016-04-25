package proxy

import (
	"github.com/fagongzi/gateway/pkg/model"
	"net/http"
)

var (
	SUPPORT_FILTERS = []string{"Log", "Headers", "XForwardFor"}
)

type filterContext struct {
	rw         http.ResponseWriter
	req        *http.Request
	outreq     *http.Request
	result     *model.RouteResult
	rb         *model.RouteTable
	startAt    int64
	endAt      int64
	runtimeVar map[string]string
}

type Filter interface {
	Name() string

	Pre(c *filterContext) (statusCode int, err error)
	Post(c *filterContext) (statusCode int, err error)
	PostErr(c *filterContext)
}

type baseFilter struct{}

func (self baseFilter) Pre(c *filterContext) (statusCode int, err error) {
	return http.StatusOK, nil
}

func (self baseFilter) Post(c *filterContext) (statusCode int, err error) {
	return http.StatusOK, nil
}

func (self baseFilter) PostErr(c *filterContext) {

}

func (self *Proxy) doPreFilters(c *filterContext) (filterName string, statusCode int, err error) {
	for iter := self.filters.Front(); iter != nil; iter = iter.Next() {
		f, _ := iter.Value.(Filter)
		filterName = f.Name()

		statusCode, err = f.Pre(c)
		if nil != err {
			return filterName, statusCode, err
		}
	}

	return "", http.StatusOK, nil
}

func (self *Proxy) doPostFilters(c *filterContext) (filterName string, statusCode int, err error) {
	for iter := self.filters.Back(); iter != nil; iter = iter.Prev() {
		f, _ := iter.Value.(Filter)

		statusCode, err = f.Post(c)
		if nil != err {
			return filterName, statusCode, err
		}
	}

	return "", http.StatusOK, nil
}

func (self *Proxy) doPostErrFilters(c *filterContext) {
	for iter := self.filters.Back(); iter != nil; iter = iter.Prev() {
		f, _ := iter.Value.(Filter)

		f.PostErr(c)
	}
}
