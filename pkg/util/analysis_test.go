package util

import (
	"sync"
	"testing"
	"time"

	"github.com/fagongzi/goetty"
)

func mlen(m *sync.Map) int {
	c := 0
	m.Range(func(key, value interface{}) bool {
		c++
		return true
	})

	return c
}

func TestAddTarget(t *testing.T) {
	key := uint64(1)
	tw := goetty.NewTimeoutWheel(goetty.WithTickInterval(time.Millisecond * 10))
	ans := NewAnalysis(tw)
	ans.AddTarget(key, time.Millisecond*10)
	if 1 != mlen(&ans.points) {
		t.Errorf("add target failed")
		return
	}

	if 1 != mlen(&ans.recentlyPoints) {
		t.Errorf("add target failed")
		return
	}

	m, _ := ans.recentlyPoints.Load(key)
	if 1 != mlen(m.(*sync.Map)) {
		t.Errorf("add target failed")
		return
	}
}

func TestRemoveTarget(t *testing.T) {
	key := uint64(1)
	tw := goetty.NewTimeoutWheel(goetty.WithTickInterval(time.Millisecond * 10))
	ans := NewAnalysis(tw)
	ans.AddTarget(key, time.Millisecond*10)
	ans.RemoveTarget(key)

	if 0 != mlen(&ans.points) {
		t.Errorf("remove target failed")
		return
	}

	if 0 != mlen(&ans.recentlyPoints) {
		t.Errorf("remove target failed")
		return
	}
}
