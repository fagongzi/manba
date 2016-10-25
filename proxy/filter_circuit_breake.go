package proxy

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/gateway/conf"
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
	count := c.rb.GetAnalysis().GetContinuousFailureCount(c.result.Svr.Addr)

	if status == model.CircuitOpen {
		if count > c.result.Svr.CloseCount {
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

	if status == model.CircuitHalf {
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
	defer server.UnLock()

	if server.GetCircuit() == model.CircuitClose {
		return
	}

	server.CloseCircuit()

	log.Warnf("Circuit Server <%s> change to close.", server.Addr)

	f.proxy.routeTable.GetTimeWheel().AddWithId(time.Second*time.Duration(server.HalfToOpen), getKey(server.Addr), f.changeToHalf)
}

func (f CircuitBreakeFilter) changeToOpen(server *model.Server) {
	server.Lock()
	defer server.UnLock()

	if server.GetCircuit() == model.CircuitOpen || server.GetCircuit() != model.CircuitHalf {
		return
	}

	server.OpenCircuit()

	log.Warnf("Circuit Server <%s> change to open.", server.Addr)
}

func (f CircuitBreakeFilter) changeToHalf(key string) {
	addr := getAddr(key)
	server := f.proxy.routeTable.GetServer(addr)

	if nil != server {
		server.HalfCircuit()

		log.Warnf("Circuit Server <%s> change to half.", server.Addr)
	}
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
