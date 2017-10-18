package goetty

import (
	"fmt"
	"sync"
)

// ConnectionPool the connect pool
type ConnectionPool struct {
	sync.RWMutex

	conns           map[string]IOSession
	createHandler   func(string) IOSession
	readLoopHandler func(addr string, conn IOSession)
}

// NewConnectionPool returns a connector pool with a read loop and create connection handler function
func NewConnectionPool(createHandler func(string) IOSession, readLoopHandler func(addr string, conn IOSession)) *ConnectionPool {
	return &ConnectionPool{
		conns:           make(map[string]IOSession),
		createHandler:   createHandler,
		readLoopHandler: readLoopHandler,
	}
}

// GetConn returns a IOSession with the target server address
// If can not connect to server, returns a error
func (p *ConnectionPool) GetConn(addr string) (IOSession, error) {
	conn := p.getConnLocked(addr)
	ok, err := p.checkConnect(addr, conn)
	if err != nil {
		return nil, err
	}

	if ok {
		return conn, nil
	}

	return conn, fmt.Errorf("not connected: %s", addr)
}

func (p *ConnectionPool) getConnLocked(addr string) IOSession {
	p.RLock()
	conn := p.conns[addr]
	p.RUnlock()

	if conn != nil {
		return conn
	}

	return p.createConn(addr)
}

func (p *ConnectionPool) createConn(addr string) IOSession {
	p.Lock()

	// double check
	if conn, ok := p.conns[addr]; ok {
		p.Unlock()
		return conn
	}

	conn := p.createHandler(addr)
	p.conns[addr] = conn
	p.Unlock()
	return conn
}

func (p *ConnectionPool) checkConnect(addr string, conn IOSession) (bool, error) {
	if nil == conn {
		return false, nil
	}

	p.Lock()
	if conn.IsConnected() {
		p.Unlock()
		return true, nil
	}

	ok, err := conn.Connect()
	if err != nil {
		p.Unlock()
		return false, err
	}

	if p.readLoopHandler != nil {
		go p.readLoopHandler(addr, conn)
	}

	p.Unlock()
	return ok, nil
}
