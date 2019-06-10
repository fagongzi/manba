package util

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/fagongzi/goetty"
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, 1, mlen(&ans.points),
		fmt.Sprintf("expect 1 points but %d", mlen(&ans.points)))

	assert.Equal(t, 1, mlen(&ans.recentlyPoints),
		fmt.Sprintf("expect 1 recently points but %d", mlen(&ans.recentlyPoints)))

	m, _ := ans.recentlyPoints.Load(key)
	assert.Equal(t, 1, mlen(m.(*sync.Map)),
		fmt.Sprintf("expect 1 recently points but %d", mlen(m.(*sync.Map))))
}

func TestRemoveTarget(t *testing.T) {
	key := uint64(1)
	tw := goetty.NewTimeoutWheel(goetty.WithTickInterval(time.Millisecond * 10))
	ans := NewAnalysis(tw)
	ans.AddTarget(key, time.Millisecond*10)
	ans.RemoveTarget(key)

	assert.Equal(t, 0, mlen(&ans.points),
		fmt.Sprintf("expect 0 points but %d", mlen(&ans.points)))

	assert.Equal(t, 0, mlen(&ans.recentlyPoints),
		fmt.Sprintf("expect 0 recently points but %d", mlen(&ans.recentlyPoints)))
}
