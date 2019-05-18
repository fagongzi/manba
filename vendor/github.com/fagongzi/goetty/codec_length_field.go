package goetty

const (
	// FieldLength field length bytes
	FieldLength = 4
)

// IntLengthFieldBasedDecoder decoder based on length filed + data
type IntLengthFieldBasedDecoder struct {
	base                Decoder
	lengthFieldOffset   int
	lengthAdjustment    int
	initialBytesToStrip int
}

// NewIntLengthFieldBasedDecoder create a IntLengthFieldBasedDecoder
func NewIntLengthFieldBasedDecoder(base Decoder) Decoder {
	return NewIntLengthFieldBasedDecoderSize(base, 0, 0, 0)
}

// NewIntLengthFieldBasedDecoderSize  create a IntLengthFieldBasedDecoder
// initialBytesToStrip + lengthFieldOffset + 4(length)
// lengthAdjustment, some case as below:
// 1. 0 :                                             base decoder received: body
// 2. -4:                                             base decoder received: 4(length) + body
// 3. -(4 + lengthFieldOffset):                       base decoder received: lengthFieldOffset + 4(length) + body
// 4. -(4 + lengthFieldOffset + initialBytesToStrip): base decoder received: initialBytesToStrip + lengthFieldOffset + 4(length)
func NewIntLengthFieldBasedDecoderSize(base Decoder, lengthFieldOffset, lengthAdjustment, initialBytesToStrip int) Decoder {
	return &IntLengthFieldBasedDecoder{
		base:                base,
		lengthFieldOffset:   lengthFieldOffset,
		lengthAdjustment:    lengthAdjustment,
		initialBytesToStrip: initialBytesToStrip,
	}
}

// Decode decode
func (decoder IntLengthFieldBasedDecoder) Decode(in *ByteBuf) (bool, interface{}, error) {
	readable := in.Readable()

	minFrameLength := decoder.initialBytesToStrip + decoder.lengthFieldOffset + FieldLength
	if readable < minFrameLength {
		return false, nil, nil
	}

	length, err := in.PeekInt(decoder.initialBytesToStrip + decoder.lengthFieldOffset)
	if err != nil {
		return true, nil, err
	}

	skip := minFrameLength + decoder.lengthAdjustment
	minFrameLength += length
	if readable < minFrameLength {
		return false, nil, nil
	}

	in.Skip(skip)
	in.MarkN(length)
	return decoder.base.Decode(in)
}

// IntLengthFieldBasedEncoder encoder based on length filed + data
type IntLengthFieldBasedEncoder struct {
	base    Encoder
	prepare func(data interface{}, out *ByteBuf) error
}

// NewIntLengthFieldBasedEncoder returns a encoder with base
func NewIntLengthFieldBasedEncoder(base Encoder) Encoder {
	return NewIntLengthFieldBasedEncoderWithPrepare(base, nil)
}

// NewIntLengthFieldBasedEncoderWithPrepare returns a encoder with base and prepare fun
func NewIntLengthFieldBasedEncoderWithPrepare(base Encoder, prepare func(data interface{}, out *ByteBuf) error) Encoder {
	return &IntLengthFieldBasedEncoder{
		base:    base,
		prepare: prepare,
	}
}

// Encode encode
func (encoder *IntLengthFieldBasedEncoder) Encode(data interface{}, out *ByteBuf) error {
	buf := NewByteBuf(32)

	if encoder.prepare != nil {
		err := encoder.prepare(data, out)
		if err != nil {
			return err
		}
	}

	err := encoder.base.Encode(data, buf)
	if err != nil {
		return err
	}

	out.WriteInt(buf.Readable())
	out.WriteByteBuf(buf)
	return nil
}
