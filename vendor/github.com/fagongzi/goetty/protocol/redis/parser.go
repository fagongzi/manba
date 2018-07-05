package redis

import (
	"fmt"
	"strconv"

	"github.com/fagongzi/goetty"
)

const (
	// CR \r
	CR = '\r'
	// LF \n
	LF = '\n'
	// ARGBegin '$'
	ARGBegin = '$'
	// CMDBegin '*'
	CMDBegin = '*'

	defaultBufferSize = 64
)

var (
	// Delims delims
	Delims = []byte("\r\n")
	// NullBulk empty bulk string
	NullBulk = []byte("-1")
	// NullArray empty array
	NullArray = []byte("-1")
)

// ReadCommand returns redis command from buffer
func ReadCommand(in *goetty.ByteBuf) (bool, Command, error) {
	for {
		// remember the begin read index,
		// if we found has no enough data, we will resume this read index,
		// and waiting for next.
		backupReaderIndex := in.GetReaderIndex()

		c, err := in.ReadByte()
		if err != nil {
			return false, nil, err
		}

		// 1. Read ( *<number of arguments> CR LF )
		if c != CMDBegin {
			return false, nil, fmt.Errorf("illegal cmd char %d", c)
		}

		// 2. Read number of arguments
		count, argsCount, err := readStringInt(in)
		if count == 0 && err == nil {
			in.SetReaderIndex(backupReaderIndex)
			return false, nil, nil
		} else if err != nil {
			return false, nil, err
		}

		data := make([][]byte, argsCount)

		// 3. Read args
		for i := 0; i < argsCount; i++ {
			if 0 == in.Readable() {
				in.SetReaderIndex(backupReaderIndex)
				return false, nil, nil
			}

			// 3.1 Read ( $<number of bytes of argument 1> CR LF )
			c, err := in.ReadByte()
			if err != nil {
				return false, nil, err
			}

			// 3.2 Read ( *<number of arguments> CR LF )
			if c != ARGBegin {
				return false, nil, fmt.Errorf("illegal arg char %d", c)
			}

			count, argBytesCount, err := readStringInt(in)
			if count == 0 && err == nil {
				in.SetReaderIndex(backupReaderIndex)
				return false, nil, nil
			} else if err != nil {
				return false, nil, err
			} else if count < 2 {
				return false, nil, fmt.Errorf("illegal arg count: %d", count)
			}

			// 3.3  Read ( <argument data> CR LF )
			if argBytesCount+2 > in.Readable() {
				in.SetReaderIndex(backupReaderIndex)
				return false, nil, nil
			}
			count, value, err := in.ReadBytes(argBytesCount + 2)
			if count == 0 && err == nil {
				in.SetReaderIndex(backupReaderIndex)
				return false, nil, nil
			} else if err != nil {
				return false, nil, err
			}

			data[i] = value[:count-2]
		}

		return true, Command(data), nil
	}
}

func readCommandReply(in *goetty.ByteBuf) (bool, interface{}, error) {
	for {
		// remember the begin read index,
		// if we found has no enough data, we will resume this read index,
		// and waiting for next.
		backupReaderIndex := in.GetReaderIndex()
		cnt, line, err := readLine(in)
		if cnt == 0 && err == nil {
			in.SetReaderIndex(backupReaderIndex)
			return false, nil, nil
		} else if err != nil {
			return false, nil, err
		}

		line = line[:cnt-2]

		switch line[0] {
		case '+':
			return true, StatusResp(line[1:]), nil
		case '-':
			return true, ErrResp(line[1:]), nil
		case ':':
			return true, IntegerResp(line[1:]), nil
		case '$':
			size, err := parseInteger(line[1:])
			if err != nil {
				return false, nil, err
			}

			if size < 0 {
				return true, NullBulkResp(0), nil
			}

			// have not enough data, wait for next
			if size+2 > in.Readable() {
				in.SetReaderIndex(backupReaderIndex)
				return false, nil, nil
			}

			c, data, err := in.ReadBytes(size + 2)
			if err != nil {
				return false, nil, err
			}

			if c == 0 {
				in.SetReaderIndex(backupReaderIndex)
				return false, nil, nil
			}

			return true, BulkResp(data[:c-2]), nil
		case '*':
			size, err := parseInteger(line[1:])
			if err != nil {
				return false, nil, err
			}

			if size < 0 {
				return true, NullArrayResp(0), nil
			}

			r := make([]interface{}, size)
			for i := range r {
				complete, value, err := readCommandReply(in)
				if err != nil {
					return false, nil, err
				}

				if !complete {
					in.SetReaderIndex(backupReaderIndex)
					return false, nil, nil
				}

				r[i] = value
			}

			return true, r, nil
		default:
			return false, nil, fmt.Errorf("illegal cmd reply char %c", line[0])
		}
	}
}

func readStringInt(in *goetty.ByteBuf) (int, int, error) {
	count, line, err := readLine(in)
	if count == 0 && err == nil {
		return 0, 0, nil
	} else if err != nil {
		return 0, 0, err
	}

	// count-2:xclude 'CR CF'
	value, err := parseInteger(line[:count-2])
	if err != nil {
		return 0, 0, err
	}

	return len(line), value, nil
}

func readLine(in *goetty.ByteBuf) (int, []byte, error) {
	offset := 0
	size := in.Readable()

	for offset < size {
		ch, _ := in.PeekByte(offset)
		if ch == LF {
			ch, _ := in.PeekByte(offset - 1)
			if ch == CR {
				return in.ReadBytes(offset + 1)
			}

			return 0, nil, fmt.Errorf("illegal line end: %d", ch)
		}
		offset++
	}

	return 0, nil, nil
}

func parseInteger(data []byte) (int, error) {
	if data[0] == '-' && len(data) == 2 && data[1] == '1' {
		// handle $-1 and $-1 null replies.
		return -1, nil
	}

	return strconv.Atoi(goetty.SliceToString(data))
}
