package goetty

import (
	"bytes"
	"errors"
	"io"
)

func ReadN(r io.Reader, n int) ([]byte, error) {
	data := make([]byte, n)
	_, err := r.Read(data)

	if err != nil {
		return nil, err
	}

	return data, nil
}

func ReadInt(r io.Reader) (int, error) {
	data, err := ReadN(r, 4)

	if err != nil {
		return 0, err
	}

	return Byte2Int(data), nil
}

func Byte2Int(data []byte) int {
	return int((int(data[0])&0xff)<<24 | (int(data[1])&0xff)<<16 | (int(data[2])&0xff)<<8 | (int(data[3]) & 0xff))
}

func Byte2Int64(data []byte) int64 {
	return int64((int64(data[0])&0xff)<<56 | (int64(data[1])&0xff)<<48 | (int64(data[2])&0xff)<<40 | (int64(data[3])&0xff)<<32 | (int64(data[4])&0xff)<<24 | (int64(data[5])&0xff)<<16 | (int64(data[6])&0xff)<<8 | (int64(data[7]) & 0xff))
}

func WriteInt(v int) []byte {
	ret := make([]byte, 4)
	ret[0] = byte(v >> 24)
	ret[1] = byte(v >> 16)
	ret[2] = byte(v >> 8)
	ret[3] = byte(v)
	return ret
}

func WriteLong(v int64) []byte {
	ret := make([]byte, 8)

	ret[0] = byte(v >> 56)
	ret[1] = byte(v >> 48)
	ret[2] = byte(v >> 40)
	ret[3] = byte(v >> 32)
	ret[4] = byte(v >> 24)
	ret[5] = byte(v >> 16)
	ret[6] = byte(v >> 8)
	ret[7] = byte(v)

	return ret
}

// makeSlice allocates a slice of size n. If the allocation fails, it panics
// with ErrTooLarge.
func makeSlice(n int) []byte {
	// If the make fails, give a known error.
	defer func() {
		if recover() != nil {
			panic(ErrTooLarge)
		}
	}()
	return make([]byte, n)
}

// ----------- bytes buffer, implemention Netty

// +--------------------+--------------------+--------------------+
// | discardable bytes  |   readable bytes   |   writeable bytes  |
// |                    |                    |                    |
// +--------------------+--------------------+--------------------|
// |                    |                    |                    |
// 0      <=       readerIndex    <=     writerIndex    <=     capacity
//
type ByteBuf struct {
	buf         []byte // buf data, auto +/- size
	readerIndex int    //
	writerIndex int    //
	markedIndex int
	scale       int // scale min size
}

var ErrTooLarge = errors.New("goetty.ByteBuf: too large")

const (
	DEFAULT_MIN_SCALE = 512
	GC                = 1024
)

func NewByteBuf(capacity int) *ByteBuf {
	return NewByteBufSize(capacity, DEFAULT_MIN_SCALE)
}

func NewByteBufSize(capacity int, scale int) *ByteBuf {
	return &ByteBuf{
		buf:         make([]byte, capacity),
		readerIndex: 0,
		writerIndex: 0,
		scale:       scale,
	}
}

func (b *ByteBuf) RawBuf() []byte {
	return b.buf
}

func (b *ByteBuf) Clear() {
	b.readerIndex = 0
	b.writerIndex = 0
}

func (b *ByteBuf) Capacity() int {
	return cap(b.buf)
}

func (b *ByteBuf) SetReaderIndex(newReaderIndex int) error {
	if newReaderIndex < 0 || newReaderIndex > b.writerIndex {
		return io.ErrShortBuffer
	}

	b.readerIndex = newReaderIndex

	return nil
}

func (b *ByteBuf) GetReaderIndex() int {
	return b.readerIndex
}

func (b *ByteBuf) GetWriteIndex() int {
	return b.writerIndex
}

func (b *ByteBuf) SetWriterIndex(newWriterIndex int) error {
	if newWriterIndex < b.readerIndex || newWriterIndex > b.Capacity() {
		return io.ErrShortBuffer
	}

	b.writerIndex = newWriterIndex

	return nil
}

