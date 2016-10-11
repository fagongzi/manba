package goetty

import (
	"net"
	"sync"
	"time"
)

type IOSession interface {
	Id() interface{}
	Hash() int
	Close() error
	Read() (interface{}, error)
	ReadTimeout(timeout time.Duration) (interface{}, error)
	Write(msg interface{}) error
	SetAttr(key string, value interface{})
	GetAttr(key string) interface{}
	RemoteAddr() string
}

type clientIOSession struct {
	sync.RWMutex
	id    interface{}
	conn  net.Conn
	svr   *Server
	attrs map[string]interface{}
	buf   *ByteBuf
}

func newClientIOSession(id interface{}, conn net.Conn, svr *Server) IOSession {
	return &clientIOSession{
		id:    id,
		conn:  conn,
		svr:   svr,
		attrs: make(map[string]interface{}),
		buf:   NewByteBuf(svr.readBufSize),
	}
}

func (self clientIOSession) Read() (interface{}, error) {
	return self.ReadTimeout(0)
}

func (self clientIOSession) ReadTimeout(timeout time.Duration) (interface{}, error) {
	var msg interface{}
	var err error
	var complete bool

	for {
		if 0 != timeout {
			self.conn.SetReadDeadline(time.Now().Add(timeout))
		}

		_, err = self.buf.ReadFrom(self.conn)

		if err != nil {
			self.buf.Clear()
			return nil, err
		}

		complete, msg, err = self.svr.decoder.Decode(self.buf)

		if nil != err {
			self.buf.Clear()
			return nil, err
		}

		if complete {
			break
		}
	}

	if self.buf.Readable() == 0 {
		self.buf.Clear()
	}

	return msg, err
}

func (self clientIOSession) Write(msg interface{}) error {
	buf, ok := out.Get().(*ByteBuf)

	if !ok {
		buf = NewByteBuf(self.svr.writeBufSize)
	}

	err := self.svr.encoder.Encode(msg, buf)

	if err != nil {
		buf.Clear()
		out.Put(buf)
		return err
	}

	_, bytes, _ := buf.ReadAll()

	n, err := self.conn.Write(bytes)

	if err != nil {
		buf.Clear()
		out.Put(buf)
		return err
	}

	if n != len(bytes) {
		buf.Clear()
		out.Put(buf)
		return WriteErr
	}

	buf.Clear()
	out.Put(buf)
	return nil
}

func (self clientIOSession) Close() error {
	return self.conn.Close()
}

func (self clientIOSession) Id() interface{} {
	return self.id
}

func (self clientIOSession) Hash() int {
	return getHash(self.id)
}

func (self clientIOSession) SetAttr(key string, value interface{}) {
	self.Lock()
	self.attrs[key] = value
	self.Unlock()
}

func (self clientIOSession) GetAttr(key string) interface{} {
	self.Lock()
	v := self.attrs[key]
	self.Unlock()
	return v
}

func (self clientIOSession) RemoteAddr() string {
	if nil != self.conn {
		return self.conn.RemoteAddr().String()
	}

	return ""
}

func getHash(id interface{}) int {
	if v, ok := id.(int64); ok {
		return int(v)
	} else if v, ok := id.(int); ok {
		return v
	} else if v, ok := id.(string); ok {
		return hashCode(v)
	}

	return 0
}

