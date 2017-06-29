package atomic

import (
	"sync/atomic"
)

// Uint64 atomic uint64
type Uint64 struct {
	value uint64
}

// Set atomic set uint64
func (u *Uint64) Set(value uint64) {
	atomic.StoreUint64(&u.value, value)
}

// Get returns atomic uint64
func (u *Uint64) Get() uint64 {
	return atomic.LoadUint64(&u.value)
}

// Incr incr atomic uint64
func (u *Uint64) Incr() uint64 {
	return u.Add(1)
}

// Add add atomic uint64
func (u *Uint64) Add(value uint64) uint64 {
	return atomic.AddUint64(&u.value, value)
}
