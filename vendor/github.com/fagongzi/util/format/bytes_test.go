package format

import (
	"bytes"
	"testing"
)

func TestUint16ToBytes(t *testing.T) {
	var value uint16
	value = 1

	data := Uint16ToBytes(value)

	expect := []byte{0, 1}
	if bytes.Compare(data, expect) != 0 {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", data, expect)
	}
}

func TestUint32ToBytes(t *testing.T) {
	var value uint32
	value = 256

	data := Uint32ToBytes(value)

	expect := []byte{0, 0, 1, 0}
	if bytes.Compare(data, expect) != 0 {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", data, expect)
	}
}

func TestUint64ToBytes(t *testing.T) {
	var value uint64
	value = 256

	data := Uint64ToBytes(value)

	expect := []byte{0, 0, 0, 0, 0, 0, 1, 0}
	if bytes.Compare(data, expect) != 0 {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", data, expect)
	}
}

func TestBytesToUint64(t *testing.T) {
	expect := 256
	target := []byte{0, 0, 0, 0, 0, 0, 1, 0}
	value, err := BytesToUint64(target)
	if err != nil {
		t.Errorf("BytesToUint64 error: %+v", err)
	}

	if value != uint64(expect) {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", value, expect)
	}
}

func TestBytesToUint32(t *testing.T) {
	expect := 256
	target := []byte{0, 0, 1, 0}
	value, err := BytesToUint32(target)
	if err != nil {
		t.Errorf("BytesToUint32 error: %+v", err)
	}

	if value != uint32(expect) {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", value, expect)
	}
}

func TestBytesToUint16(t *testing.T) {
	expect := 256
	target := []byte{1, 0}
	value, err := BytesToUint16(target)
	if err != nil {
		t.Errorf("BytesToUint16 error: %+v", err)
	}

	if value != uint16(expect) {
		t.Errorf("failed, got=<%+v>, expect=<%+v>", value, expect)
	}
}
