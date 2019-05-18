package goetty

// Encoder encode interface
type Encoder interface {
	Encode(data interface{}, out *ByteBuf) error
}

// Decoder decoder interface
type Decoder interface {
	Decode(in *ByteBuf) (complete bool, msg interface{}, err error)
}

type emptyDecoder struct{}

func (e *emptyDecoder) Decode(in *ByteBuf) (complete bool, msg interface{}, err error) {
	return true, in, nil
}

type emptyEncoder struct{}

func (e *emptyEncoder) Encode(data interface{}, out *ByteBuf) error {
	return nil
}

// NewEmptyEncoder returns a empty encoder
func NewEmptyEncoder() Encoder {
	return &emptyEncoder{}
}

// NewEmptyDecoder returns a empty decoder
func NewEmptyDecoder() Decoder {
	return &emptyDecoder{}
}
