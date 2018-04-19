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

// ParseStrInt str -> int
func ParseStrInt(data string) (int, error) {
	v, err := strconv.ParseInt(data, 10, 32)
	if err != nil {
		return 0, err
	}

	return int(v), nil
}

// MustParseStrInt str -> int
func MustParseStrInt(data string) int {
	value, err := ParseStrInt(data)
	if err != nil {
		buf := make([]byte, 4096)
		runtime.Stack(buf, true)
		log.Fatalf("parse to int failed, data=<%s> errors:\n %+v \n %s",
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

// ParseStrIntSlice parse []string -> []int
func ParseStrIntSlice(data []string) ([]int, error) {
	var target []int

	for _, str := range data {
		id, err := ParseStrInt(str)
		if err != nil {
			return nil, err
		}

		target = append(target, id)
	}

	return target, nil
}

// ParseStrInt64Slice parse []string -> []int64
func ParseStrInt64Slice(data []string) ([]int64, error) {
	var target []int64

	for _, str := range data {
		id, err := ParseStrInt64(str)
		if err != nil {
			return nil, err
		}

		target = append(target, id)
	}

	return target, nil
}

// ParseStrUInt64Slice parse []string -> []uint64
func ParseStrUInt64Slice(data []string) ([]uint64, error) {
	var target []uint64

	for _, str := range data {
		id, err := ParseStrUInt64(str)
		if err != nil {
			return nil, err
		}

		target = append(target, id)
	}

	return target, nil
}
