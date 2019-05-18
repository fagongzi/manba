package goetty

import (
	"sync"
)

const (
	// KB kb
	KB = 1024
	// MB mb
	MB = 1024 * 1024
)

var (
	lock       sync.Mutex
	mp         Pool
	defaultMin = 256
	defaultMax = 8 * MB
)

func getDefaultMP() Pool {
	lock.Lock()
	if mp == nil {
		useDefaultMemPool()
	}
	lock.Unlock()

	return mp
}

func useDefaultMemPool() {
	mp = NewSyncPool(
		defaultMin,
		defaultMax,
		2,
	)
}

// UseMemPool use the custom mem pool
func UseMemPool(min, max int) {
	mp = NewSyncPool(
		min,
		max,
		2,
	)
}
