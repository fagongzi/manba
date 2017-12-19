package util

import (
	"sync"
	"time"

	"github.com/fagongzi/goetty"
	"github.com/fagongzi/log"
	"github.com/fagongzi/util/atomic"
)

type point struct {
	requests          atomic.Int64
	rejects           atomic.Int64
	failure           atomic.Int64
	successed         atomic.Int64
	continuousFailure atomic.Int64

	costs atomic.Int64
	max   atomic.Int64
	min   atomic.Int64
}

func (p *point) dump(target *point) {
	target.requests.Set(p.requests.Get())
	target.rejects.Set(p.rejects.Get())
	target.failure.Set(p.failure.Get())
	target.successed.Set(p.successed.Get())
	target.max.Set(p.max.Get())
	target.min.Set(p.min.Get())
	target.costs.Set(p.costs.Get())

	p.min.Set(0)
	p.max.Set(0)
}

// Analysis analysis struct
type Analysis struct {
	sync.RWMutex

	tw             *goetty.TimeoutWheel
	points         map[uint64]*point
	recentlyPoints map[uint64]map[time.Duration]*Recently
}

// Recently recently point data
type Recently struct {
	key       uint64
	timeout   goetty.Timeout
	period    time.Duration
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

func newRecently(key uint64, period time.Duration) *Recently {
	return &Recently{
		key:     key,
		prev:    newPoint(),
		current: newPoint(),
		period:  period,
	}
}

func newPoint() *point {
	return &point{}
}

// NewAnalysis returns a Analysis
func NewAnalysis(tw *goetty.TimeoutWheel) *Analysis {
	return &Analysis{
		points:         make(map[uint64]*point),
		recentlyPoints: make(map[uint64]map[time.Duration]*Recently),
		tw:             tw,
	}
}

// RemoveTarget remove analysis point on a key
func (a *Analysis) RemoveTarget(key uint64) {
	a.Lock()
	defer a.Unlock()

	if m, ok := a.recentlyPoints[key]; ok {
		for _, r := range m {
			r.timeout.Stop()
		}
	}

	delete(a.points, key)
	delete(a.recentlyPoints, key)
}

// AddTarget add analysis point on a key
func (a *Analysis) AddTarget(key uint64, interval time.Duration) {
	a.Lock()
	defer a.Unlock()

	if interval == 0 {
		return
	}

	if _, ok := a.points[key]; !ok {
		a.points[key] = &point{}
	}

	if _, ok := a.recentlyPoints[key]; !ok {
		a.recentlyPoints[key] = make(map[time.Duration]*Recently)
	}

	if _, ok := a.recentlyPoints[key][interval]; ok {
		log.Infof("analysis: already added, key=<%s> interval=<%s>",
			key,
			interval)
		return
	}

	recently := newRecently(key, interval)
	a.recentlyPoints[key][interval] = recently

	log.Infof("analysis: added, key=<%d> interval=<%s>",
		key,
		interval)

	t, _ := a.tw.Schedule(interval, a.recentlyTimeout, recently)
	recently.timeout = t
}

// GetRecentlyRequestCount return the server request count in spec duration
func (a *Analysis) GetRecentlyRequestCount(server uint64, interval time.Duration) int {
	a.RLock()
	defer a.RUnlock()

	point := a.getPoint(server, interval)
	if point == nil {
		return 0
	}

	return int(point.requests)
}

// GetRecentlyMax return max latency in spec secs
func (a *Analysis) GetRecentlyMax(server uint64, interval time.Duration) int {
	a.RLock()
	defer a.RUnlock()

	point := a.getPoint(server, interval)
	if point == nil {
		return 0
	}

	return int(point.max)
}

// GetRecentlyMin return min latency in spec duration
func (a *Analysis) GetRecentlyMin(server uint64, interval time.Duration) int {
	a.RLock()
	defer a.RUnlock()

	point := a.getPoint(server, interval)
	if point == nil {
		return 0
	}

	return int(point.min)
}

// GetRecentlyAvg return avg latency in spec secs
func (a *Analysis) GetRecentlyAvg(server uint64, interval time.Duration) int {
	a.RLock()
	defer a.RUnlock()

	point := a.getPoint(server, interval)
	if point == nil {
		return 0
	}

	return int(point.avg)
}

// GetQPS return qps in spec duration
func (a *Analysis) GetQPS(server uint64, interval time.Duration) int {
	a.RLock()
	defer a.RUnlock()

	point := a.getPoint(server, interval)
	if point == nil {
		return 0
	}

	return int(point.qps)
}

// GetRecentlyRejectCount return reject count in spec duration
func (a *Analysis) GetRecentlyRejectCount(server uint64, interval time.Duration) int {
	a.RLock()
	defer a.RUnlock()

	point := a.getPoint(server, interval)
	if point == nil {
		return 0
	}

	return int(point.rejects)
}

// GetRecentlyRequestSuccessedCount return successed request count in spec secs
func (a *Analysis) GetRecentlyRequestSuccessedCount(server uint64, interval time.Duration) int {
	a.RLock()
	defer a.RUnlock()

	point := a.getPoint(server, interval)
	if point == nil {
		return 0
	}

	return int(point.successed)
}

// GetRecentlyRequestFailureCount return failure request count in spec duration
func (a *Analysis) GetRecentlyRequestFailureCount(server uint64, interval time.Duration) int {
	a.RLock()
	defer a.RUnlock()

	point := a.getPoint(server, interval)
	if point == nil {
		return 0
	}

	return int(point.failure)
}

// GetContinuousFailureCount return Continuous failure request count in spec secs
func (a *Analysis) GetContinuousFailureCount(server uint64) int {
	a.RLock()
	defer a.RUnlock()

	p, ok := a.points[server]
	if !ok {
		return 0
	}

	return int(p.continuousFailure.Get())
}

// Reject incr reject count
func (a *Analysis) Reject(key uint64) {
	a.Lock()
	p := a.points[key]
	p.rejects.Incr()
	a.Unlock()
}

// Failure incr failure count
func (a *Analysis) Failure(key uint64) {
	a.Lock()
	p := a.points[key]
	p.failure.Incr()
	p.continuousFailure.Incr()
	a.Unlock()
}

// Request incr request count
func (a *Analysis) Request(key uint64) {
	a.Lock()
	p := a.points[key]
	p.requests.Incr()
	a.Unlock()
}

// Response incr successed count
func (a *Analysis) Response(key uint64, cost int64) {
	a.Lock()
	p := a.points[key]
	p.successed.Incr()
	p.costs.Add(cost)
	p.continuousFailure.Set(0)

	if p.max.Get() < cost {
		p.max.Set(cost)
	}

	if p.min.Get() == 0 || p.min.Get() > cost {
		p.min.Set(cost)
	}
	a.Unlock()
}

func (a *Analysis) getPoint(key uint64, interval time.Duration) *Recently {
	points, ok := a.recentlyPoints[key]
	if !ok {
		return nil
	}

	point, ok := points[interval]
	if !ok {
		return nil
	}

	return point
}

func (a *Analysis) recentlyTimeout(arg interface{}) {
	recently := arg.(*Recently)

	a.RLock()
	if p, ok := a.points[recently.key]; ok {
		recently.record(p)
		t, _ := a.tw.Schedule(recently.period, a.recentlyTimeout, recently)
		recently.timeout = t
	}
	a.RUnlock()
}

func (r *Recently) record(p *point) {
	if r.dumpCurr {
		p.dump(r.current)
		r.calc()
	} else {
		p.dump(r.prev)
	}

	r.dumpCurr = !r.dumpCurr
}

func (r *Recently) calc() {
	r.requests = r.current.requests.Get() - r.prev.requests.Get()

	if r.requests < 0 {
		r.requests = 0
	}

	r.successed = r.current.successed.Get() - r.prev.successed.Get()

	if r.successed < 0 {
		r.successed = 0
	}

	r.failure = r.current.failure.Get() - r.prev.failure.Get()

	if r.failure < 0 {
		r.failure = 0
	}

	r.rejects = r.current.rejects.Get() - r.prev.rejects.Get()

	if r.rejects < 0 {
		r.rejects = 0
	}

	r.max = r.current.max.Get()

	if r.max < 0 {
		r.max = 0
	} else {
		r.max = int64(r.max / 1000 / 1000)
	}

	r.min = r.current.min.Get()

	if r.min < 0 {
		r.min = 0
	} else {
		r.min = int64(r.min / 1000 / 1000)
	}

	costs := r.current.costs.Get() - r.prev.costs.Get()

	if r.requests == 0 {
		r.avg = 0
	} else {
		r.avg = int64(costs / 1000 / 1000 / r.requests)
	}

	if r.successed > r.requests {
		r.qps = int(r.requests / int64(r.period/time.Second))
	} else {
		r.qps = int(r.successed / int64(r.period/time.Second))
	}

}
