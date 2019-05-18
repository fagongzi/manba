package goetty

import (
	"errors"
	"hash/crc32"
	"io"
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
	Write(msg interface{}) error
	WriteAndFlush(msg interface{}) error
	Flush() error
	InBuf() *ByteBuf
	OutBuf() *ByteBuf
	SetAttr(key string, value interface{})
	GetAttr(key string) interface{}
	RemoteAddr() string
	RemoteIP() string
}

type clientIOSession struct {
	sync.RWMutex

	id     interface{}
	conn   net.Conn
	closed int32
	svr    *Server
	in     *ByteBuf
	out    *ByteBuf
	attrs  map[string]interface{}
}

func newClientIOSession(id interface{}, conn net.Conn, svr *Server) IOSession {
	conn.(*net.TCPConn).SetNoDelay(true)
	conn.(*net.TCPConn).SetLinger(0)

	return &clientIOSession{
		id:    id,
		conn:  conn,
		svr:   svr,
		attrs: make(map[string]interface{}),
		in:    NewByteBuf(svr.opts.readBufSize),
		out:   NewByteBuf(svr.opts.writeBufSize),
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
	for {
		doRead, msg, err := s.doPreRead()
		if err != nil {
			return nil, err
		}
		if !doRead {
			return msg, nil
		}

		var complete bool
		for {
			if s.in.Readable() > 0 {
				complete, msg, err = s.svr.opts.decoder.Decode(s.in)

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

		returnRead, readedMsg, err := s.doPostRead(msg)
		if err != nil {
			return nil, err
		}

		if returnRead {
			return readedMsg, err
		}
	}
}

// Write wrirte a msg
func (s *clientIOSession) Write(msg interface{}) error {
	return s.write(msg, false)
}

// WriteAndFlush write a msg
func (s *clientIOSession) WriteAndFlush(msg interface{}) error {
	return s.write(msg, true)
}

// InBuf returns internal bytebuf that used for read from server
func (s *clientIOSession) InBuf() *ByteBuf {
	return s.in
}

// OutBuf returns internal bytebuf that used for write to client
func (s *clientIOSession) OutBuf() *ByteBuf {
	return s.out
}

// Flush writes bytes that in the internal bytebuf
func (s *clientIOSession) Flush() error {
	buf := s.out
	written := 0
	all := buf.Readable()
	for {
		if written == all {
			break
		}

		n, err := s.conn.Write(buf.buf[buf.readerIndex+written : buf.writerIndex])
		if err != nil {
			for _, sm := range s.svr.opts.middlewares {
				sm.WriteError(err, s)
			}
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

func (s *clientIOSession) doPreRead() (bool, interface{}, error) {
	for _, sm := range s.svr.opts.middlewares {
		doNext, msg, err := sm.PreRead(s)
		if err != nil {
			return false, false, err
		}

		if !doNext {
			return false, msg, nil
		}
	}

	return true, nil, nil
}

func (s *clientIOSession) doPostRead(msg interface{}) (bool, interface{}, error) {
	readedMsg := msg

	doNext := true
	var err error
	for _, sm := range s.svr.opts.middlewares {
		doNext, readedMsg, err = sm.PostRead(readedMsg, s)
		if err != nil {
			return false, nil, err
		}

		if !doNext {
			return false, readedMsg, nil
		}
	}

	return true, readedMsg, nil
}

func (s *clientIOSession) doPreWrite(msg interface{}) (bool, interface{}, error) {
	var err error
	var doNext bool
	writeMsg := msg

	for _, sm := range s.svr.opts.middlewares {
		doNext, writeMsg, err = sm.PreWrite(writeMsg, s)
		if err != nil {
			return false, writeMsg, err
		}

		if !doNext {
			return false, writeMsg, nil
		}
	}

	return true, writeMsg, nil
}

func (s *clientIOSession) doPostWrite(msg interface{}) error {
	for _, sm := range s.svr.opts.middlewares {
		doNext, err := sm.PostWrite(msg, s)
		if err != nil {
			return err
		}

		if !doNext {
			return nil
		}
	}

	return nil
}

func (s *clientIOSession) write(msg interface{}, flush bool) error {
	doWrite, writeMsg, err := s.doPreWrite(msg)
	if err != nil {
		return err
	}

	if !doWrite {
		return nil
	}

	err = s.svr.opts.encoder.Encode(writeMsg, s.out)
	if err != nil {
		return err
	}

	if flush {
		err = s.Flush()
		if err != nil {
			return err
		}
	}

	return s.doPostWrite(writeMsg)
}

func (s *clientIOSession) readFromConn(timeout time.Duration) (bool, interface{}, error) {
	if 0 != timeout {
		s.conn.SetReadDeadline(time.Now().Add(timeout))
	}

	_, err := io.Copy(s.in, s.conn)
	if err != nil {
		for _, sm := range s.svr.opts.middlewares {
			oerr := sm.ReadError(err, s)
			if oerr == nil {
				return false, nil, nil
			}
		}
		return false, nil, err
	}

	return s.svr.opts.decoder.Decode(s.in)
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
