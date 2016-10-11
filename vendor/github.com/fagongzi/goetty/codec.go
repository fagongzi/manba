package goetty

const (
	BUF_READ_SIZE  = 1024
	BUF_WRITE_SIZE = 1024
)

type Encoder interface {
	Encode(data interface{}, out *ByteBuf) error
}

type Decoder interface {
	Decode(in *ByteBuf) (complete bool, msg interface{}, err error)
}
