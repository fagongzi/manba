package proxy

import (
	"errors"
	"math/rand"
	"net/http"
	"time"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
)

const (
	// RateBase base rate
	RateBase = 100
)

var (
	// ErrCircuitClose server is in circuit close
	ErrCircuitClose = errors.New("server is in circuit close")
	// ErrCircuitHalf server is in circuit half
	ErrCircuitHalf = errors.New("server is in circuit half")
	// ErrCircuitHalfLimited server is in circuit half, traffic limit
	ErrCircuitHalfLimited = errors.New("server is in circuit half, traffic limit")

	rd = rand.New(rand.NewSource(time.Now().UnixNano()))
)

// CircuitBreakeFilter CircuitBreakeFilter
type CircuitBreakeFilter struct {
	filter.BaseFilter
}

func newCircuitBreakeFilter() filter.Filter {
	return CircuitBreakeFilter{}
}

// Name return name of this filter
func (f CircuitBreakeFilter) Name() string {
	return FilterCircuitBreake
}

// Pre execute before proxy
func (f CircuitBreakeFilter) Pre(c filter.Context) (statusCode int, err error) {
	cb := c.Server().CircuitBreaker
	if cb == nil {
		return f.BaseFilter.Pre(c)
	}

	switch c.CircuitStatus() {
	case metapb.Open:
		if f.getFailureRate(c) >= int(cb.FailureRateToClose) {
			c.ChangeCircuitStatusToClose()
			return http.StatusServiceUnavailable, ErrCircuitClose
		}

		return http.StatusOK, nil
	case metapb.Half:
		if limitAllow(cb.HalfTrafficRate) {
			return f.BaseFilter.Pre(c)
		}

		return http.StatusServiceUnavailable, ErrCircuitHalfLimited
	default:
		return http.StatusServiceUnavailable, ErrCircuitClose
	}
}

// Post execute after proxy
func (f CircuitBreakeFilter) Post(c filter.Context) (statusCode int, err error) {
	cb := c.Server().CircuitBreaker
	if cb == nil {
		return f.BaseFilter.Pre(c)
	}

	if c.CircuitStatus() == metapb.Half && f.getSucceedRate(c) >= int(cb.SucceedRateToOpen) {
		c.ChangeCircuitStatusToOpen()
	}

	return f.BaseFilter.Post(c)
}

// PostErr execute proxy has errors
func (f CircuitBreakeFilter) PostErr(c filter.Context) {
	if c.CircuitStatus() == metapb.Half {
		c.ChangeCircuitStatusToClose()
	}
}

func (f CircuitBreakeFilter) getFailureRate(c filter.Context) int {
	cb := c.Server().CircuitBreaker

	failureCount := c.Analysis().GetRecentlyRequestFailureCount(c.Server().ID, time.Duration(cb.RateCheckPeriod))
	totalCount := c.Analysis().GetRecentlyRequestCount(c.Server().ID, time.Duration(cb.RateCheckPeriod))

	if totalCount == 0 {
		return -1
	}

	return int(failureCount * 100 / totalCount)
}

func (f CircuitBreakeFilter) getSucceedRate(c filter.Context) int {
	cb := c.Server().CircuitBreaker

	succeedCount := c.Analysis().GetRecentlyRequestSuccessedCount(c.Server().ID, time.Duration(cb.RateCheckPeriod))
	totalCount := c.Analysis().GetRecentlyRequestCount(c.Server().ID, time.Duration(cb.RateCheckPeriod))

	if totalCount == 0 {
		return 100
	}

	return int(succeedCount * 100 / totalCount)
}

func limitAllow(rate int32) bool {
	randValue := rd.Intn(RateBase)
	return randValue < int(rate)
}
