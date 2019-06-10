package proxy

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

var (
	// ErrCircuitClose resource is in circuit close
	ErrCircuitClose = errors.New("resource is in circuit close")
	// ErrCircuitHalfLimited resource is in circuit half, traffic limit
	ErrCircuitHalfLimited = errors.New("resource is in circuit half, traffic limit")
)

// CircuitBreakeFilter CircuitBreakeFilter
type CircuitBreakeFilter struct {
	filter.BaseFilter
}

func newCircuitBreakeFilter() filter.Filter {
	return &CircuitBreakeFilter{}
}

// Init init filter
func (f *CircuitBreakeFilter) Init(cfg string) error {
	return nil
}

// Name return name of this filter
func (f *CircuitBreakeFilter) Name() string {
	return FilterCircuitBreake
}

// Pre execute before proxy
func (f *CircuitBreakeFilter) Pre(c filter.Context) (statusCode int, err error) {
	pc := c.(*proxyContext)
	cb, barrier := pc.circuitBreaker()
	if cb == nil {
		return f.BaseFilter.Pre(c)
	}

	protectedResourceStatus := pc.circuitStatus()
	protectedResource := pc.circuitResourceID()

	switch protectedResourceStatus {
	case metapb.Open:
		if c.Analysis().GetRecentlyRequestFailureRate(protectedResource, time.Duration(cb.RateCheckPeriod)) >= int(cb.FailureRateToClose) {
			pc.changeCircuitStatusToClose()
			c.Analysis().Reject(protectedResource)
			return http.StatusServiceUnavailable, ErrCircuitClose
		}

		return http.StatusOK, nil
	case metapb.Half:
		if barrier.Allow() {
			return f.BaseFilter.Pre(c)
		}

		c.Analysis().Reject(protectedResource)
		return http.StatusServiceUnavailable, ErrCircuitHalfLimited
	default:
		c.Analysis().Reject(protectedResource)
		return http.StatusServiceUnavailable, ErrCircuitClose
	}
}

// Post execute after proxy
func (f *CircuitBreakeFilter) Post(c filter.Context) (statusCode int, err error) {
	pc := c.(*proxyContext)
	cb, _ := pc.circuitBreaker()
	if cb == nil {
		return f.BaseFilter.Post(c)
	}

	protectedResourceStatus := pc.circuitStatus()
	protectedResource := pc.circuitResourceID()

	if protectedResourceStatus == metapb.Half &&
		c.Analysis().GetRecentlyRequestSuccessedRate(protectedResource, time.Duration(cb.RateCheckPeriod)) >= int(cb.SucceedRateToOpen) {
		pc.changeCircuitStatusToOpen()
	}

	return f.BaseFilter.Post(c)
}

// PostErr execute proxy has errors
func (f *CircuitBreakeFilter) PostErr(c filter.Context, code int, err error) {
	// ignore user cancel
	if nil != err && strings.HasPrefix(err.Error(), ErrPrefixRequestCancel) {
		f.BaseFilter.PostErr(c, code, err)
		return
	}

	pc := c.(*proxyContext)
	cb, _ := pc.circuitBreaker()
	if cb == nil {
		f.BaseFilter.PostErr(c, code, err)
		return
	}

	protectedResourceStatus := pc.circuitStatus()
	protectedResource := pc.circuitResourceID()

	if protectedResourceStatus == metapb.Half &&
		c.Analysis().GetRecentlyRequestFailureRate(protectedResource, time.Duration(cb.RateCheckPeriod)) >= int(cb.FailureRateToClose) {
		pc.changeCircuitStatusToClose()
	}
}
