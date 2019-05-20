package fastrand

import (
	"testing"
)

func TestUint32(t *testing.T) {
	m := make(map[uint32]struct{})
	for i := 0; i < 1e6; i++ {
		n := Uint32()
		if _, ok := m[n]; ok {
			t.Fatalf("number %v already exists", n)
		}
		m[n] = struct{}{}
	}
}

func TestRNGUint32(t *testing.T) {
	var r RNG
	m := make(map[uint32]struct{})
	for i := 0; i < 1e6; i++ {
		n := r.Uint32()
		if _, ok := m[n]; ok {
			t.Fatalf("number %v already exists", n)
		}
		m[n] = struct{}{}
	}
}

func TestUint32n(t *testing.T) {
	m := make(map[uint32]int)
	for i := 0; i < 1e6; i++ {
		n := Uint32n(1e2)
		if n >= 1e2 {
			t.Fatalf("n > 1000: %v", n)
		}
		m[n]++
	}

	// check distribution
	avg := 1e6 / 1e2
	for k, v := range m {
		p := (float64(v) - float64(avg)) / float64(avg)
		if p < 0 {
			p = -p
		}
		if p > 0.05 {
			t.Fatalf("skew more than 5%% for k=%v: %v", k, p*100)
		}
	}
}

func TestRNGUint32n(t *testing.T) {
	var r RNG
	m := make(map[uint32]int)
	for i := 0; i < 1e6; i++ {
		n := r.Uint32n(1e2)
		if n >= 1e2 {
			t.Fatalf("n > 1000: %v", n)
		}
		m[n]++
	}

	// check distribution
	avg := 1e6 / 1e2
	for k, v := range m {
		p := (float64(v) - float64(avg)) / float64(avg)
		if p < 0 {
			p = -p
		}
		if p > 0.05 {
			t.Fatalf("skew more than 5%% for k=%v: %v", k, p*100)
		}
	}
}
