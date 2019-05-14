package util

import (
	"bufio"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valyala/fasthttp"
)

var startTimeUnix = time.Now().Unix()
var clientConnPool sync.Pool

// HTTPOption http client option
type HTTPOption struct {
	// Maximum number of connections which may be established to server
	MaxConns int
	// MaxConnDuration Keep-alive connections are closed after this duration.
	MaxConnDuration time.Duration
	// MaxIdleConnDuration Idle keep-alive connections are closed after this duration.
	MaxIdleConnDuration time.Duration
	// ReadBufferSize Per-connection buffer size for responses' reading.
	ReadBufferSize int
	// WriteBufferSize Per-connection buffer size for requests' writing.
	WriteBufferSize int
	// ReadTimeout Maximum duration for full response reading (including body).
	ReadTimeout time.Duration
	// WriteTimeout Maximum duration for full request writing (including body).
	WriteTimeout time.Duration
	// MaxResponseBodySize Maximum response body size.
	MaxResponseBodySize int
}

// DefaultHTTPOption returns a HTTP Option
func DefaultHTTPOption() *HTTPOption {
	return &HTTPOption{
		MaxConns:            8,
		MaxConnDuration:     time.Minute,
		MaxIdleConnDuration: time.Second * 30,
		ReadBufferSize:      512,
		WriteBufferSize:     256,
		ReadTimeout:         time.Second * 30,
		WriteTimeout:        time.Second * 30,
		MaxResponseBodySize: 1024 * 1024 * 10,
	}
}

// FastHTTPClient fast http client
type FastHTTPClient struct {
	sync.RWMutex

	defaultOption *HTTPOption
	hostClients   map[string]*hostClients
	readerPool    sync.Pool
	writerPool    sync.Pool
}

// NewFastHTTPClient create FastHTTPClient instance
func NewFastHTTPClient() *FastHTTPClient {
	return NewFastHTTPClientOption(nil)
}

// NewFastHTTPClientOption create FastHTTPClient instance with default option
func NewFastHTTPClientOption(defaultOption *HTTPOption) *FastHTTPClient {
	return &FastHTTPClient{
		defaultOption: defaultOption,
		hostClients:   make(map[string]*hostClients),
	}
}

type hostClients struct {
	sync.Mutex

	startedCleaner uint64
	option         *HTTPOption
	lastUseTime    uint32
	connsCount     int
	conns          []*clientConn
}

func (c *hostClients) acquireConn(addr string) (*clientConn, error) {
	var cc *clientConn
	createConn := false
	startCleaner := false

	var n int
	c.Lock()
	n = len(c.conns)
	if n == 0 {
		maxConns := c.option.MaxConns
		if maxConns <= 0 {
			maxConns = fasthttp.DefaultMaxConnsPerHost
		}
		if c.connsCount < maxConns {
			c.connsCount++
			createConn = true
		}
		if createConn && c.connsCount == 1 {
			startCleaner = true
		}
	} else {
		n--
		cc = c.conns[n]
		c.conns = c.conns[:n]
	}
	c.Unlock()

	if cc != nil {
		return cc, nil
	}
	if !createConn {
		return nil, fasthttp.ErrNoFreeConns
	}

	conn, err := dialAddr(addr)
	if err != nil {
		c.decConnsCount()
		return nil, err
	}
	cc = acquireClientConn(conn)

	if startCleaner {
		if atomic.SwapUint64(&c.startedCleaner, 1) == 0 {
			go c.connsCleaner()
		}
	}
	return cc, nil
}

func (c *hostClients) decConnsCount() {
	c.Lock()
	c.connsCount--
	c.Unlock()
}

func (c *hostClients) releaseConn(cc *clientConn) {
	cc.lastUseTime = time.Now()
	c.Lock()
	c.conns = append(c.conns, cc)
	c.Unlock()
}

func (c *hostClients) connsCleaner() {
	var (
		scratch             []*clientConn
		mustStop            bool
		maxIdleConnDuration = c.option.MaxIdleConnDuration
	)

	for {
		currentTime := time.Now()

		c.Lock()
		conns := c.conns
		n := len(conns)
		i := 0
		for i < n && currentTime.Sub(conns[i].lastUseTime) > maxIdleConnDuration {
			i++
		}
		mustStop = (c.connsCount == i)
		scratch = append(scratch[:0], conns[:i]...)
		if i > 0 {
			m := copy(conns, conns[i:])
			for i = m; i < n; i++ {
				conns[i] = nil
			}
			c.conns = conns[:m]
		}
		c.Unlock()

		for i, cc := range scratch {
			c.closeConn(cc)
			scratch[i] = nil
		}
		if mustStop {
			break
		}
		time.Sleep(maxIdleConnDuration)
	}

	atomic.StoreUint64(&c.startedCleaner, 0)
}

func (c *hostClients) closeConn(cc *clientConn) {
	c.decConnsCount()
	cc.c.Close()
	releaseClientConn(cc)
}

type clientConn struct {
	c net.Conn

	createdTime time.Time
	lastUseTime time.Time

	lastReadDeadlineTime  time.Time
	lastWriteDeadlineTime time.Time
}

