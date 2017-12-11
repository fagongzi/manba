package proxy

import (
	"container/list"
	"fmt"
	"time"

	"github.com/fagongzi/gateway/pkg/lb"
	"github.com/fagongzi/gateway/pkg/model"
	"github.com/fagongzi/goetty"
	"github.com/fagongzi/log"
	"github.com/fagongzi/util/collection"
	"github.com/valyala/fasthttp"
)

type clusterRuntime struct {
	meta *model.Cluster
	svrs *list.List
	lb   lb.LoadBalance
}

func newClusterRuntime(meta *model.Cluster) *clusterRuntime {
	rt := &clusterRuntime{
		meta: meta,
		svrs: list.New(),
		lb:   lb.NewLoadBalance(meta.LbName),
	}

	return rt
}

func (c *clusterRuntime) updateMeta(meta *model.Cluster) {
	c.meta = meta
	c.lb = lb.NewLoadBalance(meta.LbName)
}

func (c *clusterRuntime) foreach(do func(string)) {
	for iter := c.svrs.Back(); iter != nil; iter = iter.Prev() {
		addr, _ := iter.Value.(string)
		do(addr)
	}
}

func (c *clusterRuntime) remove(id string) {
	collection.Remove(c.svrs, id)
	log.Infof("cluster <%s> remove server <%s> succ.", c.meta.ID, id)
}

func (c *clusterRuntime) add(id string) {
	if collection.IndexOf(c.svrs, id) >= 0 {
		log.Warnf("bind <%s,%s> already created.", c.meta.ID, id)
		return
	}

	c.svrs.PushBack(id)
	log.Infof("meta: bind <%s,%s> created.", c.meta.ID, id)
}

func (c *clusterRuntime) selectServer(req *fasthttp.Request) string {
	index := c.lb.Select(req, c.svrs)
	if 0 > index {
		return ""
	}

	e := collection.Get(c.svrs, index)
	if nil == e {
		return ""
	}

	id, _ := e.Value.(string)
	return id
}

type serverRuntime struct {
	tw               *goetty.TimeoutWheel
	meta             *model.Server
	status           model.Status
	heathTimeout     goetty.Timeout
	checkFailCount   int
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

func (s *serverRuntime) updateMeta(meta *model.Server) {
	s.meta = meta
}

func (s *serverRuntime) getCheckURL() string {
	return fmt.Sprintf("%s://%s%s", s.meta.Schema, s.meta.Addr, s.meta.HeathCheck.Path)
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
	s.status = status
}

func (s *serverRuntime) isCircuitStatus(target model.Circuit) bool {
	return s.circuit == target
}

func (s *serverRuntime) circuitToClose() {
	if s.meta.CircuitBreaker == nil ||
		s.circuit == model.CircuitClose {
		return
	}

	s.circuit = model.CircuitClose
	log.Warnf("server <%s> change to close", s.meta.ID)
	s.tw.Schedule(s.meta.CircuitBreaker.CloseTimeout, s.circuitToHalf, nil)
}

func (s *serverRuntime) circuitToOpen() {
	if s.meta.CircuitBreaker == nil ||
		s.circuit == model.CircuitOpen ||
		s.circuit != model.CircuitHalf {
		return
	}

	s.circuit = model.CircuitOpen
	log.Infof("server <%s> change to open", s.meta.ID)
}

func (s *serverRuntime) circuitToHalf(arg interface{}) {
	if s.meta.CircuitBreaker != nil {
		s.circuit = model.CircuitOpen
		log.Warnf("server <%s> change to half", s.meta.ID)
	}
}
