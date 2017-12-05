package hack

import (
	"bytes"
	"testing"
)

func TestSliceToString(t *testing.T) {
	value := []byte("hello world")
	target := SliceToString(value)

	if target != "hello world" {
		t.Errorf("failed, got=<%s> expect=<%s>", target, value)
	}

	value[0] = 'a'
	if target != "aello world" {
		t.Errorf("failed, got=<%s> expect=<%s>", target, value)
	}
}

func TestStringToSlice(t *testing.T) {
	info := "hello world"
	value := StringToSlice(info)

	if bytes.Compare(value, []byte("hello world")) != 0 {
		t.Errorf("failed, got=<%+v> expect=<%+v>", value, info)
	}
}
