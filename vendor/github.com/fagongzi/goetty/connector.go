package goetty

import (
	"errors"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// ErrWrite write error
	ErrWrite = errors.New("goetty.net: Write failed")
	// ErrEmptyServers empty server error
	ErrEmptyServers = errors.New("goetty.Connector: Empty servers pool")
	// ErrIllegalState illegal state error
	ErrIllegalState = errors.New("goetty.Connector: Not connected")
)

// Conf config for connector
type Conf struct {
	Addr                   string
	TimeoutConnectToServer time.Duration
	TimeWheel              *TimeoutWheel
	TimeoutWrite           time.Duration
	WriteTimeoutFn         func(string, IOSession)
}

type connector struct {
	sync.RWMutex

	cnf   *Conf
	attrs map[string]interface{}

	lastTimeout Timeout

	conn         net.Conn
	decoder      Decoder
	encoder      Encoder
	in           *ByteBuf
	out          *ByteBuf
	closed       int32
	writeBufSize int

	batchLimit uint64
	batchCount uint64
}

// NewConnector create a new connector
func NewConnector(cnf *Conf, decoder Decoder, encoder Encoder) IOSession {
	return NewConnectorSize(cnf, decoder, encoder, BufReadSize, BufWriteSize)
}

// NewConnectorSize create a new connector
func NewConnectorSize(cnf *Conf, decoder Decoder, encoder Encoder, readBufSize, writeBufSize int) IOSession {
	return &connector{
		cnf:          cnf,
		in:           NewByteBuf(readBufSize),
		out:          NewByteBuf(writeBufSize),
		writeBufSize: writeBufSize,
		decoder:      decoder,
		encoder:      encoder,
		attrs:        make(map[string]interface{}),
	}
}

// InBuf returns internal bytebuf that used for read from client
func (c *connector) InBuf() *ByteBuf {
	return c.in
}

// OutBuf returns internal bytebuf that used for write to client
func (c *connector) OutBuf() *ByteBuf {
	return c.out
}

// WriteOutBuf writes bytes that in the internal bytebuf
func (c *connector) WriteOutBuf() error {
	buf := c.out
	c.batchCount = 0

	written := 0
	all := buf.Readable()
	for {
		if written == all {
			break
		}

		n, err := c.conn.Write(buf.buf[buf.readerIndex+written : buf.writerIndex])
		if err != nil {
			c.writeRelease()
			return err
		}

		written += n
	}

	c.writeRelease()
	return nil
}

// SetAttr add a attr on session
func (c *connector) SetAttr(key string, value interface{}) {
	c.Lock()
	c.attrs[key] = value
	c.Unlock()
}

// GetAttr get attr from session
func (c *connector) GetAttr(key string) interface{} {
	c.RLock()
	v := c.attrs[key]
	c.RUnlock()
	return v
}

// ID get id
func (c *connector) ID() interface{} {
	return 0
}

// Hash get hash value use id
func (c *connector) Hash() int {
	return 0
}

func (c *connector) Connect() (bool, error) {
	e := c.Close() // Close current connection

	if e != nil {
		return false, e
	}

	conn, e := net.DialTimeout("tcp", c.cnf.Addr, c.cnf.TimeoutConnectToServer)

	if nil != e {
		return false, e
	}

	conn.(*net.TCPConn).SetNoDelay(true)
	conn.(*net.TCPConn).SetLinger(0)
	c.conn = conn
	atomic.StoreInt32(&c.closed, 0)
	c.bindWriteTimeout()

	return true, nil
}

// Close close
func (c *connector) Close() error {
	if nil != c.conn {
		err := c.conn.Close()
		if err != nil {
			c.reset()
			return err
		}
	}

	c.reset()
	return nil
}

// IsConnected is connected
func (c *connector) IsConnected() bool {
	return nil != c.conn && atomic.LoadInt32(&c.closed) == 0
}

func (c *connector) reset() {
	atomic.StoreInt32(&c.closed, 1)
	c.conn = nil
	c.in.Clear()
	c.out.Clear()
}

// Read read data from server, block until a msg arrived or  get a error
func (c *connector) Read() (interface{}, error) {
	return c.ReadTimeout(0)
}

// ReadTimeout read data from server with a timeout duration
func (c *connector) ReadTimeout(timeout time.Duration) (interface{}, error) {
	if !c.IsConnected() {
		return nil, ErrIllegalState
	}

	var msg interface{}
	var err error
	var complete bool

	for {
		if c.in.Readable() > 0 {
			complete, msg, err = c.decoder.Decode(c.in)

			if !complete && err == nil {
				complete, msg, err = c.readFromConn(timeout)
			}
		} else {
			complete, msg, err = c.readFromConn(timeout)
		}

		if nil != err {
			c.in.Clear()
			return nil, err
		}

		if complete {
			break
		}
	}

	if c.in.Readable() == 0 {
		c.in.Clear()
	}

	return msg, err
}

// Write write a msg to server
func (c *connector) Write(msg interface{}) error {
	err := c.encoder.Encode(msg, c.out)

	if err != nil {
		return err
	}

	return c.WriteOutBuf()
}

func (c *connector) SetBatchSize(size uint64) {
	c.batchLimit = size
}

func (c *connector) WriteBatch(msg interface{}) error {
	err := c.encoder.Encode(msg, c.out)

	if err != nil {
		return err
	}

	c.batchCount++

	if c.batchCount%c.batchLimit == 0 {
		return c.WriteOutBuf()
	}

	return nil
}

// RemoteAddr get remote address
func (c *connector) RemoteAddr() string {
	if nil != c.conn {
		return c.conn.RemoteAddr().String()
	}

	return ""
}

// RemoteIP return remote ip address
func (c *connector) RemoteIP() string {
	addr := c.RemoteAddr()
	if addr == "" {
		return ""
	}

	return strings.Split(addr, ":")[0]
}

func (c *connector) readFromConn(timeout time.Duration) (bool, interface{}, error) {
	if 0 != timeout {
		c.conn.SetReadDeadline(time.Now().Add(timeout))
	}

	_, err := c.in.ReadFrom(c.conn)

	if err != nil {
		return false, nil, err
	}

	return c.decoder.Decode(c.in)
}

func (c *connector) writeRelease() {
	c.out.Clear()
	c.bindWriteTimeout()
}

func (c *connector) bindWriteTimeout() {
	if c.cnf.WriteTimeoutFn != nil {
		c.lastTimeout.Stop()
		c.lastTimeout, _ = c.cnf.TimeWheel.Schedule(c.cnf.TimeoutWrite, c.writeTimeout, nil)
	}
}

func (c *connector) cancelWriteTimeout() {
	if c.cnf.WriteTimeoutFn != nil {
		c.lastTimeout.Stop()
	}
}

func (c *connector) writeTimeout(arg interface{}) {
	if c.cnf.WriteTimeoutFn != nil {
		c.cnf.WriteTimeoutFn(c.cnf.Addr, c)
	}
}
