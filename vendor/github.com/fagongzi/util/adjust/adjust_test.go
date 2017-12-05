package adjust

import (
	"testing"
)

func TestInt(t *testing.T) {
	var value, adjust int
	value = 0
	adjust = 1

	got := Int(value, adjust)
	if got != adjust {
		t.Errorf("failed, got=<%d> expect=<%d>", got, adjust)
	}

	value = 1
	adjust = 2
	got = Int(value, adjust)
	if got != value {
		t.Errorf("failed, got=<%d> expect=<%d>", got, value)
	}
}

func TestInt8(t *testing.T) {
	var value, adjust int8
	value = 0
	adjust = 1

	got := Int8(value, adjust)
	if got != adjust {
		t.Errorf("failed, got=<%d> expect=<%d>", got, adjust)
	}

	value = 1
	adjust = 2
	got = Int8(value, adjust)
	if got != value {
		t.Errorf("failed, got=<%d> expect=<%d>", got, value)
	}
}

func TestInt16(t *testing.T) {
	var value, adjust int16
	value = 0
	adjust = 1

	got := Int16(value, adjust)
	if got != adjust {
		t.Errorf("failed, got=<%d> expect=<%d>", got, adjust)
	}

	value = 1
	adjust = 2
	got = Int16(value, adjust)
	if got != value {
		t.Errorf("failed, got=<%d> expect=<%d>", got, value)
	}
}

func TestInt32(t *testing.T) {
	var value, adjust int32
	value = 0
	adjust = 1

	got := Int32(value, adjust)
	if got != adjust {
		t.Errorf("failed, got=<%d> expect=<%d>", got, adjust)
	}

	value = 1
	adjust = 2
	got = Int32(value, adjust)
	if got != value {
		t.Errorf("failed, got=<%d> expect=<%d>", got, value)
	}
}

func TestInt64(t *testing.T) {
	var value, adjust int64
	value = 0
	adjust = 1

	got := Int64(value, adjust)
	if got != adjust {
		t.Errorf("failed, got=<%d> expect=<%d>", got, adjust)
	}

	value = 1
	adjust = 2
	got = Int64(value, adjust)
	if got != value {
		t.Errorf("failed, got=<%d> expect=<%d>", got, value)
	}
}

func TestUInt(t *testing.T) {
	var value, adjust uint
	value = 0
	adjust = 1

	got := UInt(value, adjust)
	if got != adjust {
		t.Errorf("failed, got=<%d> expect=<%d>", got, adjust)
	}

	value = 1
	adjust = 2
	got = UInt(value, adjust)
	if got != value {
		t.Errorf("failed, got=<%d> expect=<%d>", got, value)
	}
}

func TestUInt8(t *testing.T) {
	var value, adjust uint8
	value = 0
	adjust = 1

	got := UInt8(value, adjust)
	if got != adjust {
		t.Errorf("failed, got=<%d> expect=<%d>", got, adjust)
	}

	value = 1
	adjust = 2
	got = UInt8(value, adjust)
	if got != value {
		t.Errorf("failed, got=<%d> expect=<%d>", got, value)
	}
}

func TestUInt16(t *testing.T) {
	var value, adjust uint16
	value = 0
	adjust = 1

	got := UInt16(value, adjust)
	if got != adjust {
		t.Errorf("failed, got=<%d> expect=<%d>", got, adjust)
	}

	value = 1
	adjust = 2
	got = UInt16(value, adjust)
	if got != value {
		t.Errorf("failed, got=<%d> expect=<%d>", got, value)
	}
}

func TestUInt32(t *testing.T) {
	var value, adjust uint32
	value = 0
	adjust = 1

	got := UInt32(value, adjust)
	if got != adjust {
		t.Errorf("failed, got=<%d> expect=<%d>", got, adjust)
	}

	value = 1
	adjust = 2
	got = UInt32(value, adjust)
	if got != value {
		t.Errorf("failed, got=<%d> expect=<%d>", got, value)
	}
}

func TestUInt64(t *testing.T) {
	var value, adjust uint64
	value = 0
	adjust = 1

	got := UInt64(value, adjust)
	if got != adjust {
		t.Errorf("failed, got=<%d> expect=<%d>", got, adjust)
	}

	value = 1
	adjust = 2
	got = UInt64(value, adjust)
	if got != value {
		t.Errorf("failed, got=<%d> expect=<%d>", got, value)
	}
}
