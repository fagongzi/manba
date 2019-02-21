package client

import (
	"time"

	"github.com/fagongzi/gateway/pkg/pb"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/pb/rpcpb"
)

// ServerBuilder server builder
type ServerBuilder struct {
	c     *client
	value metapb.Server
}

// NewServerBuilder return a server build
func (c *client) NewServerBuilder() *ServerBuilder {
	return &ServerBuilder{
		c:     c,
		value: metapb.Server{},
	}
}

// Use use a server
func (sb *ServerBuilder) Use(value metapb.Server) *ServerBuilder {
	sb.value = value
	return sb
}

// NoHeathCheck no heath check
func (sb *ServerBuilder) NoHeathCheck() *ServerBuilder {
	sb.value.HeathCheck = nil
	return sb
}

// CheckHTTPCode use a heath check
func (sb *ServerBuilder) CheckHTTPCode(path string, interval time.Duration, timeout time.Duration) *ServerBuilder {
	if sb.value.HeathCheck == nil {
		sb.value.HeathCheck = &metapb.HeathCheck{}

	}

	sb.value.HeathCheck.Path = path
	sb.value.HeathCheck.Body = ""
	sb.value.HeathCheck.CheckInterval = int64(interval)
	sb.value.HeathCheck.Timeout = int64(timeout)
	return sb
}

// CheckHTTPBody use a heath check
func (sb *ServerBuilder) CheckHTTPBody(path, body string, interval time.Duration, timeout time.Duration) *ServerBuilder {
	if sb.value.HeathCheck == nil {
		sb.value.HeathCheck = &metapb.HeathCheck{}

	}

	sb.value.HeathCheck.Path = path
	sb.value.HeathCheck.Body = body
	sb.value.HeathCheck.CheckInterval = int64(interval)
	sb.value.HeathCheck.Timeout = int64(timeout)
	return sb
}

// Addr set addr
func (sb *ServerBuilder) Addr(addr string) *ServerBuilder {
	sb.value.Addr = addr
	return sb
}

// HTTPBackend set backend is http backend
func (sb *ServerBuilder) HTTPBackend() *ServerBuilder {
	sb.value.Protocol = metapb.HTTP
	return sb
}

// MaxQPS set max qps
func (sb *ServerBuilder) MaxQPS(max int64) *ServerBuilder {
	sb.value.MaxQPS = max
	return sb
}

// Weight set robin weight
func (sb *ServerBuilder) Weight(weight int64) *ServerBuilder {
	sb.value.Weight = weight
	return sb
}

// NoCircuitBreaker no circuit breaker
func (sb *ServerBuilder) NoCircuitBreaker() *ServerBuilder {
	sb.value.CircuitBreaker = nil
	return sb
}

// CircuitBreakerCheckPeriod set circuit breaker period
func (sb *ServerBuilder) CircuitBreakerCheckPeriod(checkPeriod time.Duration) *ServerBuilder {
	if sb.value.CircuitBreaker == nil {
		sb.value.CircuitBreaker = &metapb.CircuitBreaker{}
	}

	sb.value.CircuitBreaker.RateCheckPeriod = int64(checkPeriod)
	return sb
}

// CircuitBreakerHalfTrafficRate set circuit breaker traffic in half status
func (sb *ServerBuilder) CircuitBreakerHalfTrafficRate(rate int) *ServerBuilder {
	if sb.value.CircuitBreaker == nil {
		sb.value.CircuitBreaker = &metapb.CircuitBreaker{}
	}

	sb.value.CircuitBreaker.HalfTrafficRate = int32(rate)
	return sb
}

// CircuitBreakerCloseToHalfTimeout set circuit breaker timeout that close status convert to half
func (sb *ServerBuilder) CircuitBreakerCloseToHalfTimeout(timeout time.Duration) *ServerBuilder {
	if sb.value.CircuitBreaker == nil {
		sb.value.CircuitBreaker = &metapb.CircuitBreaker{}
	}

	sb.value.CircuitBreaker.CloseTimeout = int64(timeout)
	return sb
}

// CircuitBreakerHalfToCloseCondition set circuit breaker condition of half convert to close
func (sb *ServerBuilder) CircuitBreakerHalfToCloseCondition(failureRate int) *ServerBuilder {
	if sb.value.CircuitBreaker == nil {
		sb.value.CircuitBreaker = &metapb.CircuitBreaker{}
	}

	sb.value.CircuitBreaker.FailureRateToClose = int32(failureRate)
	return sb
}

// CircuitBreakerHalfToOpenCondition set circuit breaker condition of half convert to open
func (sb *ServerBuilder) CircuitBreakerHalfToOpenCondition(succeedRate int) *ServerBuilder {
	if sb.value.CircuitBreaker == nil {
		sb.value.CircuitBreaker = &metapb.CircuitBreaker{}
	}

	sb.value.CircuitBreaker.SucceedRateToOpen = int32(succeedRate)
	return sb
}

// Commit commit
func (sb *ServerBuilder) Commit() (uint64, error) {
	err := pb.ValidateServer(&sb.value)
	if err != nil {
		return 0, err
	}

	return sb.c.putServer(sb.value)
}

// Build build
func (sb *ServerBuilder) Build() (*rpcpb.PutServerReq, error) {
	err := pb.ValidateServer(&sb.value)
	if err != nil {
		return nil, err
	}

	return &rpcpb.PutServerReq{
		Server: sb.value,
	}, nil
}
