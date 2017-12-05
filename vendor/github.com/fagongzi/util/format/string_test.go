package format

import (
	"testing"
)

func TestParseStrInt64(t *testing.T) {
	value := []byte("10")
	got, err := ParseStrInt64(value)
	if err != nil {
		t.Errorf("failed, errors:%+v", err)
	}

	if got != 10 {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", got, 10)
	}
}

func TestParseStrFloat64(t *testing.T) {
	value := []byte("10.10")
	got, err := ParseStrFloat64(value)
	if err != nil {
		t.Errorf("failed, errors:%+v", err)
	}

	if got != 10.10 {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", got, 10.10)
	}
}
