package goetty

import (
	"bytes"
	"errors"
	"io"
)

//ReadN read n bytes from a reader
func ReadN(r io.Reader, n int) ([]byte, error) {
	data := make([]byte, n)
	_, err := r.Read(data)

	if err != nil {
		return nil, err
	}

	return data, nil
}

//ReadInt read a int value from a reader
func ReadInt(r io.Reader) (int, error) {
	data, err := ReadN(r, 4)

	if err != nil {
		return 0, err
	}

	return Byte2Int(data), nil
}

// Byte2Int byte array to int value using big order
func Byte2Int(data []byte) int {
	return int((int(data[0])&0xff)<<24 | (int(data[1])&0xff)<<16 | (int(data[2])&0xff)<<8 | (int(data[3]) & 0xff))
}

// Byte2Int64 byte array to int64 value using big order
func Byte2Int64(data []byte) int64 {
	return int64((int64(data[0])&0xff)<<56 | (int64(data[1])&0xff)<<48 | (int64(data[2])&0xff)<<40 | (int64(data[3])&0xff)<<32 | (int64(data[4])&0xff)<<24 | (int64(data[5])&0xff)<<16 | (int64(data[6])&0xff)<<8 | (int64(data[7]) & 0xff))
}

// Int2Bytes int value to bytes array using big order
func Int2Bytes(v int) []byte {
	ret := make([]byte, 4)
	ret[0] = byte(v >> 24)
	ret[1] = byte(v >> 16)
	ret[2] = byte(v >> 8)
	ret[3] = byte(v)
	return ret
}

// Int64ToBytes int64 value to bytes array using big order
func Int64ToBytes(v int64) []byte {
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

// ByteBuf a buf with byte arrays
//
// | discardable bytes  |   readable bytes   |   writeable bytes  |
// |                    |                    |                    |
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

// ErrTooLarge too larger error
var ErrTooLarge = errors.New("goetty.ByteBuf: too large")

const (
	// DefaultScaleMinSize default size for scale byte buf size
	DefaultScaleMinSize = 512
	// GC gc
	GC = 1024
)

// NewByteBuf create a new bytebuf
func NewByteBuf(capacity int) *ByteBuf {
	return NewByteBufSize(capacity, DefaultScaleMinSize)
}

// NewByteBufSize create a new bytebuf using scale size
func NewByteBufSize(capacity int, scale int) *ByteBuf {
	return &ByteBuf{
		buf:         make([]byte, capacity),
		readerIndex: 0,
		writerIndex: 0,
		scale:       scale,
	}
}

// RawBuf get the raw byte array
func (b *ByteBuf) RawBuf() []byte {
	return b.buf
}

// Clear reset the write and read index
func (b *ByteBuf) Clear() {
	b.readerIndex = 0
	b.writerIndex = 0
}

// Capacity get the capacity
func (b *ByteBuf) Capacity() int {
	return cap(b.buf)
}

// SetReaderIndex set the read index
func (b *ByteBuf) SetReaderIndex(newReaderIndex int) error {
	if newReaderIndex < 0 || newReaderIndex > b.writerIndex {
		return io.ErrShortBuffer
	}

	b.readerIndex = newReaderIndex

	return nil
}

// GetReaderIndex get the read index
func (b *ByteBuf) GetReaderIndex() int {
	return b.readerIndex
}

// GetWriteIndex get the write index
func (b *ByteBuf) GetWriteIndex() int {
	return b.writerIndex
}

// SetWriterIndex set the write index
func (b *ByteBuf) SetWriterIndex(newWriterIndex int) error {
	if newWriterIndex < b.readerIndex || newWriterIndex > b.Capacity() {
		return io.ErrShortBuffer
	}

	b.writerIndex = newWriterIndex

	return nil
}

// MarkN mark a index offset based by currently read index
func (b *ByteBuf) MarkN(n int) error {
	return b.MarkIndex(b.readerIndex + n)
}

// MarkIndex mark a index
func (b *ByteBuf) MarkIndex(index int) error {
	if index >= b.Capacity() || index <= b.readerIndex {
		return io.ErrShortBuffer
	}

	b.markedIndex = index
	return nil
}

// Skip skip bytes, after this option, read index will change to readerIndex+n
func (b *ByteBuf) Skip(n int) error {
	if n > b.Readable() {
		return io.ErrShortBuffer
	}

	b.readerIndex += n
	return nil
}

// Readable current readable byte size
func (b *ByteBuf) Readable() int {
	return b.writerIndex - b.readerIndex
}

// ReadByte read a byte from buf
// return byte value, error
func (b *ByteBuf) ReadByte() (byte, error) {
	if b.Readable() == 0 {
		return 0, nil
	}

	v := b.buf[b.readerIndex]
	b.readerIndex++
	return v, nil
}

// ReadBytes read bytes from buf
// return readedBytesCount, byte array, error
func (b *ByteBuf) ReadBytes(n int) (int, []byte, error) {
	data := make([]byte, n)
	n, err := b.Read(data)
	return n, data, err
}

// ReadAll read all data from buf
// return readedBytesCount, byte array, error
func (b *ByteBuf) ReadAll() (int, []byte, error) {
	return b.ReadBytes(b.Readable())
}

// ReadMarkedBytes read data from buf in the range [markedIndex, readerIndex)
func (b *ByteBuf) ReadMarkedBytes() (int, []byte, error) {
	return b.ReadBytes(b.markedIndex - b.readerIndex)
}

// Read read bytes
// return readedBytesCount, byte array, error
func (b *ByteBuf) Read(p []byte) (n int, err error) {
	if len(p) > b.Readable() {
		return 0, nil
	}

	n = copy(p, b.buf[b.readerIndex:b.readerIndex+len(p)])
	b.readerIndex += n
	return n, nil
}

// PeekInt get int value from buf based on currently read index, after read, read index not modifed
func (b *ByteBuf) PeekInt(offset int) (int, error) {
	if b.Readable() < 4+offset {
		return 0, io.ErrShortBuffer
	}

	start := b.readerIndex + offset
	return ReadInt(bytes.NewReader(b.buf[start : start+4]))
}

// PeekByte get byte value from buf based on currently read index, after read, read index not modifed
func (b *ByteBuf) PeekByte(offset int) (byte, error) {
	if b.Readable() < offset || offset < 0 {
		return 0, io.ErrShortBuffer
	}

	return b.buf[b.readerIndex+offset], nil
}

// PeekN get bytes from buf based on currently read index, after read, read index not modifed
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
}

// Writeable return how many bytes can be wirte into buf
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

// WriteInt write int value to buf using big order
// return write bytes count, error
func (b *ByteBuf) WriteInt(v int) (n int, err error) {
	b.expansion(4)
	return b.Write(Int2Bytes(v))
}

// WriteInt64 write int64 value to buf using big order
// return write bytes count, error
func (b *ByteBuf) WriteInt64(v int64) (n int, err error) {
	b.expansion(8)
	return b.Write(Int64ToBytes(v))
}

// WriteByte write a byte value to buf
// return write bytes count, error
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
