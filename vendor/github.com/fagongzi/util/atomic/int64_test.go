package atomic

import (
	"testing"
)

func TestInt64SetAndGet(t *testing.T) {
	var v Int64
	v.Set(10)

	got := v.Get()

	if got != int64(10) {
		t.Errorf("failed, got=<%d> expect=<%d>",
			got,
			10)
	}
}

func TestInt64Incr(t *testing.T) {
	var v Int64

	got := v.Incr()
	if got != int64(1) {
		t.Errorf("failed, got=<%d> expect=<%d>",
			got,
			10)
	}
}
