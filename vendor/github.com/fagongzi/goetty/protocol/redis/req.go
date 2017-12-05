package redis

import (
	"strconv"

	"github.com/fagongzi/goetty"
)

const (
	lenScratch = "lenScratch"
	numScratch = "numScratch"
)

// InitRedisConn init redis conn
func InitRedisConn(conn goetty.IOSession) {
	conn.SetAttr(lenScratch, make([]byte, 32, 32))
	conn.SetAttr(numScratch, make([]byte, 40, 40))
}

// WriteCommand write redis command
func WriteCommand(conn goetty.IOSession, cmd string, args ...interface{}) error {
	lenV := conn.GetAttr(lenScratch).([]byte)
	numV := conn.GetAttr(lenScratch).([]byte)

	return doWriteCommand(cmd, lenV, numV, conn.OutBuf(), args...)
}

func doWriteCommand(cmd string, lenScratch, numScratch []byte, buf *goetty.ByteBuf, args ...interface{}) (err error) {
	writeLen('*', 1+len(args), lenScratch, buf)
	err = writeString(cmd, lenScratch, buf)

	for _, arg := range args {
		if err != nil {
			break
		}
		switch arg := arg.(type) {
		case string:
			err = writeString(arg, lenScratch, buf)
		case []byte:
			err = writeBytes(arg, lenScratch, buf)
		case int:
			err = writeInt64(int64(arg), lenScratch, numScratch, buf)
		case int64:
			err = writeInt64(arg, lenScratch, numScratch, buf)
		case float64:
			err = writeFloat64(arg, lenScratch, numScratch, buf)
		case bool:
			if arg {
				err = writeString("1", lenScratch, buf)
			} else {
				err = writeString("0", lenScratch, buf)
			}
		case nil:
			err = writeString("", lenScratch, buf)
		}
	}
	return err
}

func writeLen(prefix byte, n int, lenScratch []byte, buf *goetty.ByteBuf) error {
	lenScratch[len(lenScratch)-1] = '\n'
	lenScratch[len(lenScratch)-2] = '\r'
	i := len(lenScratch) - 3
	for {
		lenScratch[i] = byte('0' + n%10)
		i--
		n = n / 10
		if n == 0 {
			break
		}
	}
	lenScratch[i] = prefix

	_, err := buf.Write(lenScratch[i:])
	return err
}

func writeString(s string, lenScratch []byte, buf *goetty.ByteBuf) error {
	writeLen('$', len(s), lenScratch, buf)

	buf.Write([]byte(s))
	_, err := buf.Write(Delims)
	return err
}

func writeBytes(p []byte, lenScratch []byte, buf *goetty.ByteBuf) error {
	writeLen('$', len(p), lenScratch, buf)
	buf.Write(p)
	_, err := buf.Write(Delims)
	return err
}

func writeInt64(n int64, lenScratch, numScratch []byte, buf *goetty.ByteBuf) error {
	return writeBytes(strconv.AppendInt(numScratch[:0], n, 10), lenScratch, buf)
}

func writeFloat64(n float64, lenScratch, numScratch []byte, buf *goetty.ByteBuf) error {
	return writeBytes(strconv.AppendFloat(numScratch[:0], n, 'g', -1, 64), lenScratch, buf)
}
