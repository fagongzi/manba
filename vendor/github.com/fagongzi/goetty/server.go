package goetty

import (
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Closer is a closer
type Closer interface {
	Close() error
}

// MessageWriter is a message writer
type MessageWriter interface {
	Write(msg interface{}) error
}

// MessageReader is a message reader
type MessageReader interface {
	Read() (interface{}, error)
	ReadTimeout(timeout time.Duration) (interface{}, error)
}

// IDGenerator ID Generator interface
type IDGenerator interface {
	NewID() interface{}
}

// Int64IDGenerator int64 id Generator
type Int64IDGenerator struct {
	counter int64
}

// NewInt64IDGenerator create a uuid v4 generator
func NewInt64IDGenerator() IDGenerator {
	return &Int64IDGenerator{}
}

// NewID return a id
func (g *Int64IDGenerator) NewID() interface{} {
	return atomic.AddInt64(&g.counter, 1)
}

// UUIDV4IdGenerator uuid v4 generator
type UUIDV4IdGenerator struct {
}

// NewID return a id
func (g *UUIDV4IdGenerator) NewID() interface{} {
	return NewV4UUID()
}

// NewUUIDV4IdGenerator create a uuid v4 generator
func NewUUIDV4IdGenerator() IDGenerator {
	return &UUIDV4IdGenerator{}
}

type sessionMap struct {
	sync.RWMutex
	sessions map[interface{}]IOSession
}

// Server tcp server
type Server struct {
	addr     string
	listener *net.TCPListener

	opts        *serverOptions
	sessionMaps map[int]*sessionMap

	startCh  chan struct{}
	stopOnce *sync.Once
	stopped  bool
}

// NewServer create server
func NewServer(addr string, opts ...ServerOption) *Server {
	sopts := &serverOptions{}
	for _, opt := range opts {
		opt(sopts)
	}
	sopts.adjust()

	s := &Server{
		addr:        addr,
		sessionMaps: make(map[int]*sessionMap, sopts.sessionBucketSize),
		opts:        sopts,

		stopOnce: &sync.Once{},
		startCh:  make(chan struct{}, 1),
	}

	for i := 0; i < sopts.sessionBucketSize; i++ {
		s.sessionMaps[i] = &sessionMap{
			sessions: make(map[interface{}]IOSession),
		}
	}

	return s
}

// Started returns a chan that used for server started
func (s *Server) Started() chan struct{} {
	return s.startCh
}

// Stop stop server
func (s *Server) Stop() {
	s.stopOnce.Do(func() {
		s.stopped = true
		s.listener.Close()
		close(s.startCh)
	})
}

// Start start the server, this method will block until occur a error
func (s *Server) Start(loopFn func(IOSession) error) error {
	addr, err := net.ResolveTCPAddr("tcp", s.addr)

	if err != nil {
		return err
	}

	s.listener, err = net.ListenTCP("tcp", addr)

	if err != nil {
		return err
	}

	s.startCh <- struct{}{}

	var tempDelay time.Duration
	for {
		conn, err := s.listener.AcceptTCP()

		if s.stopped {
			if nil != conn {
				conn.Close()
			}

			return nil
		}

		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			return err
		}
		tempDelay = 0

		session := newClientIOSession(s.opts.generator.NewID(), conn, s)
		s.addSession(session)

		go func() {
			loopFn(session)
			session.Close()
			s.deleteSession(session)
		}()
	}
}

func (s *Server) addSession(session IOSession) {
	m := s.sessionMaps[session.Hash()%DefaultSessionBucketSize]
	m.Lock()
	m.sessions[session.ID()] = session
	m.Unlock()

	for _, sm := range s.opts.middlewares {
		sm.Connected(session)
	}
}

func (s *Server) deleteSession(session IOSession) {
	m := s.sessionMaps[session.Hash()%DefaultSessionBucketSize]
	m.Lock()
	delete(m.sessions, session.ID())
	m.Unlock()

	for _, sm := range s.opts.middlewares {
		sm.Closed(session)
	}
}

// GetSession get session by id
func (s *Server) GetSession(id interface{}) IOSession {
	m := s.sessionMaps[getHash(id)%DefaultSessionBucketSize]
	m.RLock()
	session := m.sessions[id]
	m.RUnlock()
	return session
}
