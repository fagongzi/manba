package proxy

import (
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

var (
	// ErrCircuitClose server is in circuit close
	ErrCircuitClose = errors.New("server is in circuit close")
	// ErrCircuitHalfLimited server is in circuit half, traffic limit
	ErrCircuitHalfLimited = errors.New("server is in circuit half, traffic limit")

	rd = rand.New(rand.NewSource(time.Now().UnixNano()))
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
	cb := c.Server().CircuitBreaker
	if cb == nil {
		return f.BaseFilter.Pre(c)
	}

	pc := c.(*proxyContext)

	switch pc.circuitStatus() {
	case metapb.Open:
		if c.Analysis().GetRecentlyRequestFailureRate(c.Server().ID, time.Duration(cb.RateCheckPeriod)) >= int(cb.FailureRateToClose) {
			pc.changeCircuitStatusToClose()
			c.Analysis().Reject(c.Server().ID)
			return http.StatusServiceUnavailable, ErrCircuitClose
		}

		return http.StatusOK, nil
	case metapb.Half:
		if pc.circuitRateBarrier().Allow() {
			return f.BaseFilter.Pre(c)
		}

		c.Analysis().Reject(c.Server().ID)
		return http.StatusServiceUnavailable, ErrCircuitHalfLimited
	default:
		c.Analysis().Reject(c.Server().ID)
		return http.StatusServiceUnavailable, ErrCircuitClose
	}
}

// Post execute after proxy
func (f *CircuitBreakeFilter) Post(c filter.Context) (statusCode int, err error) {
	cb := c.Server().CircuitBreaker
	if cb == nil {
		return f.BaseFilter.Pre(c)
	}

	pc := c.(*proxyContext)
	if pc.circuitStatus() == metapb.Half &&
		c.Analysis().GetRecentlyRequestSuccessedRate(c.Server().ID, time.Duration(cb.RateCheckPeriod)) >= int(cb.SucceedRateToOpen) {
		pc.changeCircuitStatusToOpen()
	}

	return f.BaseFilter.Post(c)
}

// PostErr execute proxy has errors
func (f *CircuitBreakeFilter) PostErr(c filter.Context) {
	cb := c.Server().CircuitBreaker
	if cb == nil {
		f.BaseFilter.PostErr(c)
		return
	}

	pc := c.(*proxyContext)
	if pc.circuitStatus() == metapb.Half &&
		c.Analysis().GetRecentlyRequestFailureRate(c.Server().ID, time.Duration(cb.RateCheckPeriod)) >= int(cb.FailureRateToClose) {
		pc.changeCircuitStatusToClose()
	}
}
