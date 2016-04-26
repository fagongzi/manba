package net

import (
	"container/ring"
	"net"
	"sync"
	"time"
)

type MessageHandler interface {
	MessageReceived(message interface{})
}

type ClientConf struct {
	Conf
	timeoutConnectToServer time.Duration
	timeWheel              *SimpleTimeWheel
	handler                MessageHandler
}

type ClientConfBuilder struct {
	conf *ClientConf
}

func NewClientConfBuilder() *ClientConfBuilder {
	return &ClientConfBuilder{
		conf: &ClientConf{
			Conf: Conf{
				optionProtocol: PROTOCOL_PACKET_LENGTH_FIELD_BASED,
			},
		},
	}
}

func (c *ClientConfBuilder) LimitMaxPacketLength(limit int) *ClientConfBuilder {
	c.conf.limitMaxPacketLength = limit
	return c
}

func (c *ClientConfBuilder) TimeoutConnectToServer(timeout time.Duration) *ClientConfBuilder {
	c.conf.timeoutConnectToServer = timeout
	return c
}

func (c *ClientConfBuilder) TimeoutWrite(timeout time.Duration) *ClientConfBuilder {
	c.conf.timeoutWrite = timeout
	return c
}

func (c *ClientConfBuilder) Heartbeat(heartbeat []byte) *ClientConfBuilder {
	c.conf.heartbeat = heartbeat
	return c
}

func (c *ClientConfBuilder) TimeWheel(timeWheel *SimpleTimeWheel) *ClientConfBuilder {
	c.conf.timeWheel = timeWheel
	return c
}

func (c *ClientConfBuilder) OptionProtocol(protocol Protocol) *ClientConfBuilder {
	c.conf.optionProtocol = protocol
	return c
}

func (c *ClientConfBuilder) Decoder(decoder Decoder) *ClientConfBuilder {
	c.conf.decoder = decoder
	return c
}

func (c *ClientConfBuilder) Encoder(encoder Encoder) *ClientConfBuilder {
	c.conf.encoder = encoder
	return c
}

func (c *ClientConfBuilder) Handler(handler MessageHandler) *ClientConfBuilder {
	c.conf.handler = handler
	return c
}

func (c *ClientConfBuilder) Build() *ClientConf {
	return c.conf
}

func NewConnector() *Connector {
	return &Connector{
		servers: ring.New(1),
		buf:     NewByteBuf(BUFFER_BYTEBUF),
		outBuf:  NewByteBuf(BUFFER_BYTEBUF),
	}
}

// a tcp connector
type Connector struct {
	conf       *ClientConf
	servers    *ring.Ring
	curServer  *ring.Ring
	conn       net.Conn // TCP connection
	connected  bool
	buf        *ByteBuf
	outBuf     *ByteBuf
	timeoutKey string
}

func (c *Connector) SetConf(conf *ClientConf) {
	c.conf = conf
}

func (c *Connector) AddServer(server string) {
	if c.servers.Value == nil {
		c.servers.Value = server
	} else {
		c.servers.Link(&ring.Ring{
			Value: server,
		})
	}
}

func (c *Connector) Connect() (connected bool, err error) {
	e := c.Close() // Close current connection

	if e != nil {
		return false, e
	}

	if c.servers == nil || c.servers.Len() == 0 {
		return false, EmptyServersErr
	}

	c.conn, e = net.DialTimeout("tcp", c.selectServer(), c.conf.timeoutConnectToServer)

	if nil != e {
		return false, e
	}

	c.connected = true

	c.bindTimeoutWrite()

	return true, nil
}

func (c *Connector) Write(message interface{}) error {
	if !c.IsConnected() {
		return IllegalStateErr
	}

	data := c.conf.encoder.Encode(message)
	c.outBuf.Clear()
	withoutDataLength := 0

	switch c.conf.optionProtocol {
	case PROTOCOL_PACKET_LENGTH_FIELD_BASED:
		c.outBuf.WriteInt(len(data))
		withoutDataLength = LENGTH_FILED
		break
	}

	c.outBuf.Write(data)
	sent := len(data) + withoutDataLength

	_, data, _ = c.outBuf.ReadBytes(sent)

	n, err := c.conn.Write(data)

	if err != nil {
		return err
	}

	if n != sent {
		return WriteErr
	}

	c.bindTimeoutWrite()
	return nil

}

// block until error or connection disconnected
func (c *Connector) Loop() {
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go c.loop0(wg)

	wg.Wait()

	return
}

func (c *Connector) IsConnected() bool {
	return c.connected
}

func (c *Connector) Close() error {
	defer c.init()

	if nil != c.conn {
		return c.conn.Close()
	}

	return nil
}

func (c *Connector) init() {
	c.conn = nil
	c.curServer = nil
	c.connected = false
}

func (c *Connector) selectServer() string {
	if nil == c.curServer {
		c.curServer = c.servers.Next()
	} else {
		c.curServer = c.curServer.Next()

		if nil == c.curServer {
			c.curServer = c.servers.Next()
		}
	}

	addr, _ := c.curServer.Value.(string)

	return addr
}

func (c *Connector) loop0(wg *sync.WaitGroup) error {
	defer func() {
		c.Close()
		wg.Done()
	}()

	if !c.IsConnected() {
		return IllegalStateErr
	}

	c.buf.Clear()

	for {
		_, err := c.buf.ReadFrom(c.conn)

		if err != nil {
			return err
		}

		for {
			if c.buf.Readable() > 0 {
				if c.decode() {
					continue
				} else {
					break
				}
			} else {
				c.buf.Clear()
				break
			}
		}
	}
}

func (c *Connector) bindTimeoutWrite() {
	if len(c.timeoutKey) > 0 {
		c.conf.timeWheel.Cancel(c.timeoutKey)
		c.conf.timeWheel.AddWithId(c.conf.Conf.timeoutWrite, c.timeoutKey, c.writeTimeout)
	} else {
		c.timeoutKey = c.conf.timeWheel.Add(c.conf.Conf.timeoutWrite, c.writeTimeout)
	}
}

func (c *Connector) writeTimeout(key string) {
	if c.IsConnected() {
		c.conn.Write(c.conf.heartbeat)
		c.bindTimeoutWrite()
	}
}

func (c *Connector) decode() bool {
	complete := true

	switch c.conf.optionProtocol {
	case PROTOCOL_PACKET_LENGTH_FIELD_BASED:
		complete = c.decodeWithLengthFieldProtocol()
		break
	}

	return complete
}

func (c *Connector) decodeWithLengthFieldProtocol() bool {
	buf := c.buf

	readable := buf.Readable()

	if readable < LENGTH_FILED {
		return false
	}

	n, _ := buf.PeekInt(0)

	if readable < LENGTH_FILED+n {
		return false
	}

	buf.Skip(LENGTH_FILED)

	// heartbeat
	if 0 == n {
		return true
	}

	data := make([]byte, n)
	buf.Read(data)

	// biz decode
	message := c.conf.decoder.Decode(data)

	// do biz
	c.conf.handler.MessageReceived(message)

	return true
}
