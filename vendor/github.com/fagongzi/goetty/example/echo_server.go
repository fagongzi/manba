package example

import (
	"fmt"

	"github.com/fagongzi/goetty"
)

// EchoServer echo server
type EchoServer struct {
	addr   string
	server *goetty.Server
}

// NewEchoServer create new server
func NewEchoServer(addr string) *EchoServer {
	return &EchoServer{
		addr: addr,
		server: goetty.NewServer(addr,
			goetty.WithServerDecoder(goetty.NewIntLengthFieldBasedDecoder(&StringDecoder{})),
			goetty.WithServerEncoder(&StringEncoder{})),
	}
}

// Start start
func (e *EchoServer) Start() error {
	return e.server.Start(e.doConnection)
}

func (e *EchoServer) doConnection(session goetty.IOSession) error {
	fmt.Printf("A new connection from <%s>", session.RemoteAddr())

	// start loop for read msg from this connection
	for {
		msg, err := session.Read() // if you want set a read deadline, you can use 'session.ReadTimeout(timeout)'
		if err != nil {
			return err
		}

		fmt.Printf("receive a msg<%s> from <%s>", msg, session.RemoteAddr())

		// echo msg back
		session.Write(msg)
	}
}
