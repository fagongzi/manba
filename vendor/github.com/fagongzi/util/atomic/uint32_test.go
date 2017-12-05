package atomic

import (
	"testing"
)

func TestUint32SetAndGet(t *testing.T) {
	var v Uint32
	v.Set(10)

	got := v.Get()

	if got != uint32(10) {
		t.Errorf("failed, got=<%d> expect=<%d>",
			got,
			10)
	}
}

func TestUint32Incr(t *testing.T) {
	var v Uint32

	got := v.Incr()
	if got != uint32(1) {
		t.Errorf("failed, got=<%d> expect=<%d>",
			got,
			10)
	}
}
