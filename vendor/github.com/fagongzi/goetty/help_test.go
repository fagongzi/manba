package goetty

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"unicode"
	"unsafe"
)

var (
	sizeOfInt     = int(unsafe.Sizeof(int(0)))
	sizeOfInt16   = int(unsafe.Sizeof(int16(0)))
	sizeOfInt32   = int(unsafe.Sizeof(int32(0)))
	sizeOfInt64   = int(unsafe.Sizeof(int64(0)))
	sizeOfUint    = int(unsafe.Sizeof(uint(0)))
	sizeOfUint16  = int(unsafe.Sizeof(uint16(0)))
	sizeOfUint32  = int(unsafe.Sizeof(uint32(0)))
	sizeOfUint64  = int(unsafe.Sizeof(uint64(0)))
	sizeOfFloat32 = int(unsafe.Sizeof(float32(0)))
	sizeOfFloat64 = int(unsafe.Sizeof(float64(0)))
)

type toString interface {
	String() string
}

type Equals interface {
	Equals(interface{}) bool
}

func Check(t *testing.T, condition bool, args ...interface{}) bool {
	return check(t, condition, "check", t.Fail, args...)
}

func Assert(t *testing.T, condition bool, args ...interface{}) {
	check(t, condition, "assert", t.FailNow, args...)
}

func check(t *testing.T, condition bool, tp string, f func(), args ...interface{}) bool {
	if condition {
		return true
	}
	var buf bytes.Buffer
	buf.WriteString(tp + " fail\n")
	for i := 0; i < len(args); i++ {
		buf.WriteString("        args[")
		buf.WriteString(strconv.Itoa(i))
		buf.WriteString("] = %#v")
		if i < len(args)-1 {
			buf.WriteByte('\n')
		}
	}
	log(2, fmt.Sprintf(buf.String(), args...))
	f()
	return false
}

func IsNil(t *testing.T, v interface{}) bool {
	return isNil(t, v, t.Fail)
}

func IsNilNow(t *testing.T, v interface{}) {
	isNil(t, v, t.FailNow)
}

func isNil(t *testing.T, v interface{}, f func()) bool {
	if v == nil {
		return true
	}
	switch vv := v.(type) {
	case toString:
		log(2, fmt.Sprintf(`not nil
        v = %#v`, vv.String()))
	case error:
		log(2, fmt.Sprintf(`not nil
        v = %#v`, vv.Error()))
	case []byte:
		log(2, fmt.Sprintf(`not nil
        v = %#v`, vv))
	default:
		log(2, fmt.Sprintf(`not nil
        v = %v`, vv))
	}
	f()
	return false
}

func NotNil(t *testing.T, v interface{}) bool {
	return notNil(t, v, t.Fail)
}

func NotNilNow(t *testing.T, v interface{}) {
	notNil(t, v, t.FailNow)
}

func notNil(t *testing.T, v interface{}, f func()) bool {
	if v != nil {
		return true
	}
	log(2, fmt.Sprintf(`is nil`))
	f()
	return false
}

func DeepEqual(t *testing.T, a, b interface{}) bool {
	return deepEqual(t, a, b, t.Fail)
}

func DeepEqualNow(t *testing.T, a, b interface{}) {
	deepEqual(t, a, b, t.FailNow)
}

func deepEqual(t *testing.T, a, b interface{}, f func()) bool {
	if reflect.DeepEqual(a, b) {
		return true
	}
	log(2, fmt.Sprintf(`not deep equal
        a = %#v
        b = %#v`, a, b))
	f()
	return false
}

func Equal(t *testing.T, a, b interface{}) bool {
	return equal(t, a, b, t.Fail)
}

func EqualNow(t *testing.T, a, b interface{}) {
	equal(t, a, b, t.FailNow)
}

