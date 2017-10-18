package goetty

import (
	"errors"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

// Timeout Hash Wheel / Calendar Queue implementation of a callout wheel for timeouts.
// Ticking happens from a time.Ticker

const (
	defaultTickInterval = time.Millisecond
	defaultNumBuckets   = 2048

	cacheline    = 64
	bitsInUint64 = 64
)

const (
	// states of the TimeoutWheel
	stopped = iota
	stopping
	running
)

const (
	// states of a Timeout
	timeoutInactive = iota
	timeoutExpired
	timeoutActive
)

var (
	// ErrSystemStopped is returned when a user tries to schedule a timeout after stopping the
	// timeout system.
	ErrSystemStopped = errors.New("Timeout System is stopped")

	// 4*(NumCPU rounded down to the next power of 2.)
	defaultPoolSize = uint64(1 << uint(findMSB(uint64(runtime.NumCPU()))+2))
)

// Timeout represents a single timeout function pending expiration.
type Timeout struct {
	generation uint64
	timeout    *timeout
}

// Stop stops the scheduled timeout so that the callback will not be called. It returns true if it
// successfully canceled
func (t *Timeout) Stop() bool {
	if t.timeout == nil {
		return false
	}

	t.timeout.mtx.Lock()
	if t.timeout.generation != t.generation || t.timeout.state != timeoutActive {
		t.timeout.mtx.Unlock()
		return false
	}

	t.timeout.removeLocked()
	t.timeout.wheel.putTimeoutLocked(t.timeout)
	t.timeout.mtx.Unlock()
	return true
}

type timeout struct {
	mtx       *paddedMutex
	expireCb  func(interface{})
	expireArg interface{}
	deadline  uint64

	// list pointers for the freelist/buckets of the queue. The list is implemented as a forward
	// pointer and a pointer to the address of the previous next field. It is doubly linked in this
	// manner so that removal does not require traversal. It can only be iterated in the forward.
	next  *timeout
	prev  **timeout
	state int32

	wheel      *TimeoutWheel
	generation uint64
}

type timeoutList struct {
	lastTick uint64
	head     *timeout
}

func (t *timeout) prependLocked(list *timeoutList) {
	if list.head != nil {
		t.prev = list.head.prev
		list.head.prev = &t.next
	} else {
		t.prev = &list.head
	}
	t.next = list.head
	list.head = t
}

func (t *timeout) removeLocked() {
	if t.next != nil {
		t.next.prev = t.prev
	}

	*t.prev = t.next
	t.next = nil
	t.prev = nil
}

// TimeoutWheel is a bucketed collection of Timeouts that have a deadline in the future.
// (current tick granularity is 1ms).
type TimeoutWheel struct {
	// ticks is an atomic
	ticks uint64

	// buckets[i] and freelists[i] is locked by mtxPool[i&poolMask]
	mtxPool      []paddedMutex
	bucketMask   uint64
	poolMask     uint64
	tickInterval time.Duration
	buckets      []timeoutList
	freelists    []timeoutList

	state     int
	calloutCh chan timeoutList
	done      chan struct{}
}

// Option is a configuration option to NewTimeoutWheel
type Option func(*opts)

type opts struct {
	tickInterval time.Duration
	size         uint64
	poolsize     uint64
}

// sync.Mutex padded to a cache line to keep the locks from false sharing with each other
type paddedMutex struct {
	sync.Mutex
	_ [cacheline - unsafe.Sizeof(sync.Mutex{})]byte
}

// WithTickInterval sets the frequency of ticks.
func WithTickInterval(interval time.Duration) Option {
	return func(opts *opts) { opts.tickInterval = interval }
}

// WithBucketsExponent sets the number of buckets in the hash table.
func WithBucketsExponent(bucketExp uint) Option {
	return func(opts *opts) {
		opts.size = uint64(1 << bucketExp)
	}
}

// WithLocksExponent sets the number locks in the lockpool used to lock the time buckets. If the
// number is greater than the number of buckets, the number of buckets will be used instead.
func WithLocksExponent(lockExp uint) Option {
	return func(opts *opts) {
		opts.poolsize = uint64(1 << lockExp)
	}
}

// NewTimeoutWheel creates and starts a new TimeoutWheel collection.
func NewTimeoutWheel(options ...Option) *TimeoutWheel {
	opts := &opts{
		tickInterval: defaultTickInterval,
		size:         defaultNumBuckets,
		poolsize:     defaultPoolSize,
	}

	for _, option := range options {
		option(opts)
	}

	poolsize := opts.poolsize
	if opts.size < opts.poolsize {
		poolsize = opts.size
	}

	t := &TimeoutWheel{
		mtxPool:      make([]paddedMutex, poolsize),
		freelists:    make([]timeoutList, poolsize),
		state:        stopped,
		poolMask:     poolsize - 1,
		buckets:      make([]timeoutList, opts.size),
		tickInterval: opts.tickInterval,
		bucketMask:   opts.size - 1,
		ticks:        0,
	}
	t.Start()
	return t
}

// Start starts a stopped timeout wheel. Subsequent calls to Start panic.
func (t *TimeoutWheel) Start() {
	t.lockAllBuckets()
	defer t.unlockAllBuckets()

	for t.state != stopped {
		switch t.state {
		case stopping:
			t.unlockAllBuckets()
			<-t.done
			t.lockAllBuckets()
		case running:
			panic("Tried to start a running TimeoutWheel")
		}
	}

	t.state = running
	t.done = make(chan struct{})
	t.calloutCh = make(chan timeoutList)

	go t.doTick()
	go t.doExpired()
}

// Stop stops tick processing, and deletes any remaining timeouts.
func (t *TimeoutWheel) Stop() {
	t.lockAllBuckets()

	if t.state == running {
		t.state = stopping
		close(t.calloutCh)
		for i := range t.buckets {
			t.freeBucketLocked(t.buckets[i])
		}
	}

	// unlock so the callouts can finish.
	t.unlockAllBuckets()
	<-t.done
}

// Schedule adds a new function to be called after some duration of time has
// elapsed. The returned Timeout can be used to cancel calling the function. If the duration falls
// between two ticks, the latter tick is used.
func (t *TimeoutWheel) Schedule(
	d time.Duration,
	expireCb func(interface{}),
	arg interface{},
) (Timeout, error) {
	dTicks := (d + t.tickInterval - 1) / t.tickInterval
	deadline := atomic.LoadUint64(&t.ticks) + uint64(dTicks)
	timeout := t.getTimeoutLocked(deadline)

	if t.state != running {
		t.putTimeoutLocked(timeout)
		timeout.mtx.Unlock()
		return Timeout{}, ErrSystemStopped
	}

	bucket := &t.buckets[deadline&t.bucketMask]
	timeout.expireCb = expireCb
	timeout.expireArg = arg
	timeout.deadline = deadline
	timeout.state = timeoutActive
	out := Timeout{timeout: timeout, generation: timeout.generation}

	// execute the callback now, return a Timeout that is already Stopped (generation is bumped)
	if bucket.lastTick >= deadline {
		t.putTimeoutLocked(timeout)
		timeout.mtx.Unlock()
		expireCb(arg)
		return out, nil
	}

	timeout.prependLocked(bucket)
	timeout.mtx.Unlock()
	return out, nil
}

// doTick handles the ticker goroutine of the timer system
func (t *TimeoutWheel) doTick() {
	var expiredList timeoutList

	ticker := time.NewTicker(t.tickInterval)
	for range ticker.C {
		atomic.AddUint64(&t.ticks, 1)

		mtx := t.lockBucket(t.ticks)
		if t.state != running {
			mtx.Unlock()
			break
		}

		bucket := &t.buckets[t.ticks&t.bucketMask]
		timeout := bucket.head
		bucket.lastTick = t.ticks

		// find all the expired timeouts in the bucket.
		for timeout != nil {
			next := timeout.next
			if timeout.deadline <= t.ticks {
				timeout.state = timeoutExpired
				timeout.removeLocked()
				timeout.prependLocked(&expiredList)
			}
			timeout = next
		}

		mtx.Unlock()
		if expiredList.head == nil {
			continue
		}

		select {
		case t.calloutCh <- expiredList:
			expiredList.head = nil
		default:
		}
	}

	ticker.Stop()
}

func (t *TimeoutWheel) getTimeoutLocked(deadline uint64) *timeout {
	mtx := &t.mtxPool[deadline&t.poolMask]
	mtx.Lock()
	freelist := &t.freelists[deadline&t.poolMask]
	if freelist.head == nil {
		timeout := &timeout{mtx: mtx, wheel: t}
		return timeout
	}
	timeout := freelist.head
	timeout.removeLocked()
	return timeout
}

func (t *TimeoutWheel) putTimeoutLocked(timeout *timeout) {
	freelist := &t.freelists[timeout.deadline&t.poolMask]
	timeout.state = timeoutInactive
	timeout.generation++
	timeout.prependLocked(freelist)
}

func (t *TimeoutWheel) lockBucket(bucket uint64) *paddedMutex {
	mtx := &t.mtxPool[bucket&t.poolMask]
	mtx.Lock()
	return mtx
}

func (t *TimeoutWheel) lockAllBuckets() {
	for i := range t.mtxPool {
		t.mtxPool[i].Lock()
	}
}

func (t *TimeoutWheel) unlockAllBuckets() {
	for i := len(t.mtxPool) - 1; i >= 0; i-- {
		t.mtxPool[i].Unlock()
	}
}

func (t *TimeoutWheel) freeBucketLocked(head timeoutList) {
	timeout := head.head
	for timeout != nil {
		next := timeout.next
		timeout.removeLocked()
		t.putTimeoutLocked(timeout)
		timeout = next
	}
}

func (t *TimeoutWheel) doExpired() {
	for list := range t.calloutCh {
		timeout := list.head
		for timeout != nil {
			timeout.mtx.Lock()
			next := timeout.next
			expireCb := timeout.expireCb
			expireArg := timeout.expireArg
			t.putTimeoutLocked(timeout)
			timeout.mtx.Unlock()

			if expireCb != nil {
				expireCb(expireArg)
			}
			timeout = next
		}
	}

	t.lockAllBuckets()
	t.state = stopped
	t.unlockAllBuckets()
	close(t.done)
}

// return the bit position of the most significant bit of the number passed in (base zero)
func findMSB(value uint64) int {
	for i := bitsInUint64 - 1; i >= 0; i-- {
		if value&(1<<uint(i)) != 0 {
			return int(i)
		}
	}
	return -1
}
