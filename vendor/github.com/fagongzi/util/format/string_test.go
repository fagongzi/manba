package format

import (
	"testing"
)

func TestParseStrInt64(t *testing.T) {
	value := "10"
	got, err := ParseStrInt64(value)
	if err != nil {
		t.Errorf("failed, errors:%+v", err)
	}

	if got != 10 {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", got, 10)
	}
}

func TestParseStrFloat64(t *testing.T) {
	value := "10.10"
	got, err := ParseStrFloat64(value)
	if err != nil {
		t.Errorf("failed, errors:%+v", err)
	}

	if got != 10.10 {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", got, 10.10)
	}
}

func TestParseStrInt(t *testing.T) {
	value := "10"
	got, err := ParseStrInt(value)
	if err != nil {
		t.Errorf("failed, errors:%+v", err)
	}

	if got != 10 {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", got, 10)
	}
}

func TestParseStrIntSlice(t *testing.T) {
	value := []string{"10", "11", "12"}
	got, err := ParseStrIntSlice(value)
	if err != nil {
		t.Errorf("failed, errors:%+v", err)
	}

	if len(got) != len(value) {
		t.Errorf("failed, got=<%d>, expect=<%d>", len(got), len(value))
		return
	}

	if got[0] != 10 {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", got[0], 10)
		return
	}

	if got[1] != 11 {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", got[1], 11)
		return
	}

	if got[2] != 12 {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", got[1], 11)
		return
	}
}

func TestParseStrInt64Slice(t *testing.T) {
	value := []string{"10", "11", "12"}
	got, err := ParseStrInt64Slice(value)
	if err != nil {
		t.Errorf("failed, errors:%+v", err)
	}

	if len(got) != len(value) {
		t.Errorf("failed, got=<%d>, expect=<%d>", len(got), len(value))
		return
	}

	if got[0] != 10 {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", got[0], 10)
		return
	}

	if got[1] != 11 {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", got[1], 11)
		return
	}

	if got[2] != 12 {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", got[1], 11)
		return
	}
}

func TestParseStrUInt64Slice(t *testing.T) {
	value := []string{"10", "11", "12"}
	got, err := ParseStrUInt64Slice(value)
	if err != nil {
		t.Errorf("failed, errors:%+v", err)
	}

	if len(got) != len(value) {
		t.Errorf("failed, got=<%d>, expect=<%d>", len(got), len(value))
		return
	}

	if got[0] != 10 {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", got[0], 10)
		return
	}

	if got[1] != 11 {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", got[1], 11)
		return
	}

	if got[2] != 12 {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", got[1], 11)
		return
	}
}
