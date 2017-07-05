package goetty

import (
	"testing"
	"time"
)

func TestSimpleTimeWheelInOnePeriod(t *testing.T) {
	called := false
	w := NewSimpleTimeWheel(time.Millisecond*10, 60)
	w.Start()

	w.Add(time.Millisecond*10, func(key string) {
		called = true
	})

	time.Sleep(time.Millisecond * 20)

	if !called {
		t.Error("failure.")
	}
}

func TestSimpleTimeWheelInOnePeriodNot(t *testing.T) {
	called := false
	w := NewSimpleTimeWheel(time.Millisecond, 60)
	w.Start()

	w.Add(time.Millisecond*10, func(key string) {
		called = true
	})

	time.Sleep(time.Millisecond * 9)

	if called {
		t.Error("failure.")
	}
}

func TestSimpleTimeWheelInMorePeriod(t *testing.T) {
	called := false
	w := NewSimpleTimeWheel(time.Millisecond, 60)
	w.Start()

	w.Add(time.Millisecond*100, func(key string) {
		called = true
	})

	time.Sleep(time.Millisecond * 200)

	if !called {
		t.Error("failure.")
	}
}

func TestSimpleTimeWheelInMorePeriodNot(t *testing.T) {
	called := false
	w := NewSimpleTimeWheel(time.Millisecond, 60)
	w.Start()

	w.Add(time.Millisecond*100, func(key string) {
		called = true
	})

	time.Sleep(time.Millisecond * 99)

	if called {
		t.Error("failure.")
	}
}

func TestHashTimeWheelInOnePeriod(t *testing.T) {
	called := false
	w := NewHashedTimeWheel(time.Millisecond*10, 60, 2)
	w.Start()

	w.Add(time.Millisecond*10, func(key string) {
		called = true
	})

	time.Sleep(time.Millisecond * 20)

	if !called {
		t.Error("failure.")
	}
}

func TestHashTimeWheelInOnePeriodNot(t *testing.T) {
	called := false
	w := NewHashedTimeWheel(time.Millisecond, 60, 2)
	w.Start()

	w.Add(time.Millisecond*10, func(key string) {
		called = true
	})

	time.Sleep(time.Millisecond * 9)

	if called {
		t.Error("failure.")
	}
}

func TestHashTimeWheelInMorePeriod(t *testing.T) {
	called := false
	w := NewHashedTimeWheel(time.Millisecond*10, 60, 2)
	w.Start()

	w.Add(time.Millisecond*100, func(key string) {
		called = true
	})

	time.Sleep(time.Millisecond * 110)

	if !called {
		t.Error("failure.")
	}
}

func TestHashTimeWheelInMorePeriodNot(t *testing.T) {
	called := false
	w := NewHashedTimeWheel(time.Millisecond, 60, 2)
	w.Start()

	w.Add(time.Millisecond*100, func(key string) {
		called = true
	})

	time.Sleep(time.Millisecond * 99)

	if called {
		t.Error("failure.")
	}
}
