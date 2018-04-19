package goetty

import (
	"encoding/binary"
	"errors"
	"io"
)

const (
	minScale = 128
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

// Byte2UInt64 byte array to int64 value using big order
func Byte2UInt64(data []byte) uint64 {
	return binary.BigEndian.Uint64(data)
}

// Byte2UInt32 byte array to uint32 value using big order
func Byte2UInt32(data []byte) uint32 {
	return binary.BigEndian.Uint32(data)
}

// Int2BytesTo int value to bytes array using big order
func Int2BytesTo(v int, ret []byte) {
	ret[0] = byte(v >> 24)
	ret[1] = byte(v >> 16)
	ret[2] = byte(v >> 8)
	ret[3] = byte(v)
}

// Int2Bytes int value to bytes array using big order
func Int2Bytes(v int) []byte {
	ret := make([]byte, 4)
	Int2BytesTo(v, ret)
	return ret
}

// Int64ToBytesTo int64 value to bytes array using big order
func Int64ToBytesTo(v int64, ret []byte) {
	ret[0] = byte(v >> 56)
	ret[1] = byte(v >> 48)
	ret[2] = byte(v >> 40)
	ret[3] = byte(v >> 32)
	ret[4] = byte(v >> 24)
	ret[5] = byte(v >> 16)
	ret[6] = byte(v >> 8)
	ret[7] = byte(v)
}

// Uint64ToBytesTo uint64 value to bytes array using big order
func Uint64ToBytesTo(v uint64, ret []byte) {
	binary.BigEndian.PutUint64(ret, v)
}

// Int64ToBytes int64 value to bytes array using big order
func Int64ToBytes(v int64) []byte {
	ret := make([]byte, 8)
	Int64ToBytesTo(v, ret)
	return ret
}

// ByteBuf a buf with byte arrays
//
// | discardable bytes  |   readable bytes   |   writeable bytes  |
// |                    |                    |                    |
// |                    |                    |                    |
// 0      <=       readerIndex    <=     writerIndex    <=     capacity
//
type ByteBuf struct {
	capacity    int
	pool        Pool
	buf         []byte // buf data, auto +/- size
	readerIndex int    //
	writerIndex int    //
	markedIndex int
}

// ErrTooLarge too larger error
var ErrTooLarge = errors.New("goetty.ByteBuf: too large")

// NewByteBuf create a new bytebuf
func NewByteBuf(capacity int) *ByteBuf {
	return NewByteBufPool(capacity, getDefaultMP())
}

// NewByteBufPool create a new bytebuf using a mem pool
func NewByteBufPool(capacity int, pool Pool) *ByteBuf {
	return &ByteBuf{
		capacity:    capacity,
		buf:         pool.Alloc(capacity),
		readerIndex: 0,
		writerIndex: 0,
		pool:        pool,
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
	b.markedIndex = 0
}

// Release release buf
func (b *ByteBuf) Release() {
	b.pool.Free(b.buf)
	b.buf = nil
}

// Resume resume the buf
func (b *ByteBuf) Resume(capacity int) {
	b.buf = b.pool.Alloc(b.capacity)
}

// Capacity get the capacity
func (b *ByteBuf) Capacity() int {
	return len(b.buf) // use len to avoid slice scale
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

// GetMarkerIndex returns markerIndex
func (b *ByteBuf) GetMarkerIndex() int {
	return b.markedIndex
}

// GetMarkedRemind returns size in [readerIndex, markedIndex)
func (b *ByteBuf) GetMarkedRemind() int {
	return b.markedIndex - b.readerIndex
}

// GetMarkedRemindData returns data in [readerIndex, markedIndex)
func (b *ByteBuf) GetMarkedRemindData() []byte {
	return b.buf[b.readerIndex:b.markedIndex]
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
	if index > b.Capacity() || index <= b.readerIndex {
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

// ReadRawBytes read bytes from buf without mem copy
// Note. If used complete, you must call b.Skip(n) to reset reader index
func (b *ByteBuf) ReadRawBytes(n int) (int, []byte, error) {
	if n > b.Readable() {
		return 0, nil, nil
	}

	return n, b.buf[b.readerIndex : b.readerIndex+n], nil
}

// ReadBytes read bytes from buf
// It's will copy the data to a new byte arrary
// return readedBytesCount, byte array, error
func (b *ByteBuf) ReadBytes(n int) (int, []byte, error) {
	data := make([]byte, n)
	n, err := b.Read(data)
	return n, data, err
}

// ReadAll read all data from buf
// It's will copy the data to a new byte arrary
// return readedBytesCount, byte array, error
func (b *ByteBuf) ReadAll() (int, []byte, error) {
	return b.ReadBytes(b.Readable())
}

// ReadMarkedBytes read data from buf in the range [readerIndex, markedIndex)
func (b *ByteBuf) ReadMarkedBytes() (int, []byte, error) {
	return b.ReadBytes(b.GetMarkedRemind())
}

// MarkedBytesReaded reset reader index
func (b *ByteBuf) MarkedBytesReaded() {
	b.readerIndex = b.markedIndex
}

// Read read bytes
// return readedBytesCount, byte array, error
func (b *ByteBuf) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}

	size := len(p)
	if len(p) > b.Readable() {
		size = b.Readable()
	}

	n = copy(p, b.buf[b.readerIndex:b.readerIndex+size])
	b.readerIndex += n
	return n, nil
}

// ReadInt get int value from buf
func (b *ByteBuf) ReadInt() (int, error) {
	if b.Readable() < 4 {
		return 0, io.ErrShortBuffer
	}

	b.readerIndex += 4
	return Byte2Int(b.buf[b.readerIndex-4 : b.readerIndex]), nil
}

// ReadUInt32 get uint32 value from buf
func (b *ByteBuf) ReadUInt32() (uint32, error) {
	if b.Readable() < 8 {
		return 0, io.ErrShortBuffer
	}

	b.readerIndex += 4
	return Byte2UInt32(b.buf[b.readerIndex-4 : b.readerIndex]), nil
}

// ReadInt64 get int64 value from buf
func (b *ByteBuf) ReadInt64() (int64, error) {
	if b.Readable() < 8 {
		return 0, io.ErrShortBuffer
	}

	b.readerIndex += 8
	return Byte2Int64(b.buf[b.readerIndex-8 : b.readerIndex]), nil
}

// ReadUInt64 get uint64 value from buf
func (b *ByteBuf) ReadUInt64() (uint64, error) {
	if b.Readable() < 8 {
		return 0, io.ErrShortBuffer
	}

	b.readerIndex += 8
	return Byte2UInt64(b.buf[b.readerIndex-8 : b.readerIndex]), nil
}

// PeekInt get int value from buf based on currently read index, after read, read index not modifed
func (b *ByteBuf) PeekInt(offset int) (int, error) {
	if b.Readable() < 4+offset {
		return 0, io.ErrShortBuffer
	}

	start := b.readerIndex + offset
	return Byte2Int(b.buf[start : start+4]), nil
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
	return b.buf[start : start+n], nil
}

// ReadFrom reads data from r until EOF and appends it to the buffer, growing
// the buffer as needed. The return value n is the number of bytes read. Any
// error except io.EOF encountered during the read is also returned. If the
// buffer becomes too large, ReadFrom will panic with ErrTooLarge.
func (b *ByteBuf) ReadFrom(r io.Reader) (n int64, err error) {
	for {
		b.Expansion(minScale)
		m, e := r.Read(b.buf[b.writerIndex : b.writerIndex+minScale])
		if m < 0 {
			panic("bug: negative Read")
		}

		b.writerIndex += m
		n += int64(m)
		if e == io.EOF {
			return n, nil // e is EOF, so return nil explicitly
		}
		if e != nil {
			return n, e
		}

		if m < minScale {
			return n, e
		}
	}
}

// Writeable return how many bytes can be wirte into buf
func (b *ByteBuf) Writeable() int {
	return b.Capacity() - b.writerIndex
}

// Write appends the contents of p to the buffer, growing the buffer as
// needed.
func (b *ByteBuf) Write(p []byte) (int, error) {
	n := len(p)
	b.Expansion(n)
	copy(b.buf[b.writerIndex:], p)
	b.writerIndex += n
	return n, nil
}

// WriteInt write int value to buf using big order
// return write bytes count, error
func (b *ByteBuf) WriteInt(v int) (n int, err error) {
	b.Expansion(4)
	Int2BytesTo(v, b.buf[b.writerIndex:b.writerIndex+4])
	b.writerIndex += 4
	return 4, nil
}

// WriteInt64 write int64 value to buf using big order
// return write bytes count, error
func (b *ByteBuf) WriteInt64(v int64) (n int, err error) {
	b.Expansion(8)
	Int64ToBytesTo(v, b.buf[b.writerIndex:b.writerIndex+8])
	b.writerIndex += 8
	return 8, nil
}

// WriteUint64 write uint64 value to buf using big order
// return write bytes count, error
func (b *ByteBuf) WriteUint64(v uint64) (n int, err error) {
	b.Expansion(8)
	Uint64ToBytesTo(v, b.buf[b.writerIndex:b.writerIndex+8])
	b.writerIndex += 8
	return 8, nil
}

// WriteByte write a byte value to buf
func (b *ByteBuf) WriteByte(v byte) error {
	b.Expansion(1)
	b.buf[b.writerIndex] = v
	b.writerIndex++
	return nil
}

// WriteString write a string value to buf
func (b *ByteBuf) WriteString(v string) error {
	_, err := b.Write(StringToSlice(v))
	return err
}

// WriteByteBuf write all readable data to this buf
func (b *ByteBuf) WriteByteBuf(from *ByteBuf) error {
	size := from.Readable()
	b.Expansion(size)
	copy(b.buf[b.writerIndex:b.writerIndex+size], from.buf[from.readerIndex:from.writerIndex])
	b.writerIndex += size
	from.readerIndex = from.writerIndex
	return nil
}

// Expansion expansion buf size
func (b *ByteBuf) Expansion(n int) {
	if free := b.Writeable(); free < n {
		newBuf := b.pool.Alloc(b.Capacity() + n)
		offset := b.writerIndex - b.readerIndex
		copy(newBuf, b.buf[b.readerIndex:b.writerIndex])
		b.readerIndex = 0
		b.writerIndex = offset
		b.pool.Free(b.buf)
		b.buf = newBuf
	}
}
