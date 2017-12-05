package format

import (
	"strconv"

	"github.com/fagongzi/util/hack"
)

// ParseStrInt64 str -> int64
func ParseStrInt64(v []byte) (int64, error) {
	return strconv.ParseInt(hack.SliceToString(v), 10, 64)
}

// ParseStrFloat64 str -> float64
func ParseStrFloat64(v []byte) (float64, error) {
	return strconv.ParseFloat(hack.SliceToString(v), 64)
}
