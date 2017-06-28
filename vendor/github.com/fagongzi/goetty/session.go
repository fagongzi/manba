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
	OutBuf() *ByteBuf
	WriteOutBuf() error
	SetAttr(key string, value interface{})
	GetAttr(key string) interface{}
	RemoteAddr() string
}

type clientIOSession struct {
	id   interface{}
	conn net.Conn
	svr  *Server

	sync.RWMutex
	attrs map[string]interface{}

	in  *ByteBuf
	out *ByteBuf
}

func newClientIOSession(id interface{}, conn net.Conn, svr *Server) IOSession {
	return &clientIOSession{
		id:    id,
		conn:  conn,
		svr:   svr,
		attrs: make(map[string]interface{}),
		in:    NewByteBuf(svr.readBufSize),
		out:   NewByteBuf(svr.writeBufSize),
	}
}

// Read read a msg, block until read msg or get a error
func (s *clientIOSession) Read() (interface{}, error) {
	return s.ReadTimeout(0)
}

// ReadTimeout read a msg  with a timeout duration
func (s *clientIOSession) ReadTimeout(timeout time.Duration) (interface{}, error) {
	var msg interface{}
	var err error
	var complete bool

	for {
		if 0 != timeout {
			s.conn.SetReadDeadline(time.Now().Add(timeout))
		}

		_, err = s.in.ReadFrom(s.conn)

		if err != nil {
			s.in.Clear()
			return nil, err
		}

		complete, msg, err = s.svr.decoder.Decode(s.in)

		if nil != err {
			s.in.Clear()
			return nil, err
		}

		if complete {
			break
		}
	}

	if s.in.Readable() == 0 {
		s.in.Clear()
	}

	return msg, err
}

// Write wrirte a msg
func (s *clientIOSession) Write(msg interface{}) error {
	err := s.svr.encoder.Encode(msg, s.out)

	if err != nil {
		return err
	}

	return s.WriteOutBuf()
}

// OutBuf returns internal bytebuf that used for write to client
func (s *clientIOSession) OutBuf() *ByteBuf {
	return s.out
}

// WriteOutBuf writes bytes that in the internal bytebuf
func (s *clientIOSession) WriteOutBuf() error {
	_, bytes, _ := s.out.ReadAll()

	n, err := s.conn.Write(bytes)

	if err != nil {
		s.out.Clear()
		return err
	}

	if n != len(bytes) {
		s.out.Clear()
		return ErrWrite
	}

	s.out.Clear()
	return nil
}

// Close close
func (s *clientIOSession) Close() error {
	return s.conn.Close()
}

// ID get id
func (s *clientIOSession) ID() interface{} {
	return s.id
}

// Hash get hash value use id
func (s *clientIOSession) Hash() int {
	return getHash(s.id)
}

// SetAttr add a attr on session
func (s *clientIOSession) SetAttr(key string, value interface{}) {
	s.Lock()
	s.attrs[key] = value
	s.Unlock()
}

// GetAttr get attr from session
func (s *clientIOSession) GetAttr(key string) interface{} {
	s.RLock()
	v := s.attrs[key]
	s.RUnlock()
	return v
}

// RemoteAddr get remote address
func (s *clientIOSession) RemoteAddr() string {
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
