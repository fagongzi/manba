package atomic

import (
	"sync/atomic"
)

// Int64 atomic int32
type Int64 struct {
	value int64
}

// Set atomic set int64
func (u *Int64) Set(value int64) {
	atomic.StoreInt64(&u.value, value)
}

// Get returns atomic int64
func (u *Int64) Get() int64 {
	return atomic.LoadInt64(&u.value)
}

// Incr incr atomic int64
func (u *Int64) Incr() int64 {
	return u.Add(1)
}

// Add add atomic int32
func (u *Int64) Add(value int64) int64 {
	return atomic.AddInt64(&u.value, value)
}
