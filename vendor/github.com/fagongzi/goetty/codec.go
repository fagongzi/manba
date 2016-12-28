package goetty

const (
	// BufReadSize read buf size
	BufReadSize = 1024
	// BufWriteSize write buf size
	BufWriteSize = 1024
)

// Encoder encode interface
type Encoder interface {
	Encode(data interface{}, out *ByteBuf) error
}

// Decoder decoder interface
type Decoder interface {
	Decode(in *ByteBuf) (complete bool, msg interface{}, err error)
}
