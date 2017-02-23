package util

import (
	"bufio"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fagongzi/gateway/pkg/conf"
	"github.com/valyala/fasthttp"
)

var (
	requestPool  sync.Pool
	responsePool sync.Pool
)

var startTimeUnix = time.Now().Unix()
var clientConnPool sync.Pool

// FastHTTPClient fast http client
type FastHTTPClient struct {
	MaxConnDuration     time.Duration
	MaxIdleConnDuration time.Duration
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration

	MaxConns            int
	MaxResponseBodySize int

	WriteBufferSize int
	ReadBufferSize  int

	clientName  atomic.Value
	lastUseTime uint32

	connsLock  sync.Mutex
	connsCount int
	conns      []*clientConn

	readerPool sync.Pool
	writerPool sync.Pool
}

// NewFastHTTPClient create FastHTTPClient instance
func NewFastHTTPClient(conf *conf.Conf) *FastHTTPClient {
	return &FastHTTPClient{
		MaxConnDuration:     time.Duration(conf.MaxConnDuration) * time.Second,
		MaxIdleConnDuration: time.Duration(conf.MaxIdleConnDuration) * time.Second,
		ReadTimeout:         time.Duration(conf.ReadTimeout) * time.Second,
		WriteTimeout:        time.Duration(conf.WriteTimeout) * time.Second,

		MaxResponseBodySize: conf.MaxResponseBodySize,
		WriteBufferSize:     conf.WriteBufferSize,
		ReadBufferSize:      conf.ReadBufferSize,
		MaxConns:            conf.MaxConns,
	}
}

type clientConn struct {
	c net.Conn

	createdTime time.Time
	lastUseTime time.Time

	lastReadDeadlineTime  time.Time
	lastWriteDeadlineTime time.Time
}

// Do do a http request
func (c *FastHTTPClient) Do(req *fasthttp.Request, addr string) (*fasthttp.Response, error) {
	resp, retry, err := c.do(req, addr)
	if err != nil && retry && isIdempotent(req) {
		resp, _, err = c.do(req, addr)
	}
	if err == io.EOF {
		err = fasthttp.ErrConnectionClosed
	}
	return resp, err
}

func (c *FastHTTPClient) do(req *fasthttp.Request, addr string) (*fasthttp.Response, bool, error) {
	resp := fasthttp.AcquireResponse()

	ok, err := c.doNonNilReqResp(req, resp, addr)

	return resp, ok, err
}

func (c *FastHTTPClient) doNonNilReqResp(req *fasthttp.Request, resp *fasthttp.Response, addr string) (bool, error) {
	if req == nil {
		panic("BUG: req cannot be nil")
	}
	if resp == nil {
		panic("BUG: resp cannot be nil")
	}

	atomic.StoreUint32(&c.lastUseTime, uint32(time.Now().Unix()-startTimeUnix))

	// Free up resources occupied by response before sending the request,
	// so the GC may reclaim these resources (e.g. response body).
	resp.Reset()

	cc, err := c.acquireConn(addr)
	if err != nil {
		return false, err
	}
	conn := cc.c

	// set write deadline
	if c.WriteTimeout > 0 {
		// Optimization: update write deadline only if more than 25%
		// of the last write deadline exceeded.
		// See https://github.com/golang/go/issues/15133 for details.
		currentTime := time.Now()
		if currentTime.Sub(cc.lastWriteDeadlineTime) > (c.WriteTimeout >> 2) {
			if err = conn.SetWriteDeadline(currentTime.Add(c.WriteTimeout)); err != nil {
				c.closeConn(cc)
				return true, err
			}
			cc.lastWriteDeadlineTime = currentTime
		}
	}

	resetConnection := false
	if c.MaxConnDuration > 0 && time.Since(cc.createdTime) > c.MaxConnDuration && !req.ConnectionClose() {
		req.SetConnectionClose()
		resetConnection = true
	}

	bw := c.acquireWriter(conn)
	err = req.Write(bw)

	if resetConnection {
		req.Header.ResetConnectionClose()
	}

	if err == nil {
		err = bw.Flush()
	}
	if err != nil {
		c.releaseWriter(bw)
		c.closeConn(cc)
		return true, err
	}
	c.releaseWriter(bw)

	// set read readline
	if c.ReadTimeout > 0 {
		// Optimization: update read deadline only if more than 25%
		// of the last read deadline exceeded.
		// See https://github.com/golang/go/issues/15133 for details.
		currentTime := time.Now()
		if currentTime.Sub(cc.lastReadDeadlineTime) > (c.ReadTimeout >> 2) {
			if err = conn.SetReadDeadline(currentTime.Add(c.ReadTimeout)); err != nil {
				c.closeConn(cc)
				return true, err
			}
			cc.lastReadDeadlineTime = currentTime
		}
	}

	if !req.Header.IsGet() && req.Header.IsHead() {
		resp.SkipBody = true
	}

	br := c.acquireReader(conn)
	if err = resp.ReadLimitBody(br, c.MaxResponseBodySize); err != nil {
		c.releaseReader(br)
		c.closeConn(cc)
		if err == io.EOF {
			return true, err
		}
		return false, err
	}
	c.releaseReader(br)

	if resetConnection || req.ConnectionClose() || resp.ConnectionClose() {
		c.closeConn(cc)
	} else {
		c.releaseConn(cc)
	}

	return false, err
}

