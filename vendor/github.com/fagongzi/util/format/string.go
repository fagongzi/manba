package format

import (
	"log"
	"runtime"
	"strconv"
)

// ParseStrUInt64 str -> uint64
func ParseStrUInt64(data string) (uint64, error) {
	ret, err := strconv.ParseInt(data, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint64(ret), nil
}

// MustParseStrUInt64 str -> uint64
func MustParseStrUInt64(data string) uint64 {
	value, err := ParseStrUInt64(data)
	if err != nil {
		buf := make([]byte, 4096)
		runtime.Stack(buf, true)
		log.Fatalf("parse to uint64 failed, data=<%s> errors:\n %+v \n %s",
			data,
			err,
			buf)
	}

	return value
}

// ParseStrInt64 str -> int64
func ParseStrInt64(data string) (int64, error) {
	return strconv.ParseInt(data, 10, 64)
}

// MustParseStrInt64 str -> int64
func MustParseStrInt64(data string) int64 {
	value, err := ParseStrInt64(data)
	if err != nil {
		buf := make([]byte, 4096)
		runtime.Stack(buf, true)
		log.Fatalf("parse to int64 failed, data=<%s> errors:\n %+v \n %s",
			data,
			err,
			buf)
	}

	return value
}

// ParseStrFloat64 str -> float64
func ParseStrFloat64(data string) (float64, error) {
	return strconv.ParseFloat(data, 64)
}

// MustParseStrFloat64 str -> float64
func MustParseStrFloat64(data string) float64 {
	value, err := ParseStrFloat64(data)
	if err != nil {
		buf := make([]byte, 4096)
		runtime.Stack(buf, true)
		log.Fatalf("parse to float64 failed, data=<%s> errors:\n %+v \n %s",
			data,
			err,
			buf)
	}

	return value
}
