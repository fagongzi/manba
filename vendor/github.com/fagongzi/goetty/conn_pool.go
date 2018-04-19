package goetty

import (
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrClosed is the error resulting if the pool is closed via pool.Close().
	ErrClosed = errors.New("pool is closed")
)

// IOSessionPool interface describes a pool implementation. A pool should have maximum
// capacity. An ideal pool is threadsafe and easy to use.
type IOSessionPool interface {
	// Get returns a new connection from the pool. Closing the connections puts
	// it back to the Pool. Closing it when the pool is destroyed or full will
	// be counted as an error.
	Get() (IOSession, error)

	// Put puts the connection back to the pool. If the pool is full or closed,
	// conn is simply closed. A nil conn will be rejected.
	Put(IOSession) error

	// Close closes the pool and all its connections. After Close() the pool is
	// no longer usable.
	Close()

	// Len returns the current number of connections of the pool.
	Len() int
}

// Factory is a function to create new connections.
type Factory func() (IOSession, error)

// NewIOSessionPool returns a new pool based on buffered channels with an initial
// capacity and maximum capacity. Factory is used when initial capacity is
// greater than zero to fill the pool. A zero initialCap doesn't fill the Pool
// until a new Get() is called. During a Get(), If there is no new connection
// available in the pool, a new connection will be created via the Factory()
// method.
func NewIOSessionPool(initialCap, maxCap int, factory Factory) (IOSessionPool, error) {
	if initialCap < 0 || maxCap <= 0 || initialCap > maxCap {
		return nil, errors.New("invalid capacity settings")
	}

	c := &channelPool{
		conns:   make(chan IOSession, maxCap),
		factory: factory,
	}

	// create initial connections, if something goes wrong,
	// just close the pool error out.
	for i := 0; i < initialCap; i++ {
		conn, err := factory()
		if err != nil {
			c.Close()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
		c.conns <- conn
	}

	return c, nil
}

type channelPool struct {
	sync.Mutex

	conns   chan IOSession
	factory Factory
}

func (c *channelPool) getConns() chan IOSession {
	c.Lock()
	conns := c.conns
	c.Unlock()
	return conns
}

// Get implements the Pool interfaces Get() method. If there is no new
// connection available in the pool, a new connection will be created via the
// Factory() method.
func (c *channelPool) Get() (IOSession, error) {
	conns := c.getConns()
	if conns == nil {
		return nil, ErrClosed
	}

	// wrap our connections with out custom net.Conn implementation (wrapConn
	// method) that puts the connection back to the pool if it's closed.
	select {
	case conn := <-conns:
		if conn == nil {
			return nil, ErrClosed
		}

		return conn, nil
	default:
		conn, err := c.factory()
		if err != nil {
			return nil, err
		}

		return conn, nil
	}
}

// Put implements the Pool interfaces Put() method. If the pool is full or closed,
// conn is simply closed. A nil conn will be rejected.
func (c *channelPool) Put(conn IOSession) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}

	c.Lock()

	if c.conns == nil {
		c.Unlock()
		// pool is closed, close passed connection
		return conn.Close()
	}

	// put the resource back into the pool. If the pool is full, this will
	// block and the default case will be executed.
	select {
	case c.conns <- conn:
		c.Unlock()
		return nil
	default:
		// pool is full, close passed connection
		c.Unlock()
		return conn.Close()
	}
}

func (c *channelPool) Close() {
	c.Lock()
	conns := c.conns
	c.conns = nil
	c.factory = nil
	c.Unlock()

	if conns == nil {
		return
	}

	close(conns)
	for conn := range conns {
		conn.Close()
	}
}

func (c *channelPool) Len() int {
	return len(c.getConns())
}

// ConnStatusHandler handler for conn status
type ConnStatusHandler interface {
	ConnectFailed(addr string, err error)
	Connected(addr string, conn IOSession)
}

// AddressBasedPool is a address based conn pool.
// Only one conn per address in the pool.
type AddressBasedPool struct {
	sync.RWMutex

	handler ConnStatusHandler
	factory func(string) IOSession
	conns   map[string]IOSession
}

// NewAddressBasedPool returns a AddressBasedPool with a factory fun
func NewAddressBasedPool(factory func(string) IOSession, handler ConnStatusHandler) *AddressBasedPool {
	return &AddressBasedPool{
		handler: handler,
		factory: factory,
		conns:   make(map[string]IOSession),
	}
}

// GetConn returns a IOSession that connected to the address
// Every address has only one connection in the pool
func (pool *AddressBasedPool) GetConn(addr string) (IOSession, error) {
	conn := pool.getConnLocked(addr)
	if err := pool.checkConnect(addr, conn); err != nil {
		return nil, err
	}

	return conn, nil
}

// RemoveConn close the conn, and remove from the pool
func (pool *AddressBasedPool) RemoveConn(addr string) {
	pool.Lock()
	if conn, ok := pool.conns[addr]; ok {
		conn.Close()
		delete(pool.conns, addr)
	}
	pool.Unlock()
}

// RemoveConnIfMatches close the conn, and remove from the pool if the conn in the pool is match the given
func (pool *AddressBasedPool) RemoveConnIfMatches(addr string, target IOSession) bool {
	removed := false

	pool.Lock()
	if conn, ok := pool.conns[addr]; ok && conn == target {
		conn.Close()
		delete(pool.conns, addr)
		removed = true
	}
	pool.Unlock()

	return removed
}

func (pool *AddressBasedPool) getConnLocked(addr string) IOSession {
	pool.RLock()
	conn := pool.conns[addr]
	pool.RUnlock()

	if conn != nil {
		return conn
	}

	return pool.createConn(addr)
}

func (pool *AddressBasedPool) checkConnect(addr string, conn IOSession) error {
	if nil == conn {
		return fmt.Errorf("nil connection")
	}

	pool.Lock()
	if conn.IsConnected() {
		pool.Unlock()
		return nil
	}

	_, err := conn.Connect()
	if err != nil {
		if pool.handler != nil {
			pool.handler.ConnectFailed(addr, err)
		}
		pool.Unlock()
		return err
	}

	if pool.handler != nil {
		pool.handler.Connected(addr, conn)
	}
	pool.Unlock()
	return nil
}

func (pool *AddressBasedPool) createConn(addr string) IOSession {
	pool.Lock()

	// double check
	if conn, ok := pool.conns[addr]; ok {
		pool.Unlock()
		return conn
	}

	conn := pool.factory(addr)
	pool.conns[addr] = conn
	pool.Unlock()
	return conn
}
