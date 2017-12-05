package goetty

import (
	"math/rand"
	"sync/atomic"
	"testing"
	"time"
)

func TestFindMSB(t *testing.T) {
	if findMSB(2048) != 11 {
		t.Fail()
	}
	if findMSB(2049) != 11 {
		t.Fail()
	}
	if findMSB(0) != -1 {
		t.Fail()
	}
	if findMSB(1) != 0 {
		t.Fail()
	}
	if findMSB(^uint64(0)) != 63 {
		t.Fail()
	}
}

func TestLinkedList(t *testing.T) {
	var list timeoutList

	timeout1 := &timeout{}
	timeout2 := &timeout{}

	timeout1.prependLocked(&list)
	timeout2.prependLocked(&list)

	if list.head != timeout2 {
		t.FailNow()
		return
	}

	timeout2.removeLocked()
	if list.head != timeout1 {
		t.FailNow()
		return
	}
	timeout1.removeLocked()
	if list.head != nil {
		t.FailNow()
		return
	}
}

func TestStartStop(t *testing.T) {
	tw := NewTimeoutWheel()
	for i := 0; i < 10; i++ {
		tw.Stop()
		time.Sleep(5 * time.Millisecond)
		tw.Start()
	}
	tw.Stop()
}

func TestStartStopConcurrent(t *testing.T) {
	tw := NewTimeoutWheel()

	for i := 0; i < 10; i++ {
		tw.Stop()
		go tw.Start()
		time.Sleep(1 * time.Millisecond)
	}
	tw.Stop()
}

func TestExpire(t *testing.T) {
	tw := NewTimeoutWheel()
	ch := make(chan int, 3)

	_, err := tw.Schedule(20*time.Millisecond, func(_ interface{}) { ch <- 20 }, nil)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = tw.Schedule(10*time.Millisecond, func(_ interface{}) { ch <- 10 }, nil)
	if err != nil {
		t.Error(err)
		return
	}

	_, err = tw.Schedule(5*time.Millisecond, func(_ interface{}) { ch <- 5 }, nil)
	if err != nil {
		t.Error(err)
		return
	}

	output := make([]int, 0, 3)
	for i := 0; i < 3; i++ {
		d := time.Duration(((i+1)*5)+1) * time.Millisecond * 2
		select {
		case d := <-ch:
			output = append(output, d)
		case <-time.After(d):
			t.Error("after timeout")
			return
		}
	}
	if output[0] != 5 || output[1] != 10 || output[2] != 20 {
		t.Fail()
	}
	tw.Stop()
}

func TestTimeoutStop(t *testing.T) {
	tw := NewTimeoutWheel()
	ch := make(chan struct{})

	timeout, err := tw.Schedule(20*time.Millisecond, func(_ interface{}) { close(ch) }, nil)
	if err != nil {
		t.Fail()
		return
	}

	timeout.Stop()

	select {
	case <-time.After(30 * time.Millisecond):
	case <-ch:
		t.Fail()
	}

	if timeout.generation == timeout.timeout.generation {
		t.Fail()
	}
}

func TestScheduleExpired(t *testing.T) {
	ch := make(chan struct{})
	tw := NewTimeoutWheel()
	tw.Stop()
	tw.ticks = 0
	tw.state = running

	tw.buckets[0].lastTick = 1
	timeout, _ := tw.Schedule(0, func(_ interface{}) { close(ch) }, nil)

	select {
	case <-ch:
	default:
		t.Fail()
	}

	if timeout.generation == timeout.timeout.generation {
		t.Fail()
	}
}

func BenchmarkInsertDelete(b *testing.B) {
	tw := NewTimeoutWheel(WithLocksExponent(11))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			timeout, _ := tw.Schedule(time.Second+time.Millisecond*time.Duration(r.Intn(1000)),
				nil, nil)
			timeout.Stop()
		}
	})
	tw.Stop()
}

func BenchmarkStdlibInsertDelete(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			timer := time.AfterFunc(time.Second+time.Millisecond*time.Duration(r.Intn(1000)), nil)
			timer.Stop()
		}
	})
}

func BenchmarkInsertDeleteWithPending(b *testing.B) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	tw := NewTimeoutWheel(WithLocksExponent(11))
	for i := 0; i < 8192; i++ {
		tw.Schedule(time.Second*10+time.Millisecond*time.Duration(r.Intn(1000)), nil, nil)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		for pb.Next() {
			timeout, _ := tw.Schedule(time.Second+time.Millisecond*time.Duration(r.Intn(1000)),
				nil, nil)
			timeout.Stop()
		}
	})
	tw.Stop()
}

func BenchmarkReplacement(b *testing.B) {
	// uses resevoir sampling to randomly replace timers, so that some have a
	// chance of expiring
	tw := NewTimeoutWheel(WithLocksExponent(11))
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		mine := make([]Timeout, 0, 10000)
		n, k := 0, 0

		// fill
		for pb.Next() && len(mine) < cap(mine) {
			dur := time.Second + time.Millisecond*time.Duration(r.Intn(1000))
			n++

			timeout, _ := tw.Schedule(dur, nil, nil)
			mine = append(mine, timeout)
			k++
		}

		// replace randomly
		for pb.Next() {
			dur := time.Second + time.Millisecond*time.Duration(r.Intn(1000))
			n++

			if r.Float32() <= float32(k)/float32(n) {
				i := r.Intn(len(mine))
				timeout := mine[i]
				timeout.Stop()
				mine[i], _ = tw.Schedule(dur, nil, nil)
				k++
			}
		}
	})
	tw.Stop()
}

func BenchmarkExpiration(b *testing.B) {
	tw := NewTimeoutWheel()
	d := time.Millisecond
	b.ResetTimer()
	var sum int64
	f := func(interface{}) {
		atomic.AddInt64(&sum, 1)
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tw.Schedule(d, f, nil)
		}
	})
	//time.Sleep(5 * time.Millisecond)
	tw.Stop()
}
