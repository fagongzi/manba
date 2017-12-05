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
