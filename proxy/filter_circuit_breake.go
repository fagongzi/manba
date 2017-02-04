package proxy

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/gateway/pkg/conf"
	"github.com/fagongzi/gateway/pkg/model"
)

const (
	// RateBase base rate
	RateBase = 100
)

const (
	// TimerPrefix timer prefix
	TimerPrefix = "Circuit-"
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
	baseFilter
	proxy  *Proxy
	config *conf.Conf
}

func newCircuitBreakeFilter(config *conf.Conf, proxy *Proxy) Filter {
	return CircuitBreakeFilter{
		config: config,
		proxy:  proxy,
	}
}

// Name return name of this filter
func (f CircuitBreakeFilter) Name() string {
	return FilterCircuitBreake
}

// Pre execute before proxy
func (f CircuitBreakeFilter) Pre(c *filterContext) (statusCode int, err error) {
	status := c.result.Svr.GetCircuit()

	if status == model.CircuitOpen {
		if f.getFailureRate(c) >= c.result.Svr.OpenToCloseFailureRate {
			f.changeToClose(c.result.Svr)
			return http.StatusServiceUnavailable, ErrCircuitClose
		}

		return http.StatusOK, nil
	} else if status == model.CircuitHalf {
		if limitAllow(c.result.Svr.HalfTrafficRate) {
			return f.baseFilter.Pre(c)
		}

		return http.StatusServiceUnavailable, ErrCircuitHalfLimited
	} else {
		return http.StatusServiceUnavailable, ErrCircuitClose
	}
}

// Post execute after proxy
func (f CircuitBreakeFilter) Post(c *filterContext) (statusCode int, err error) {
	status := c.result.Svr.GetCircuit()

	if status == model.CircuitHalf && f.getSucceedRate(c) >= c.result.Svr.HalfToOpenSucceedRate {
		f.changeToOpen(c.result.Svr)
	}

	return f.baseFilter.Post(c)
}

// PostErr execute proxy has errors
func (f CircuitBreakeFilter) PostErr(c *filterContext) {
	status := c.result.Svr.GetCircuit()

	if status == model.CircuitHalf {
		f.changeToClose(c.result.Svr)
	}
}

func (f CircuitBreakeFilter) changeToClose(server *model.Server) {
	server.Lock()

	if server.GetCircuit() == model.CircuitClose {
		server.UnLock()
		return
	}

	server.CloseCircuit()

	log.Warnf("Circuit Server <%s> change to close.", server.Addr)

	f.proxy.routeTable.GetTimeWheel().AddWithID(time.Second*time.Duration(server.HalfToOpenSeconds), getKey(server.Addr), f.changeToHalf)

	server.UnLock()
}

func (f CircuitBreakeFilter) changeToOpen(server *model.Server) {
	server.Lock()

	if server.GetCircuit() == model.CircuitOpen || server.GetCircuit() != model.CircuitHalf {
		server.UnLock()
		return
	}

	server.OpenCircuit()

	log.Warnf("Circuit Server <%s> change to open.", server.Addr)

	server.UnLock()
}

func (f CircuitBreakeFilter) changeToHalf(key string) {
	addr := getAddr(key)
	server := f.proxy.routeTable.GetServer(addr)

	if nil != server {
		server.Lock()
		server.HalfCircuit()
		server.UnLock()

		log.Warnf("Circuit Server <%s> change to half.", server.Addr)
	}
}

func (f CircuitBreakeFilter) getFailureRate(c *filterContext) int {
	failureCount := c.rb.GetAnalysis().GetRecentlyRequestFailureCount(c.result.Svr.Addr, c.result.Svr.OpenToCloseCollectSeconds)
	totalCount := c.rb.GetAnalysis().GetRecentlyRequestCount(c.result.Svr.Addr, c.result.Svr.OpenToCloseCollectSeconds)

	if totalCount == 0 {
		return 0
	}

	return int(failureCount * 100 / totalCount)
}

func (f CircuitBreakeFilter) getSucceedRate(c *filterContext) int {
	succeedCount := c.rb.GetAnalysis().GetRecentlyRequestSuccessedCount(c.result.Svr.Addr, c.result.Svr.OpenToCloseCollectSeconds)
	totalCount := c.rb.GetAnalysis().GetRecentlyRequestCount(c.result.Svr.Addr, c.result.Svr.OpenToCloseCollectSeconds)

	if totalCount == 0 {
		return 0
	}

	return int(succeedCount * 100 / totalCount)
}

func getKey(addr string) string {
	return fmt.Sprintf("%s%s", TimerPrefix, addr)
}

func getAddr(key string) string {
	info := strings.Split(key, "-")
	return info[1]
}

func limitAllow(rate int) bool {
	randValue := rand.Intn(RateBase)
	return randValue < rate
}
