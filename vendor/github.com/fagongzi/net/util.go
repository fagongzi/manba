package net

// timer wheel impl, support net read/write timeout
// author:  zhangxu
// date:    2015-6-26
// version: 1.0.0

import (
	"container/ring"
	"github.com/fgrid/uuid"
	"hash/fnv"
	"math"
	"sync"
	"time"
)

// Simple time wheel impl
type SimpleTimeWheel struct {
	duration  time.Duration
	size      int
	timer     *time.Ticker
	list      []*ring.Ring
	callbacks map[string]func(key string)
	currPtr   int
	mutex     *sync.Mutex
}

func NewSimpleTimeWheel(duration time.Duration, size int) *SimpleTimeWheel {
	timeWheel := &SimpleTimeWheel{
		duration:  duration,
		size:      size,
		callbacks: make(map[string]func(key string)),
		currPtr:   0,
		mutex:     &sync.Mutex{},
	}

	timeWheel.init()

	return timeWheel
}

func (t *SimpleTimeWheel) init() {
	t.list = make([]*ring.Ring, t.size)

	for i := 0; i < t.size; i++ {
		t.list[i] = ring.New(1)
	}
}

func (t *SimpleTimeWheel) AddWithId(timeout time.Duration, key string, callback func(key string)) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// calc how time wheel turn.
	needTick := int(float32(timeout)/float32(t.duration) + 0.5)
	index := (t.currPtr + needTick) % t.size

	r := t.list[index]

	if r.Value != nil {
		r.Link(&ring.Ring{
			Value: key,
		})
	} else {
		r.Value = key
	}

	t.callbacks[key] = callback
}

func (t *SimpleTimeWheel) Add(timeout time.Duration, callback func(key string)) string {
	key := NewKey()
	t.AddWithId(timeout, key, callback)
	return key
}

func (t *SimpleTimeWheel) Cancel(key string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	delete(t.callbacks, key)
}

func (t *SimpleTimeWheel) Start() {
	t.timer = time.NewTicker(t.duration)

	go func() {
		for {
			<-t.timer.C
			t.currPtr++

			if t.currPtr == t.size {
				t.currPtr = 0
			}

			t.mutex.Lock()
			r := t.list[t.currPtr]

			if nil != r && nil != r.Value {
				l := r.Len()
				for i := 0; i < l; i++ {
					key, _ := r.Value.(string)
					f, _ := t.callbacks[key]
					if nil != f {
						delete(t.callbacks, key)
						go f(key) // 防止f函数执行过长，导致阻塞
					}

					r.Value = ""
					r = r.Next()
				}
			}

			t.list[t.currPtr] = ring.New(1)
			t.mutex.Unlock()
		}
	}()
}

func (t *SimpleTimeWheel) Stop() {
	if nil != t.timer {
		t.timer.Stop()
	}

	t.currPtr = 0
	t.list = nil

	t.init()
}

type HashedTimeWheel struct {
	mask        int
	wheelBucket []*SimpleTimeWheel
}

func NewHashedTimeWheel(duration time.Duration, size int, powOf2 int) *HashedTimeWheel {
	max := int(math.Pow(2.0, float64(powOf2)))
	h := &HashedTimeWheel{
		mask:        max - 1,
		wheelBucket: make([]*SimpleTimeWheel, max),
	}

	h.init(duration, size, max)

	return h
}

func (h *HashedTimeWheel) Add(timeout time.Duration, callback func(key string)) string {
	key := NewKey()
	h.AddWithId(timeout, key, callback)
	return key
}

func (h *HashedTimeWheel) AddWithId(timeout time.Duration, key string, callback func(key string)) string {
	index := hashCode(key) & h.mask
	h.wheelBucket[index].AddWithId(timeout, key, callback)
	return key
}

func (h *HashedTimeWheel) Cancel(key string) {
	index := hashCode(key) & h.mask
	h.wheelBucket[index].Cancel(key)
}

func (h *HashedTimeWheel) init(duration time.Duration, size int, max int) {
	for i := 0; i < max; i++ {
		h.wheelBucket[i] = NewSimpleTimeWheel(duration, size)
	}
}

func (h *HashedTimeWheel) Start() {
	for i := 0; i < len(h.wheelBucket); i++ {
		go h.wheelBucket[i].Start()
	}
}

func (h *HashedTimeWheel) Stop() {
	for i := 0; i < len(h.wheelBucket); i++ {
		go h.wheelBucket[i].Stop()
	}
}

func NewKey() string {
	return uuid.NewV4().String()
}

func hashCode(v string) int {
	h := fnv.New32a()
	h.Write([]byte(v))
	code := h.Sum32()
	return int(code)
}

