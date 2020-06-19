package filter

import (
	"github.com/fagongzi/goetty"
	"github.com/valyala/fasthttp"
)

// NewCachedValue returns a cached value
func NewCachedValue(resp *fasthttp.Response) *goetty.ByteBuf {
	buf := goetty.NewByteBuf(128)
	buf.WriteInt(0)
	n := 0
	resp.Header.VisitAll(func(key, value []byte) {
		buf.WriteInt(len(key))
		buf.Write(key)
		buf.WriteInt(len(value))
		buf.Write(value)
		n++
	})
	buf.WriteInt(len(resp.Body()))
	buf.Write(resp.Body())

	goetty.Int2BytesTo(n, buf.RawBuf())
	return buf
}

// ReadCachedValueTo read cached value to response
func ReadCachedValueTo(buf *goetty.ByteBuf, resp *fasthttp.Response) {
	headers, _ := buf.ReadInt()
	for i := 0; i < headers; i++ {
		resp.Header.SetBytesKV(readBytes(buf), readBytes(buf))
	}

	resp.SetBody(readBytes(buf))
}

func readBytes(buf *goetty.ByteBuf) []byte {
	n, _ := buf.ReadInt()
	_, value, _ := buf.ReadBytes(n)
	return value
}
