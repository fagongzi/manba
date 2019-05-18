package goetty

import (
	"fmt"
)

// StringCodec a simple string encoder and decoder
type StringCodec struct{}

// Decode decode
func (codec *StringCodec) Decode(in *ByteBuf) (bool, interface{}, error) {
	value := string(in.GetMarkedRemindData())
	in.MarkedBytesReaded()
	return true, value, nil
}

// Encode encode
func (codec *StringCodec) Encode(data interface{}, out *ByteBuf) error {
	if msg, ok := data.(string); ok {
		return out.WriteString(msg)
	}

	return fmt.Errorf("not string: %+v", data)
}
