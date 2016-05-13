package proxy

import (
	"errors"
	"fmt"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"github.com/fagongzi/gateway/conf"
	"github.com/fagongzi/gateway/pkg/model"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

const (
	RATE_BASE = 100
)

const (
	TIMER_PREFIX = "Circuit-"
)

var (
	ERR_CIRCUIT_CLOSE        = errors.New("server is in circuit close")
	ERR_CIRCUIT_HALF         = errors.New("server is in circuit half")
	ERR_CIRCUIT_HALF_LIMITED = errors.New("server is in circuit half, traffic limit")
)

type evt struct {
	status model.Circuit
	server *model.Server
}

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

func (self CircuitBreakeFilter) Name() string {
	return FILTER_CIRCUIT_BREAKE
}

func (self CircuitBreakeFilter) Pre(c *filterContext) (statusCode int, err error) {
	status := c.result.Svr.GetCircuit()
	count := c.rb.GetAnalysis().GetContinuousFailureCount(c.result.Svr.Addr)

	if status == model.CIRCUIT_OPEN {
		if count > c.result.Svr.CloseCount {
			self.changeToClose(c.result.Svr)
			return http.StatusServiceUnavailable, ERR_CIRCUIT_CLOSE
		}

		return http.StatusOK, nil
	} else if status == model.CIRCUIT_HALF {
		if limitAllow(c.result.Svr.HalfTrafficRate) {
			return self.baseFilter.Pre(c)
		}

		return http.StatusServiceUnavailable, ERR_CIRCUIT_HALF_LIMITED
	} else {
		return http.StatusServiceUnavailable, ERR_CIRCUIT_CLOSE
	}
}

func (self CircuitBreakeFilter) Post(c *filterContext) (statusCode int, err error) {
	status := c.result.Svr.GetCircuit()

	if status == model.CIRCUIT_HALF {
		self.changeToOpen(c.result.Svr)
	}

	return self.baseFilter.Post(c)
}

func (self CircuitBreakeFilter) PostErr(c *filterContext) {
	status := c.result.Svr.GetCircuit()

	if status == model.CIRCUIT_HALF {
		self.changeToClose(c.result.Svr)
	}
}

func (self CircuitBreakeFilter) changeToClose(server *model.Server) {
	server.Lock()
	defer server.UnLock()

	if server.GetCircuit() == model.CIRCUIT_CLOSE {
		return
	}

	server.CloseCircuit()

	log.Warnf("Circuit Server <%s> change to close.", server.Addr)

	self.proxy.routeTable.GetTimeWheel().AddWithId(time.Second*time.Duration(server.HalfToOpen), getKey(server.Addr), self.changeToHalf)
}

func (self CircuitBreakeFilter) changeToOpen(server *model.Server) {
	server.Lock()
	defer server.UnLock()

	if server.GetCircuit() == model.CIRCUIT_OPEN || server.GetCircuit() != model.CIRCUIT_HALF {
		return
	}

	server.OpenCircuit()

	log.Warnf("Circuit Server <%s> change to open.", server.Addr)
}

func (self CircuitBreakeFilter) changeToHalf(key string) {
	addr := getAddr(key)
	server := self.proxy.routeTable.GetServer(addr)

	if nil != server {
		server.HalfCircuit()

		log.Warnf("Circuit Server <%s> change to half.", server.Addr)
	}
}

func getKey(addr string) string {
	return fmt.Sprintf("%s%s", TIMER_PREFIX, addr)
}

func getAddr(key string) string {
	info := strings.Split(key, "-")
	return info[1]
}

func limitAllow(rate int) bool {
	randValue := rand.Intn(RATE_BASE)
	return randValue < rate
}
