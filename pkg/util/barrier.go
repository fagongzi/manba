package util

import (
	"math/rand"
	"sync"
	"sync/atomic"
)

var (
	lock        sync.Mutex
	randSources = make(map[int][]int)
)

func createRandSourceByBase(base int) []int {
	lock.Lock()
	defer lock.Unlock()

	if value, ok := randSources[base]; ok {
		return value
	}

	value := make([]int, base, base)
	for i := 0; i < base; i++ {
		value[i] = i
	}

	rand.Shuffle(base, func(i, j int) {
		value[i], value[j] = value[j], value[i]
	})
	randSources[base] = value
	return value
}

// RateBarrier rand barrier
type RateBarrier struct {
	source []int
	op     uint64
	rate   int
	base   int
}

// NewRateBarrier returns a barrier based by 100
func NewRateBarrier(rate int) *RateBarrier {
	return NewRateBarrierBase(rate, 100)
}

// NewRateBarrierBase returns a barrier with base
func NewRateBarrierBase(rate, base int) *RateBarrier {
	return &RateBarrier{
		source: createRandSourceByBase(base),
		rate:   rate,
		base:   base,
	}
}

// Allow returns true if allowed
func (b *RateBarrier) Allow() bool {
	return b.source[int(atomic.AddUint64(&b.op, 1))%b.base] < b.rate
}
