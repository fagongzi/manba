package goetty

const (
	INT_LENGTH_FIELD_LENGTH = 4
)

type IntLengthFieldBasedDecoder struct {
	base                Decoder
	lengthFieldOffset   int
	lengthAdjustment    int
	initialBytesToStrip int
}

func NewIntLengthFieldBasedDecoder(base Decoder) Decoder {
	return NewIntLengthFieldBasedDecoderSize(base, 0, 0, 0)
}

func NewIntLengthFieldBasedDecoderSize(base Decoder, lengthFieldOffset, lengthAdjustment, initialBytesToStrip int) Decoder {
	return &IntLengthFieldBasedDecoder{
		base:                base,
		lengthFieldOffset:   lengthFieldOffset,
		lengthAdjustment:    lengthAdjustment,
		initialBytesToStrip: initialBytesToStrip,
	}
}

func (decoder IntLengthFieldBasedDecoder) Decode(in *ByteBuf) (bool, interface{}, error) {
	readable := in.Readable()

	minFrameLength := decoder.initialBytesToStrip + decoder.lengthFieldOffset + INT_LENGTH_FIELD_LENGTH

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
