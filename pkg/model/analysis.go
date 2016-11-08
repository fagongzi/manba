package model

import (
	"time"

	"github.com/CodisLabs/codis/pkg/utils/atomic2"
	"github.com/CodisLabs/codis/pkg/utils/log"
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
	points         map[string]*point
	recentlyPoints map[string]map[int]*Recently
}

// Recently recently point data
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
		r.qps = int(r.requests / r.period)
	} else {
		r.qps = int(r.successed / r.period)
	}

}

// AddRecentCount add analysis point on a key
func (a *Analysis) AddRecentCount(key string, secs int) {
	_, ok := a.recentlyPoints[key][secs]
	if ok {
		log.Infof("Analysis already <%s,%d> added", key, secs)
		return
	}

	recently := newRecently(int64(secs))
	a.recentlyPoints[key][secs] = recently
	timer := time.NewTicker(time.Duration(secs) * time.Second)

	go func() {
		for {
			// TODO: remove
			<-timer.C

			p, ok := a.points[key]

			if ok {
				recently.record(p)
			}
		}
	}()

	log.Infof("Analysis <%s,%d> added", key, secs)
}

func (a *Analysis) addNewAnalysis(key string) {
	a.points[key] = &point{}
	a.recentlyPoints[key] = make(map[int]*Recently)
}

// GetRecentlyRequestCount return the server request count in spec seconds
func (a *Analysis) GetRecentlyRequestCount(server string, secs int) int {
	points, ok := a.recentlyPoints[server]

	if !ok {
		return 0
	}

	point, ok := points[secs]

	if !ok {
		return 0
	}

	return int(point.requests)
}

// GetRecentlyMax return max latency in spec secs
func (a *Analysis) GetRecentlyMax(server string, secs int) int {
	points, ok := a.recentlyPoints[server]

	if !ok {
		return 0
	}

	point, ok := points[secs]

	if !ok {
		return 0
	}

	return int(point.max)
}

// GetRecentlyMin return min latency in spec secs
func (a *Analysis) GetRecentlyMin(server string, secs int) int {
	points, ok := a.recentlyPoints[server]

	if !ok {
		return 0
	}

	point, ok := points[secs]

	if !ok {
		return 0
	}

	return int(point.min)
}

// GetRecentlyAvg return avg latency in spec secs
func (a *Analysis) GetRecentlyAvg(server string, secs int) int {
	points, ok := a.recentlyPoints[server]

	if !ok {
		return 0
	}

	point, ok := points[secs]

	if !ok {
		return 0
	}

	return int(point.avg)
}

// GetQPS return qps in spec secs
func (a *Analysis) GetQPS(server string, secs int) int {
	points, ok := a.recentlyPoints[server]

	if !ok {
		return 0
	}

	point, ok := points[secs]

	if !ok {
		return 0
	}

	return int(point.qps)
}

// GetRecentlyRejectCount return reject count in spec secs
func (a *Analysis) GetRecentlyRejectCount(server string, secs int) int {
	points, ok := a.recentlyPoints[server]

	if !ok {
		return 0
	}

	point, ok := points[secs]

	if !ok {
		return 0
	}

	return int(point.rejects)
}

// GetRecentlyRequestSuccessedCount return successed request count in spec secs
func (a *Analysis) GetRecentlyRequestSuccessedCount(server string, secs int) int {
	points, ok := a.recentlyPoints[server]

	if !ok {
		return 0
	}

	point, ok := points[secs]

	if !ok {
		return 0
	}

	return int(point.successed)
}

// GetRecentlyRequestFailureCount return failure request count in spec secs
func (a *Analysis) GetRecentlyRequestFailureCount(server string, secs int) int {
	points, ok := a.recentlyPoints[server]

	if !ok {
		return 0
	}

	point, ok := points[secs]

	if !ok {
		return 0
	}

	return int(point.failure)
}

// GetContinuousFailureCount return Continuous failure request count in spec secs
func (a *Analysis) GetContinuousFailureCount(server string) int {
	p, ok := a.points[server]

	if !ok {
		return 0
	}

	return int(p.continuousFailure.Get())
}

// Reject incr reject count
func (a *Analysis) Reject(key string) {
	p := a.points[key]
	p.rejects.Incr()
}

// Failure incr failure count
func (a *Analysis) Failure(key string) {
	p := a.points[key]
	p.failure.Incr()
	p.continuousFailure.Incr()
}

// Request incr request count
func (a *Analysis) Request(key string) {
	p := a.points[key]
	p.requests.Incr()
}

// Response incr successed count
func (a *Analysis) Response(key string, cost int64) {
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
}
