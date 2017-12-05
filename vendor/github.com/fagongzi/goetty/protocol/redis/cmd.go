package redis

import (
	"bytes"
	"strings"

	"github.com/fagongzi/goetty"
)

// Command redis command
type Command [][]byte

// Cmd returns redis command
func (c Command) Cmd() []byte {
	return c[0]
}

// CmdString returns redis command use lower string
func (c Command) CmdString() string {
	return strings.ToLower(goetty.SliceToString(c[0]))
}

// Args returns redis command args
func (c Command) Args() [][]byte {
	return c[1:]
}

// ToString returns a redis command as string
func (c Command) ToString() string {
	buf := new(bytes.Buffer)
	for _, arg := range c {
		buf.Write(arg)
		buf.WriteString(" ")
	}

	return strings.TrimSpace(buf.String())
}

// StatusResp status resp
type StatusResp []byte

// ErrResp error resp
type ErrResp []byte

// IntegerResp integer resp
type IntegerResp []byte

// BulkResp bulk resp
type BulkResp []byte

// NullBulkResp null bulk resp
type NullBulkResp int

// NullArrayResp null array resp
type NullArrayResp int
