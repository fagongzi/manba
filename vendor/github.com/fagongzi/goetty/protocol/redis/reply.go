package redis

import (
	"fmt"
	"strconv"

	"github.com/fagongzi/goetty"
)

// WriteError write error resp
func WriteError(err []byte, buf *goetty.ByteBuf) {
	buf.WriteByte('-')
	if err != nil {
		buf.WriteByte(' ')
		buf.Write(err)
	}
	buf.Write(Delims)
}

// WriteStatus write status resp
func WriteStatus(status []byte, buf *goetty.ByteBuf) {
	buf.WriteByte('+')
	buf.Write(status)
	buf.Write(Delims)
}

// WriteInteger write integer resp
func WriteInteger(n int64, buf *goetty.ByteBuf) {
	buf.WriteByte(':')
	buf.Write(goetty.FormatInt64ToBytes(n))
	buf.Write(Delims)
}

// WriteBulk write bulk resp
func WriteBulk(b []byte, buf *goetty.ByteBuf) {
	buf.WriteByte('$')
	if len(b) == 0 {
		buf.Write(NullBulk)
	} else {
		buf.Write(goetty.StringToSlice(strconv.Itoa(len(b))))
		buf.Write(Delims)
		buf.Write(b)
	}

	buf.Write(Delims)
}

// WriteArray write array resp
func WriteArray(lst []interface{}, buf *goetty.ByteBuf) {
	buf.WriteByte('*')
	if len(lst) == 0 {
		buf.Write(NullArray)
		buf.Write(Delims)
	} else {
		buf.Write(goetty.StringToSlice(strconv.Itoa(len(lst))))
		buf.Write(Delims)

		for i := 0; i < len(lst); i++ {
			switch v := lst[i].(type) {
			case []interface{}:
				WriteArray(v, buf)
			case [][]byte:
				WriteSliceArray(v, buf)
			case []byte:
				WriteBulk(v, buf)
			case nil:
				WriteBulk(nil, buf)
			case int64:
				WriteInteger(v, buf)
			case string:
				WriteStatus(goetty.StringToSlice(v), buf)
			case error:
				WriteError(goetty.StringToSlice(v.Error()), buf)
			default:
				panic(fmt.Sprintf("invalid array type %T %v", lst[i], v))
			}
		}
	}
}

// WriteSliceArray write slice array resp
func WriteSliceArray(lst [][]byte, buf *goetty.ByteBuf) {
	buf.WriteByte('*')
	if len(lst) == 0 {
		buf.Write(NullArray)
		buf.Write(Delims)
	} else {
		buf.Write(goetty.StringToSlice(strconv.Itoa(len(lst))))
		buf.Write(Delims)

		for i := 0; i < len(lst); i++ {
			WriteBulk(lst[i], buf)
		}
	}
}

// WriteFVPairArray write field-value pair array resp
func WriteFVPairArray(fields, values [][]byte, buf *goetty.ByteBuf) {
	buf.WriteByte('*')
	if len(fields) == 0 || len(values) == 0 {
		buf.Write(NullArray)
		buf.Write(Delims)
	} else {
		buf.Write(goetty.StringToSlice(strconv.Itoa(len(fields) * 2)))
		buf.Write(Delims)

		for i := 0; i < len(values); i++ {
			WriteBulk(fields[i], buf)
			WriteBulk(values[i], buf)
		}
	}
}

// WriteScorePairArray write score-member pair array resp
func WriteScorePairArray(members [][]byte, scores []float64, withScores bool, buf *goetty.ByteBuf) {
	buf.WriteByte('*')
	if len(members) == 0 || len(scores) == 0 {
		buf.Write(NullArray)
		buf.Write(Delims)
	} else {
		if withScores {
			buf.Write(goetty.StringToSlice(strconv.Itoa(len(members) * 2)))
			buf.Write(Delims)
		} else {
			buf.Write(goetty.StringToSlice(strconv.Itoa(len(members))))
			buf.Write(Delims)

		}

		for i := 0; i < len(members); i++ {
			WriteBulk(members[i], buf)

			if withScores {
				WriteBulk(goetty.FormatFloat64ToBytes(scores[i]), buf)
			}
		}
	}
}