func (b *ByteBuf) MarkN(n int) error {
	return b.MarkIndex(b.readerIndex + n)
}

func (b *ByteBuf) MarkIndex(index int) error {
	if index >= b.Capacity() || index <= b.readerIndex {
		return io.ErrShortBuffer
	}

	b.markedIndex = index
	return nil
}

func (b *ByteBuf) Skip(n int) error {
	if n > b.Readable() {
		return io.ErrShortBuffer
	}

	b.readerIndex += n
	return nil
}

func (b *ByteBuf) Readable() int {
	return b.writerIndex - b.readerIndex
}

func (b *ByteBuf) ReadBytes(n int) (int, []byte, error) {
	data := make([]byte, n)
	n, err := b.Read(data)
	return n, data, err
}

func (b *ByteBuf) ReadAll() (int, []byte, error) {
	return b.ReadBytes(b.Readable())
}

func (b *ByteBuf) ReadMarkedBytes() (int, []byte, error) {
	return b.ReadBytes(b.markedIndex - b.readerIndex)
}

func (b *ByteBuf) Read(p []byte) (n int, err error) {
	if len(p) > b.Readable() {
		return 0, io.ErrShortBuffer
	}

	n = copy(p, b.buf[b.readerIndex:b.readerIndex+len(p)])
	b.readerIndex += n
	return n, nil
}

func (b *ByteBuf) PeekInt(offset int) (int, error) {
	if b.Readable() < 4+offset {
		return 0, io.ErrShortBuffer
	}

	start := b.readerIndex + offset
	return ReadInt(bytes.NewReader(b.buf[start : start+4]))
}

func (b *ByteBuf) PeekN(offset int, n int) ([]byte, error) {
	if b.Readable() < n+offset {
		return nil, io.ErrShortBuffer
	}

	start := b.readerIndex + offset
	return ReadN(bytes.NewReader(b.buf[start:start+n]), n)
}

// ReadFrom reads data from r until EOF and appends it to the buffer, growing
// the buffer as needed. The return value n is the number of bytes read. Any
// error except io.EOF encountered during the read is also returned. If the
// buffer becomes too large, ReadFrom will panic with ErrTooLarge.
func (b *ByteBuf) ReadFrom(r io.Reader) (n int64, err error) {
	for {
		b.expansion(b.scale)

		if r == nil {
			return 0, io.EOF
		}

		m, e := r.Read(b.buf[b.writerIndex:b.Capacity()])

		b.buf = b.buf[0 : b.writerIndex+m]

		b.writerIndex += m
		n += int64(m)

		if e == io.EOF {
			return n, e
		}

		if e != nil {
			return n, e
		}

		if n > 0 {
			return n, nil
		}
	}

	return n, nil // err is EOF, so return nil explicitly
}

func (b *ByteBuf) Writeable() int {
	return b.Capacity() - b.writerIndex
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed. The return value n is the length of p; err is always nil. If the
// buffer becomes too large, Write will panic with ErrTooLarge.
func (b *ByteBuf) Write(p []byte) (n int, err error) {
	b.expansion(len(p))
	m, err := b.ReadFrom(bytes.NewBuffer(p))
	return int(m), err
}

func (b *ByteBuf) WriteInt(v int) (n int, err error) {
	b.expansion(4)
	return b.Write(WriteInt(v))
}

func (b *ByteBuf) WriteLong(v int64) (n int, err error) {
	b.expansion(8)
	return b.Write(WriteLong(v))
}

func (b *ByteBuf) WriteByte(v byte) (n int, err error) {
	b.expansion(1)
	return b.Write([]byte{v})
}

func (b *ByteBuf) expansion(n int) {
	ex := b.scale

	if n > b.scale {
		ex = n
	}

	if free := b.Writeable(); free < ex {
		newBuf := makeSlice(cap(b.buf) + ex)
		offset := b.writerIndex - b.readerIndex
		copy(newBuf, b.buf[b.readerIndex:])
		b.readerIndex = 0
		b.writerIndex = offset
		b.buf = newBuf
	}
}
