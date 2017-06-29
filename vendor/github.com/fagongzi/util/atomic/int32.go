package atomic

import (
	"sync/atomic"
)

// Int32 atomic int32
type Int32 struct {
	value int32
}

// Set atomic set int32
func (u *Int32) Set(value int32) {
	atomic.StoreInt32(&u.value, value)
}

// Get returns atomic int32
func (u *Int32) Get() int32 {
	return atomic.LoadInt32(&u.value)
}

// Incr incr atomic int32
func (u *Int32) Incr() int32 {
	return u.Add(1)
}

// Add add atomic int32
func (u *Int32) Add(value int32) int32 {
	return atomic.AddInt32(&u.value, value)
}
