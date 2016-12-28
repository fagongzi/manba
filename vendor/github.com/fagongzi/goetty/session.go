package goetty

import (
	"net"
	"sync"
	"time"
)

// IOSession session
type IOSession interface {
	ID() interface{}
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

// Read read a msg, block until read msg or get a error
func (s clientIOSession) Read() (interface{}, error) {
	return s.ReadTimeout(0)
}

// ReadTimeout read a msg  with a timeout duration
func (s clientIOSession) ReadTimeout(timeout time.Duration) (interface{}, error) {
	var msg interface{}
	var err error
	var complete bool

	for {
		if 0 != timeout {
			s.conn.SetReadDeadline(time.Now().Add(timeout))
		}

		_, err = s.buf.ReadFrom(s.conn)

		if err != nil {
			s.buf.Clear()
			return nil, err
		}

		complete, msg, err = s.svr.decoder.Decode(s.buf)

		if nil != err {
			s.buf.Clear()
			return nil, err
		}

		if complete {
			break
		}
	}

	if s.buf.Readable() == 0 {
		s.buf.Clear()
	}

	return msg, err
}

// Write wrirte a msg
func (s clientIOSession) Write(msg interface{}) error {
	buf, ok := out.Get().(*ByteBuf)

	if !ok {
		buf = NewByteBuf(s.svr.writeBufSize)
	}

	err := s.svr.encoder.Encode(msg, buf)

	if err != nil {
		buf.Clear()
		out.Put(buf)
		return err
	}

	_, bytes, _ := buf.ReadAll()

	n, err := s.conn.Write(bytes)

	if err != nil {
		buf.Clear()
		out.Put(buf)
		return err
	}

	if n != len(bytes) {
		buf.Clear()
		out.Put(buf)
		return ErrWrite
	}

	buf.Clear()
	out.Put(buf)
	return nil
}

// Close close
func (s clientIOSession) Close() error {
	return s.conn.Close()
}

// Id get id
func (s clientIOSession) ID() interface{} {
	return s.id
}

func (s clientIOSession) Hash() int {
	return getHash(s.id)
}

func (s clientIOSession) SetAttr(key string, value interface{}) {
	s.Lock()
	s.attrs[key] = value
	s.Unlock()
}

func (s clientIOSession) GetAttr(key string) interface{} {
	s.Lock()
	v := s.attrs[key]
	s.Unlock()
	return v
}

// RemoteAddr get remote address
func (s clientIOSession) RemoteAddr() string {
	if nil != s.conn {
		return s.conn.RemoteAddr().String()
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
