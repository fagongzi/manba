package net

import (
	"github.com/fgrid/uuid"
	"net"
	"time"
)

type ClientMessageHandler interface {
	MessageReceived(client *Client, message interface{})
	Closed(client *Client)
	Opened(client *Client)
}

type ReceiverConf struct {
	Conf
	timeWheel *HashedTimeWheel
	handler   ClientMessageHandler
}

type ReceiverConfBuilder struct {
	conf *ReceiverConf
}

func NewReceiverConfBuilder() *ReceiverConfBuilder {
	return &ReceiverConfBuilder{
		conf: &ReceiverConf{
			Conf: Conf{
				optionProtocol: PROTOCOL_PACKET_LENGTH_FIELD_BASED,
			},
		},
	}
}

func (c *ReceiverConfBuilder) LimitMaxPacketLength(limit int) *ReceiverConfBuilder {
	c.conf.limitMaxPacketLength = limit
	return c
}

func (c *ReceiverConfBuilder) TimeoutRead(timeout time.Duration) *ReceiverConfBuilder {
	c.conf.timeoutRead = timeout
	return c
}

func (c *ReceiverConfBuilder) TimeWheel(timeWheel *HashedTimeWheel) *ReceiverConfBuilder {
	c.conf.timeWheel = timeWheel
	return c
}

func (c *ReceiverConfBuilder) OptionProtocol(protocol Protocol) *ReceiverConfBuilder {
	c.conf.optionProtocol = protocol
	return c
}

func (c *ReceiverConfBuilder) Decoder(decoder Decoder) *ReceiverConfBuilder {
	c.conf.decoder = decoder
	return c
}

func (c *ReceiverConfBuilder) Encoder(encoder Encoder) *ReceiverConfBuilder {
	c.conf.encoder = encoder
	return c
}

func (c *ReceiverConfBuilder) Handler(handler ClientMessageHandler) *ReceiverConfBuilder {
	c.conf.handler = handler
	return c
}

func (c *ReceiverConfBuilder) Build() *ReceiverConf {
	return c.conf
}

type Receiver struct {
	running bool
	conf    *ReceiverConf
	clients map[string]*Client
}

type Client struct {
	id       string
	createAt time.Time
	conn     *net.TCPConn
	outBuf   *ByteBuf
	inBuf    *ByteBuf
	server   *Receiver
}

func NewReceiver(conf *ReceiverConf) *Receiver {
	return &Receiver{
		conf:    conf,
		clients: make(map[string]*Client),
	}
}

func (r *Receiver) Loop(listen string) error {
	addr, err := net.ResolveTCPAddr("tcp", listen)

	if err != nil {
		return err
	}

	server, err := net.ListenTCP("tcp", addr)

	if err != nil {
		return err
	}

	for {
		conn, err := server.AcceptTCP()

		if err != nil {
			continue
		}

		c := &Client{
			id:       uuid.NewV4().String(),
			conn:     conn,
			createAt: time.Now(),
			inBuf:    NewByteBuf(BUFFER_BYTEBUF),
			outBuf:   NewByteBuf(BUFFER_BYTEBUF),
			server:   r,
		}

		r.connected(c)

		go func() {
			defer r.disconnected(c.id)
			c.loop()
		}()
	}
}

func (r *Receiver) connected(client *Client) {
	r.clients[client.id] = client
	r.conf.handler.Opened(client)
}

func (r *Receiver) disconnected(id string) {
	client, ok := r.clients[id]

	if ok {
		delete(r.clients, id)
		r.conf.handler.Closed(client)
	}
}

func (r *Receiver) Write(id string, message interface{}) error {
	client, ok := r.clients[id]

	if !ok {
		return ClientClosedErr
	}

	return client.Write(message)
}

func (r *Receiver) Broadcast(message interface{}) error {
	return nil
}

func (c *Client) Close() error {
	if c != nil && c.conn != nil {
		return c.conn.Close()
	}

	return nil
}

func (c *Client) loop() error {
	defer func() {
		c.Close()
	}()

	c.inBuf.Clear()

	for {
		now := time.Now().Add(c.server.conf.timeoutRead)
		c.conn.SetReadDeadline(now)
		_, err := c.inBuf.ReadFrom(c.conn)

		if err != nil {
			return err
		}

		for {
			if c.inBuf.Readable() > 0 {
				if c.decode() {
					continue
				} else {
					break
				}
			} else {
				c.inBuf.Clear()
				break
			}
		}
	}
}

func (c *Client) Write(message interface{}) error {
	if nil == c {
		return ClientClosedErr
	}

	data := c.server.conf.encoder.Encode(message)
	c.outBuf.Clear()
	withoutDataLength := 0

	switch c.server.conf.optionProtocol {
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

	return nil
}

func (c *Client) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Client) Id() string {
	return c.id
}

func (c *Client) decode() bool {
	complete := true
	switch c.server.conf.optionProtocol {
	case PROTOCOL_PACKET_LENGTH_FIELD_BASED:
		complete = c.decodeWithLengthFieldProtocol()
		break
	}

	return complete
}

func (c *Client) decodeWithLengthFieldProtocol() bool {
	buf := c.inBuf

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
	_, err := buf.Read(data)

	if err != nil {
		return true
	}

	message := c.server.conf.decoder.Decode(data)

	c.server.conf.handler.MessageReceived(c, message)

	return true
}
