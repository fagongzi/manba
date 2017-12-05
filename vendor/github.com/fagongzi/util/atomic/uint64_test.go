package atomic

import (
	"testing"
)

func TestUint64SetAndGet(t *testing.T) {
	var v Uint64
	v.Set(10)

	got := v.Get()

	if got != uint64(10) {
		t.Errorf("failed, got=<%d> expect=<%d>",
			got,
			10)
	}
}

func TestUint64Incr(t *testing.T) {
	var v Uint64

	got := v.Incr()
	if got != uint64(1) {
		t.Errorf("failed, got=<%d> expect=<%d>",
			got,
			10)
	}
}
