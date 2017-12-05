package format

import (
	"encoding/binary"
	"fmt"
)

// Uint16ToBytes uint16 -> bytes
func Uint16ToBytes(v uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, v)
	return b
}

// Uint32ToBytes uint32 -> bytes
func Uint32ToBytes(v uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, v)
	return b
}

// Uint64ToBytes uint64 -> bytes
func Uint64ToBytes(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

// BytesToUint16 bytes -> uint16
func BytesToUint16(b []byte) (uint16, error) {
	if len(b) != 2 {
		return 0, fmt.Errorf("invalid data, must 2 bytes, but %d", len(b))
	}

	return binary.BigEndian.Uint16(b), nil
}

// BytesToUint32 bytes -> uint32
func BytesToUint32(b []byte) (uint32, error) {
	if len(b) != 4 {
		return 0, fmt.Errorf("invalid data, must 4 bytes, but %d", len(b))
	}

	return binary.BigEndian.Uint32(b), nil
}

// BytesToUint64 bytes -> uint64
func BytesToUint64(b []byte) (uint64, error) {
	if len(b) != 8 {
		return 0, fmt.Errorf("invalid data, must 8 bytes, but %d", len(b))
	}

	return binary.BigEndian.Uint64(b), nil
}
