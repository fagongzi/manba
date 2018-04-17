package util

import (
	"testing"
	"time"

	"github.com/fagongzi/goetty"
)

func TestAddTarget(t *testing.T) {
	key := uint64(1)
	tw := goetty.NewTimeoutWheel(goetty.WithTickInterval(time.Millisecond * 10))
	ans := NewAnalysis(tw)
	ans.AddTarget(key, time.Millisecond*10)
	if 1 != len(ans.points) {
		t.Errorf("add target failed")
		return
	}

	if 1 != len(ans.recentlyPoints) {
		t.Errorf("add target failed")
		return
	}

	if 1 != len(ans.recentlyPoints[key]) {
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

	if 0 != len(ans.points) {
		t.Errorf("remove target failed")
		return
	}

	if 0 != len(ans.recentlyPoints) {
		t.Errorf("remove target failed")
		return
	}
}