func (c *FastHTTPClient) acquireConn(addr string) (*clientConn, error) {
	var cc *clientConn
	createConn := false
	startCleaner := false

	var n int
	c.connsLock.Lock()
	n = len(c.conns)
	if n == 0 {
		maxConns := c.MaxConns
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
	c.connsLock.Unlock()

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
		go c.connsCleaner()
	}
	return cc, nil
}

func (c *FastHTTPClient) releaseConn(cc *clientConn) {
	cc.lastUseTime = time.Now()
	c.connsLock.Lock()
	c.conns = append(c.conns, cc)
	c.connsLock.Unlock()
}

func (c *FastHTTPClient) connsCleaner() {
	var (
		scratch             []*clientConn
		mustStop            bool
		maxIdleConnDuration = c.MaxIdleConnDuration
	)

	for {
		currentTime := time.Now()

		c.connsLock.Lock()
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
		c.connsLock.Unlock()

		for i, cc := range scratch {
			c.closeConn(cc)
			scratch[i] = nil
		}
		if mustStop {
			break
		}
		time.Sleep(maxIdleConnDuration)
	}
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

func (c *FastHTTPClient) closeConn(cc *clientConn) {
	c.decConnsCount()
	cc.c.Close()
	releaseClientConn(cc)
}

func (c *FastHTTPClient) acquireWriter(conn net.Conn) *bufio.Writer {
	v := c.writerPool.Get()
	if v == nil {
		return bufio.NewWriterSize(conn, c.WriteBufferSize)
	}
	bw := v.(*bufio.Writer)
	bw.Reset(conn)
	return bw
}

func (c *FastHTTPClient) releaseWriter(bw *bufio.Writer) {
	c.writerPool.Put(bw)
}

func (c *FastHTTPClient) acquireReader(conn net.Conn) *bufio.Reader {
	v := c.readerPool.Get()
	if v == nil {
		return bufio.NewReaderSize(conn, c.ReadBufferSize)
	}
	br := v.(*bufio.Reader)
	br.Reset(conn)
	return br
}

func (c *FastHTTPClient) releaseReader(br *bufio.Reader) {
	c.readerPool.Put(br)
}

func (c *FastHTTPClient) decConnsCount() {
	c.connsLock.Lock()
	c.connsCount--
	c.connsLock.Unlock()
}

func isIdempotent(req *fasthttp.Request) bool {
	return req.Header.IsGet() || req.Header.IsHead() || req.Header.IsPut()
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
