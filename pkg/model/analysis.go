package model

import (
	"github.com/CodisLabs/codis/pkg/utils/atomic2"
	"github.com/CodisLabs/codis/pkg/utils/log"
	"time"
)

type point struct {
	requests          atomic2.Int64
	rejects           atomic2.Int64
	failure           atomic2.Int64
	successed         atomic2.Int64
	continuousFailure atomic2.Int64

	costs atomic2.Int64
	max   atomic2.Int64
	min   atomic2.Int64
}

func (self *point) dump(target *point) {
	target.requests.Set(self.requests.Get())
	target.rejects.Set(self.rejects.Get())
	target.failure.Set(self.failure.Get())
	target.successed.Set(self.successed.Get())
	target.max.Set(self.max.Get())
	target.min.Set(self.min.Get())
	target.costs.Set(self.costs.Get())

	self.min.Set(0)
	self.max.Set(0)
}

type Analysis struct {
	points         map[string]*point
	recentlyPoints map[string]map[int]*Recently
}

type Recently struct {
	period    int64
	prev      *point
	current   *point
	dumpCurr  bool
	qps       int
	requests  int64
	successed int64
	failure   int64
	rejects   int64
	max       int64
	min       int64
	avg       int64
}

func newRecently(period int64) *Recently {
	return &Recently{
		prev:    newPoint(),
		current: newPoint(),
		period:  period,
	}
}

func newPoint() *point {
	return &point{}
}

func newAnalysis() *Analysis {
	return &Analysis{
		points:         make(map[string]*point),
		recentlyPoints: make(map[string]map[int]*Recently),
	}
}

func (self *Recently) record(p *point) {
	if self.dumpCurr {
		p.dump(self.current)
		self.calc()
	} else {
		p.dump(self.prev)
	}

	self.dumpCurr = !self.dumpCurr
}

func (self *Recently) calc() {

	self.requests = self.current.requests.Get() - self.prev.requests.Get()

	if self.requests < 0 {
		self.requests = 0
	}

	self.successed = self.current.successed.Get() - self.prev.successed.Get()

	if self.successed < 0 {
		self.successed = 0
	}

	self.failure = self.current.failure.Get() - self.prev.failure.Get()

	if self.failure < 0 {
		self.failure = 0
	}

	self.rejects = self.current.rejects.Get() - self.prev.rejects.Get()

	if self.rejects < 0 {
		self.rejects = 0
	}

	self.max = self.current.max.Get()

	if self.max < 0 {
		self.max = 0
	} else {
		self.max = int64(self.max / 1000 / 1000)
	}

	self.min = self.current.min.Get()

	if self.min < 0 {
		self.min = 0
	} else {
		self.min = int64(self.min / 1000 / 1000)
	}

	costs := self.current.costs.Get() - self.prev.costs.Get()

	if self.requests == 0 {
		self.avg = 0
	} else {
		self.avg = int64(costs / 1000 / 1000 / self.requests)
	}

	if self.successed > self.requests {
		self.qps = int(self.requests / self.period)
	} else {
		self.qps = int(self.successed / self.period)
	}

}

func (self *Analysis) AddRecentCount(key string, secs int) {
	_, ok := self.recentlyPoints[key][secs]
	if ok {
		log.Infof("Analysis already <%s,%d> added", key, secs)
		return
	}

	recently := newRecently(int64(secs))
	self.recentlyPoints[key][secs] = recently
	timer := time.NewTicker(time.Duration(secs) * time.Second)

	go func() {
		for {
			// TODO: remove
			<-timer.C

			p, ok := self.points[key]

			if ok {
				recently.record(p)
			}
		}
	}()

	log.Infof("Analysis <%s,%d> added", key, secs)
}

func (self *Analysis) addNewAnalysis(key string) {
	self.points[key] = &point{}
	self.recentlyPoints[key] = make(map[int]*Recently)
}

func (self *Analysis) GetRecentlyRequestCount(server string, secs int) int {
	points, ok := self.recentlyPoints[server]

	if !ok {
		return 0
	}

	point, ok := points[secs]

	if !ok {
		return 0
	}

	return int(point.requests)
}

func (self *Analysis) GetRecentlyMax(server string, secs int) int {
	points, ok := self.recentlyPoints[server]

	if !ok {
		return 0
	}

	point, ok := points[secs]

	if !ok {
		return 0
	}

	return int(point.max)
}

func (self *Analysis) GetRecentlyMin(server string, secs int) int {
	points, ok := self.recentlyPoints[server]

	if !ok {
		return 0
	}

	point, ok := points[secs]

	if !ok {
		return 0
	}

	return int(point.min)
}

func (self *Analysis) GetRecentlyAvg(server string, secs int) int {
	points, ok := self.recentlyPoints[server]

	if !ok {
		return 0
	}

	point, ok := points[secs]

	if !ok {
		return 0
	}

	return int(point.avg)
}

func (self *Analysis) GetQPS(server string, secs int) int {
	points, ok := self.recentlyPoints[server]

	if !ok {
		return 0
	}

	point, ok := points[secs]

	if !ok {
		return 0
	}

	return int(point.qps)
}

func (self *Analysis) GetRecentlyRejectCount(server string, secs int) int {
	points, ok := self.recentlyPoints[server]

	if !ok {
		return 0
	}

	point, ok := points[secs]

	if !ok {
		return 0
	}

	return int(point.rejects)
}

func (self *Analysis) GetRecentlyRequestSuccessedCount(server string, secs int) int {
	points, ok := self.recentlyPoints[server]

	if !ok {
		return 0
	}

	point, ok := points[secs]

	if !ok {
		return 0
	}

	return int(point.successed)
}

func (self *Analysis) GetRecentlyRequestFailureCount(server string, secs int) int {
	points, ok := self.recentlyPoints[server]

	if !ok {
		return 0
	}

	point, ok := points[secs]

	if !ok {
		return 0
	}

	return int(point.failure)
}

func (self *Analysis) GetContinuousFailureCount(server string) int {
	p, ok := self.points[server]

	if !ok {
		return 0
	}

	return int(p.continuousFailure.Get())
}

func (self *Analysis) Reject(key string) {
	p := self.points[key]
	p.rejects.Incr()
}

func (self *Analysis) Failure(key string) {
	p := self.points[key]
	p.failure.Incr()
	p.continuousFailure.Incr()
}

func (self *Analysis) Request(key string) {
	p := self.points[key]
	p.requests.Incr()
}

func (self *Analysis) Response(key string, cost int64) {
	p := self.points[key]
	p.successed.Incr()
	p.costs.Add(cost)
	p.continuousFailure.Set(0)

	if p.max.Get() < cost {
		p.max.Set(cost)
	}

	if p.min.Get() == 0 || p.min.Get() > cost {
		p.min.Set(cost)
	}
}