// Do do a http request
func (c *FastHTTPClient) Do(req *fasthttp.Request, addr string, option *HTTPOption) (*fasthttp.Response, error) {
	resp, err := c.do(req, addr, option)
	return resp, err
}

func (c *FastHTTPClient) do(req *fasthttp.Request, addr string, option *HTTPOption) (*fasthttp.Response, error) {
	resp := fasthttp.AcquireResponse()
	err := c.doNonNilReqResp(req, resp, addr, option)
	return resp, err
}

func (c *FastHTTPClient) doNonNilReqResp(req *fasthttp.Request, resp *fasthttp.Response, addr string, option *HTTPOption) error {
	if req == nil {
		panic("BUG: req cannot be nil")
	}
	if resp == nil {
		panic("BUG: resp cannot be nil")
	}

	opt := option
	if opt == nil {
		opt = c.defaultOption
	}

	var hc *hostClients
	var ok bool
	c.Lock()
	if hc, ok = c.hostClients[addr]; !ok {
		hc = &hostClients{option: opt}
		c.hostClients[addr] = hc
	}
	c.Unlock()

	atomic.StoreUint32(&hc.lastUseTime, uint32(time.Now().Unix()-startTimeUnix))

	// Free up resources occupied by response before sending the request,
	// so the GC may reclaim these resources (e.g. response body).
	resp.Reset()

	cc, err := hc.acquireConn(addr)
	if err != nil {
		return err
	}
	conn := cc.c

	// set write deadline
	if opt.WriteTimeout > 0 {
		// Optimization: update write deadline only if more than 25%
		// of the last write deadline exceeded.
		// See https://github.com/golang/go/issues/15133 for details.
		currentTime := time.Now()
		if err = conn.SetWriteDeadline(currentTime.Add(opt.WriteTimeout)); err != nil {
			hc.closeConn(cc)
			return err
		}
		cc.lastWriteDeadlineTime = currentTime
	}

	resetConnection := false
	if opt.MaxConnDuration > 0 && time.Since(cc.createdTime) > opt.MaxConnDuration && !req.ConnectionClose() {
		req.SetConnectionClose()
		resetConnection = true
	}

	bw := c.acquireWriter(conn, opt)
	err = req.Write(bw)

	if resetConnection {
		req.Header.ResetConnectionClose()
	}

	if err == nil {
		err = bw.Flush()
	}
	if err != nil {
		c.releaseWriter(bw)
		hc.closeConn(cc)
		return err
	}
	c.releaseWriter(bw)

	// set read readline
	if opt.ReadTimeout > 0 {
		// Optimization: update read deadline only if more than 25%
		// of the last read deadline exceeded.
		// See https://github.com/golang/go/issues/15133 for details.
		currentTime := time.Now()
		if err = conn.SetReadDeadline(currentTime.Add(opt.ReadTimeout)); err != nil {
			hc.closeConn(cc)
			return err
		}
		cc.lastReadDeadlineTime = currentTime
	}

	if !req.Header.IsGet() && req.Header.IsHead() {
		resp.SkipBody = true
	}

	br := c.acquireReader(conn, opt)
	if err = resp.ReadLimitBody(br, opt.MaxResponseBodySize); err != nil {
		c.releaseReader(br)
		hc.closeConn(cc)
		if err == io.EOF {
			return err
		}
		return err
	}
	c.releaseReader(br)

	if resetConnection || req.ConnectionClose() || resp.ConnectionClose() {
		hc.closeConn(cc)
	} else {
		hc.releaseConn(cc)
	}

	return err
}

func dialAddr(addr string) (net.Conn, error) {
	conn, err := fasthttp.Dial(addr)
	if err != nil {
		return nil, err
	}
	if conn == nil {
		panic("BUG: DialFunc returned (nil, nil)")
	}

	return conn, nil
}

func (c *FastHTTPClient) acquireWriter(conn net.Conn, opt *HTTPOption) *bufio.Writer {
	v := c.writerPool.Get()
	if v == nil {
		return bufio.NewWriterSize(conn, opt.WriteBufferSize)
	}
	bw := v.(*bufio.Writer)
	bw.Reset(conn)
	return bw
}

func (c *FastHTTPClient) releaseWriter(bw *bufio.Writer) {
	c.writerPool.Put(bw)
}

func (c *FastHTTPClient) acquireReader(conn net.Conn, opt *HTTPOption) *bufio.Reader {
	v := c.readerPool.Get()
	if v == nil {
		return bufio.NewReaderSize(conn, opt.ReadBufferSize)
	}
	br := v.(*bufio.Reader)
	br.Reset(conn)
	return br
}

func (c *FastHTTPClient) releaseReader(br *bufio.Reader) {
	c.readerPool.Put(br)
}

func acquireClientConn(conn net.Conn) *clientConn {
	v := clientConnPool.Get()
	if v == nil {
		v = &clientConn{}
	}
	cc := v.(*clientConn)
	cc.c = conn
	cc.createdTime = time.Now()
	return cc
}

func releaseClientConn(cc *clientConn) {
	cc.c = nil
	clientConnPool.Put(cc)
}
