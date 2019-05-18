package goetty

import (
	"encoding/binary"
	"fmt"
	"strconv"
)

// BytesToUint64 bytes -> uint64
func BytesToUint64(b []byte) (uint64, error) {
	if len(b) != 8 {
		return 0, fmt.Errorf("invalid data, must 8 bytes, but %d", len(b))
	}

	return binary.BigEndian.Uint64(b), nil
}

// Uint64ToBytes uint64 -> bytes
func Uint64ToBytes(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

// StrInt64 str -> int64
func StrInt64(v []byte) (int64, error) {
	return strconv.ParseInt(SliceToString(v), 10, 64)
}

// StrFloat64 str -> float64
func StrFloat64(v []byte) (float64, error) {
	return strconv.ParseFloat(SliceToString(v), 64)
}

// FormatInt64ToBytes int64 -> string
func FormatInt64ToBytes(v int64) []byte {
	return strconv.AppendInt(nil, v, 10)
}

// FormatFloat64ToBytes float64 -> string
func FormatFloat64ToBytes(v float64) []byte {
	return strconv.AppendFloat(nil, v, 'f', -1, 64)
}
