package example

import (
	"github.com/fagongzi/goetty"
)

// StringDecoder string decode
type StringDecoder struct {
}

// Decode decode
func (decoder StringDecoder) Decode(in *goetty.ByteBuf) (bool, interface{}, error) {
	_, data, err := in.ReadMarkedBytes()

	if err != nil {
		return true, "", err
	}

	return true, string(data), nil
}

// StringEncoder string encode
type StringEncoder struct {
}

// Encode encode
func (e StringEncoder) Encode(data interface{}, out *goetty.ByteBuf) error {
	msg, _ := data.(string)
	bytes := []byte(msg)
	out.WriteInt(len(bytes))
	out.Write(bytes)
	return nil
}
