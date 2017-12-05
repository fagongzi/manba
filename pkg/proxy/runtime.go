package proxy

import (
	"container/list"
	"fmt"
	"sync"
	"time"

	"github.com/fagongzi/gateway/pkg/lb"
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/fagongzi/goetty"
	"github.com/fagongzi/log"
	"github.com/fagongzi/util/collection"
	"github.com/valyala/fasthttp"
)

type clusterRuntime struct {
	sync.RWMutex

	meta model.Cluster
	svrs *list.List
	lb   lb.LoadBalance
}

func newClusterRuntime(meta *model.Cluster) *clusterRuntime {
	rt := &clusterRuntime{
		meta: *meta,
		svrs: list.New(),
		lb:   lb.NewLoadBalance(meta.LbName),
	}

	return rt
}

func (c *clusterRuntime) updateMeta(meta *model.Cluster) {
	c.Lock()
	c.meta = *meta
	c.lb = lb.NewLoadBalance(meta.LbName)
	c.Unlock()
}

func (c *clusterRuntime) foreach(do func(string)) {
	c.Lock()
	defer c.Unlock()

	for iter := c.svrs.Back(); iter != nil; iter = iter.Prev() {
		addr, _ := iter.Value.(string)
		do(addr)
	}
}

func (c *clusterRuntime) remove(id string) {
	c.Lock()
	defer c.Unlock()

	c.doRemove(id)
}

func (c *clusterRuntime) doRemove(id string) {
	collection.Remove(c.svrs, id)
	log.Infof("runtime: remove <%s, %s> succ.", c.meta.ID, id)
}

func (c *clusterRuntime) add(id string) {
	c.Lock()
	defer c.Unlock()

	if collection.IndexOf(c.svrs, id) >= 0 {
		log.Warnf("bind <%s,%s> already created.", c.meta.ID, id)
		return
	}

	c.svrs.PushBack(id)
	log.Infof("meta: bind <%s,%s> created.", c.meta.ID, id)
}

func (c *clusterRuntime) selectServer(req *fasthttp.Request) string {
	c.RLock()

	index := c.lb.Select(req, c.svrs)
	if 0 > index {
		c.RUnlock()
		return ""
	}

	e := collection.Get(c.svrs, index)
	if nil == e {
		c.RUnlock()
		return ""
	}

	s, _ := e.Value.(string)
	c.RUnlock()
	return s
}

type serverRuntime struct {
	sync.RWMutex

	meta             *model.Server
	tw               *goetty.TimeoutWheel
	status           model.Status
	checkFailCount   int
	prevStatus       model.Status
	useCheckDuration time.Duration
	circuit          model.Circuit
}

func newServerRuntime(meta *model.Server, tw *goetty.TimeoutWheel) *serverRuntime {
	rt := &serverRuntime{
		meta:    meta,
		tw:      tw,
		status:  model.Down,
		circuit: model.CircuitOpen,
	}

	return rt
}

func (s *serverRuntime) updateMeta(meta *model.Server, cb func()) {
	s.Lock()
	s.meta = meta
	cb()
	s.Unlock()
}

func (s *serverRuntime) getCheckURL() string {
	s.RLock()
	v := fmt.Sprintf("%s://%s%s", s.meta.Schema, s.meta.Addr, s.meta.HeathCheck.Path)
	s.RUnlock()

	return v
}

func (s *serverRuntime) fail() {
	s.checkFailCount++
	s.useCheckDuration += s.useCheckDuration / 2
}

func (s *serverRuntime) reset() {
	s.checkFailCount = 0
	s.useCheckDuration = s.meta.HeathCheck.Interval
}

func (s *serverRuntime) changeTo(status model.Status) {
	s.prevStatus = s.status
	s.status = status
}

func (s *serverRuntime) statusChanged() bool {
	return s.prevStatus != s.status
}

func (s *serverRuntime) isCircuitStatus(target model.Circuit) bool {
	s.RLock()
	v := s.circuit == target
	s.RUnlock()
	return v
}

func (s *serverRuntime) circuitToClose() {
	s.Lock()
	if s.meta.CircuitBreaker == nil ||
		s.circuit == model.CircuitClose {
		s.Unlock()
		return
	}

	s.circuit = model.CircuitClose
	log.Warnf("server <%s> change to close", s.meta.ID)

	s.tw.Schedule(s.meta.CircuitBreaker.CloseToHalf, s.circuitToHalf, nil)
	s.Unlock()
}

func (s *serverRuntime) circuitToOpen() {
	s.Lock()
	if s.meta.CircuitBreaker == nil ||
		s.circuit == model.CircuitOpen ||
		s.circuit != model.CircuitHalf {
		s.Unlock()
		return
	}

	s.circuit = model.CircuitOpen
	log.Infof("server <%s> change to open", s.meta.ID)
	s.Unlock()
}

func (s *serverRuntime) circuitToHalf(arg interface{}) {
	s.Lock()
	if s.meta.CircuitBreaker != nil {
		s.circuit = model.CircuitOpen
		log.Warnf("server <%s> change to half", s.meta.ID)
	}
	s.Unlock()
}
