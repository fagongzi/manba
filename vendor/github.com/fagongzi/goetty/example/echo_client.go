package example

import (
	"fmt"
	"time"

	"github.com/fagongzi/goetty"
)

// EchoClient echo client
type EchoClient struct {
	serverAddr string
	conn       goetty.IOSession
}

// NewEchoClient new client
func NewEchoClient(serverAddr string) (*EchoClient, error) {
	c := &EchoClient{
		serverAddr: serverAddr,
	}

	c.conn = goetty.NewConnector(serverAddr,
		goetty.WithClientConnectTimeout(time.Second*3),
		goetty.WithClientDecoder(&StringDecoder{}),
		goetty.WithClientEncoder(&StringEncoder{}),
		// if you want to send heartbeat to server, you can set conf as below, otherwise not set
		goetty.WithClientWriteTimeoutHandler(time.Second*3, c.writeHeartbeat, goetty.NewTimeoutWheel(goetty.WithTickInterval(time.Second))))
	_, err := c.conn.Connect()

	return c, err
}

func (c *EchoClient) writeHeartbeat(serverAddr string, conn goetty.IOSession) {
	c.SendMsg("this is a heartbeat msg")
}

// SendMsg send msg to server
func (c *EchoClient) SendMsg(msg string) error {
	return c.conn.Write(msg)
}

// ReadLoop read loop
func (c *EchoClient) ReadLoop() error {
	// start loop to read msg from server
	for {
		msg, err := c.conn.Read() // if you want set a read deadline, you can use 'connector.ReadTimeout(timeout)'
		if err != nil {
			fmt.Printf("read msg from server<%s> failure", c.serverAddr)
			return err
		}

		fmt.Printf("receive a msg<%s> from <%s>", msg, c.serverAddr)
	}
}
