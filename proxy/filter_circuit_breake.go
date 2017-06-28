package proxy

import (
	"errors"
	"math/rand"
	"net/http"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/gateway/pkg/model"
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
)

type evt struct {
	status model.Circuit
	server *model.Server
}

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
	if c.IsCircuitOpen() {
		if f.getFailureRate(c) >= c.GetOpenToCloseFailureRate() {
			c.ChangeCircuitStatusToClose()
			return http.StatusServiceUnavailable, ErrCircuitClose
		}

		return http.StatusOK, nil
	} else if c.IsCircuitHalf() {
		if limitAllow(c.GetHalfTrafficRate()) {
			return f.BaseFilter.Pre(c)
		}

		return http.StatusServiceUnavailable, ErrCircuitHalfLimited
	} else {
		return http.StatusServiceUnavailable, ErrCircuitClose
	}
}

// Post execute after proxy
func (f CircuitBreakeFilter) Post(c filter.Context) (statusCode int, err error) {
	if c.IsCircuitHalf() && f.getSucceedRate(c) >= c.GetHalfToOpenSucceedRate() {
		c.ChangeCircuitStatusToOpen()
	}

	return f.BaseFilter.Post(c)
}

// PostErr execute proxy has errors
func (f CircuitBreakeFilter) PostErr(c filter.Context) {
	if c.IsCircuitHalf() {
		c.ChangeCircuitStatusToClose()
	}
}

func (f CircuitBreakeFilter) getFailureRate(c filter.Context) int {
	failureCount := c.GetRecentlyRequestFailureCount(c.GetOpenToCloseCollectSeconds())
	totalCount := c.GetRecentlyRequestCount(c.GetOpenToCloseCollectSeconds())

	if totalCount == 0 {
		return -1
	}

	return int(failureCount * 100 / totalCount)
}

func (f CircuitBreakeFilter) getSucceedRate(c filter.Context) int {
	succeedCount := c.GetRecentlyRequestSuccessedCount(c.GetOpenToCloseCollectSeconds())
	totalCount := c.GetRecentlyRequestCount(c.GetOpenToCloseCollectSeconds())

	if totalCount == 0 {
		return 100
	}

	return int(succeedCount * 100 / totalCount)
}

func limitAllow(rate int) bool {
	randValue := rand.Intn(RateBase)
	return randValue < rate
}