func equal(t *testing.T, a, b interface{}, f func()) bool {
	if a == nil || b == nil {
		return a == b
	}
	var ok bool
	var printable bool
	switch va := a.(type) {
	case int:
		ok = va == b.(int)
	case int8:
		ok = va == int8Val(b)
	case int16:
		ok = va == int16Val(b)
	//case int32:
	//ok = va == b.(int32)
	case rune:
		vb := int32Val(b)
		ok = va == vb
		printable = unicode.IsPrint(va) && unicode.IsPrint(rune(vb))
	case int64:
		ok = va == int64Val(b)
	case uint:
		ok = va == uintVal(b)
	case uint8:
		ok = va == uint8Val(b)
	case uint16:
		ok = va == uint16Val(b)
	case uint32:
		ok = va == uint32Val(b)
	case uint64:
		ok = va == uint64Val(b)
	case float32:
		ok = va == float32Val(b)
	case float64:
		ok = va == float64Val(b)
	case string:
		ok = va == b.(string)
	case []byte:
		ok = bytes.Equal(va, b.([]byte))
	case []int:
		ok = unsafeEqual(a, b.([]int), sizeOfInt)
	case []int16:
		ok = unsafeEqual(a, b.([]int16), sizeOfInt16)
	case []int32:
		ok = unsafeEqual(a, b.([]int32), sizeOfInt32)
	case []int64:
		ok = unsafeEqual(a, b.([]int64), sizeOfInt64)
	case []uint:
		ok = unsafeEqual(a, b.([]uint), sizeOfUint)
	case []uint16:
		ok = unsafeEqual(a, b.([]uint16), sizeOfUint16)
	case []uint32:
		ok = unsafeEqual(a, b.([]uint32), sizeOfUint32)
	case []uint64:
		ok = unsafeEqual(a, b.([]uint64), sizeOfUint64)
	case []float32:
		ok = unsafeEqual(a, b.([]float32), sizeOfFloat32)
	case []float64:
		ok = unsafeEqual(a, b.([]float64), sizeOfFloat64)
	case Equals:
		ok = va.Equals(b)
	default:
		ok = reflect.DeepEqual(a, b)
	}
	if ok {
		return true
	}
	if printable {
		log(2, fmt.Sprintf(`not equal
        a = '%c' = %#v
        b = '%c' = %#v`, a, a, b, b))
	} else {
		log(2, fmt.Sprintf(`not equal
        a = %#v
        b = %#v`, a, b))
	}
	f()
	return false
}

func int8Val(b interface{}) int8 {
	switch v := b.(type) {
	case int:
		if int(math.MinInt8) <= v && v <= int(math.MaxInt8) {
			return int8(v)
		}
	case int8:
		return int8(v)
	}
	panic("can't convert to int8 value")
}

func uint8Val(b interface{}) uint8 {
	switch v := b.(type) {
	case int:
		if 0 <= v && v <= int(math.MaxUint8) {
			return uint8(v)
		}
	case uint8:
		return uint8(v)
	}
	panic("can't convert to uint8 value")
}

func int16Val(b interface{}) int16 {
	switch v := b.(type) {
	case int:
		if int(math.MinInt16) <= v && v <= int(math.MaxInt16) {
			return int16(v)
		}
	case int16:
		return int16(v)
	}
	panic("can't convert to int16 value")
}

func uint16Val(b interface{}) uint16 {
	switch v := b.(type) {
	case int:
		if 0 <= v && v <= int(math.MaxUint16) {
			return uint16(v)
		}
	case uint16:
		return uint16(v)
	}
	panic("can't convert to uint16 value")
}

func int32Val(b interface{}) int32 {
	switch v := b.(type) {
	case int:
		if int(math.MinInt32) <= v && v <= int(math.MaxInt32) {
			return int32(v)
		}
	case int32:
		return int32(v)
	}
	panic("can't convert to int32 value")
}

func uint32Val(b interface{}) uint32 {
	switch v := b.(type) {
	case int:
		if 0 <= v && v <= int(math.MaxUint32) {
			return uint32(v)
		}
	case uint32:
		return uint32(v)
	}
	panic("can't convert to uint32 value")
}

func int64Val(b interface{}) int64 {
	switch v := b.(type) {
	case int:
		return int64(v)
	case int64:
		return int64(v)
	}
	panic("can't convert to int64 value")
}

func uintVal(b interface{}) uint {
	switch v := b.(type) {
	case int:
		if 0 <= v {
			return uint(v)
		}
	case uint:
		return uint(v)
	}
	panic("can't convert to uint value")
}

func uint64Val(b interface{}) uint64 {
	switch v := b.(type) {
	case int:
		if 0 <= v {
			return uint64(v)
		}
	case uint64:
		return uint64(v)
	}
	panic("can't convert to uint64 value")
}

func float32Val(b interface{}) float32 {
	switch v := b.(type) {
	case int:
		return float32(v)
	case float32:
		return float32(v)
	}
	panic("can't convert to float32 value")
}

func float64Val(b interface{}) float64 {
	switch v := b.(type) {
	case int:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return float64(v)
	}
	panic("can't convert to float64 value")
}

func unsafeEqual(ia, ib interface{}, size int) bool {
	a := (*reflect.SliceHeader)((*(*[2]unsafe.Pointer)(unsafe.Pointer(&ia)))[1])
	b := (*reflect.SliceHeader)((*(*[2]unsafe.Pointer)(unsafe.Pointer(&ib)))[1])
	return bytes.Equal(
		*(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
			Data: a.Data,
			Cap:  a.Cap * size,
			Len:  a.Len * size,
		})),
		*(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
			Data: b.Data,
			Cap:  b.Cap * size,
			Len:  b.Len * size,
		})),
	)
}

func log(depth int, val string) {
	if _, file, line, ok := runtime.Caller(1 + depth); ok {
		// Truncate file name at last file name separator.
		if index := strings.LastIndex(file, "/"); index >= 0 {
			file = file[index+1:]
		} else if index = strings.LastIndex(file, "\\"); index >= 0 {
			file = file[index+1:]
		}
		fmt.Fprintf(os.Stderr, "    %s:%d: %s\n", file, line, val)
	}
}
