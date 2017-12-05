package goetty

import (
	"errors"
	"hash/crc32"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrConnectServerSide error for can't connect to client at server side
	ErrConnectServerSide = errors.New("can't connect to client at server side")
)

// IOSession session
type IOSession interface {
	ID() interface{}
	Hash() int
	Close() error
	IsConnected() bool
	Connect() (bool, error)
	Read() (interface{}, error)
	ReadTimeout(timeout time.Duration) (interface{}, error)
	SetBatchSize(size uint64)
	Write(msg interface{}) error
	WriteBatch(msg interface{}) error
	InBuf() *ByteBuf
	OutBuf() *ByteBuf
	WriteOutBuf() error
	SetAttr(key string, value interface{})
	GetAttr(key string) interface{}
	RemoteAddr() string
	RemoteIP() string
}

type clientIOSession struct {
	sync.RWMutex

	id  interface{}
	svr *Server

	conn   net.Conn
	closed int32

	in  *ByteBuf
	out *ByteBuf

	batchLimit uint64
	batchCount uint64

	attrs map[string]interface{}
}

func newClientIOSession(id interface{}, conn net.Conn, svr *Server) IOSession {
	conn.(*net.TCPConn).SetNoDelay(true)
	conn.(*net.TCPConn).SetLinger(0)

	return &clientIOSession{
		id:    id,
		conn:  conn,
		svr:   svr,
		attrs: make(map[string]interface{}),
		in:    NewByteBuf(svr.readBufSize),
		out:   NewByteBuf(svr.writeBufSize),
	}
}

func (s *clientIOSession) Connect() (bool, error) {
	return false, ErrConnectServerSide
}

func (s *clientIOSession) IsConnected() bool {
	return nil != s.conn && atomic.LoadInt32(&s.closed) == 0
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
		if s.in.Readable() > 0 {
			complete, msg, err = s.svr.decoder.Decode(s.in)

			if !complete && err == nil {
				complete, msg, err = s.readFromConn(timeout)
			}
		} else {
			complete, msg, err = s.readFromConn(timeout)
		}

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

func (s *clientIOSession) SetBatchSize(size uint64) {
	s.batchLimit = size
}

func (s *clientIOSession) WriteBatch(msg interface{}) error {
	err := s.svr.encoder.Encode(msg, s.out)

	if err != nil {
		return err
	}

	s.batchCount++

	if s.batchCount%s.batchLimit == 0 {
		return s.WriteOutBuf()
	}

	return nil
}

// InBuf returns internal bytebuf that used for read from server
func (s *clientIOSession) InBuf() *ByteBuf {
	return s.in
}

// OutBuf returns internal bytebuf that used for write to client
func (s *clientIOSession) OutBuf() *ByteBuf {
	return s.out
}

// WriteOutBuf writes bytes that in the internal bytebuf
func (s *clientIOSession) WriteOutBuf() error {
	s.batchCount = 0
	buf := s.out
	written := 0
	all := buf.Readable()
	for {
		if written == all {
			break
		}

		n, err := s.conn.Write(buf.buf[buf.readerIndex+written : buf.writerIndex])
		if err != nil {
			s.out.Clear()
			return err
		}

		written += n
	}

	s.out.Clear()
	return nil
}

// Close close
func (s *clientIOSession) Close() error {
	s.Lock()
	s.closed = 1

	if s.conn == nil {
		return nil
	}

	err := s.conn.Close()
	s.conn = nil
	s.in.Release()
	s.out.Release()
	s.Unlock()

	return err
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

// RemoteIP return remote ip address
func (s *clientIOSession) RemoteIP() string {
	addr := s.RemoteAddr()
	if addr == "" {
		return ""
	}

	return strings.Split(addr, ":")[0]
}

func (s *clientIOSession) readFromConn(timeout time.Duration) (bool, interface{}, error) {
	if 0 != timeout {
		s.conn.SetReadDeadline(time.Now().Add(timeout))
	}

	_, err := s.in.ReadFrom(s.conn)

	if err != nil {
		return false, nil, err
	}

	return s.svr.decoder.Decode(s.in)
}

func getHash(id interface{}) int {
	if v, ok := id.(int64); ok {
		return int(v)
	} else if v, ok := id.(int); ok {
		return v
	} else if v, ok := id.(string); ok {
		return int(crc32.ChecksumIEEE([]byte(v)))
	}

	return 0
}
