package net

import (
	"errors"
	"time"
)

var (
	BUFFER_BYTEBUF = 1024
)

var (
	WriteErr        = errors.New("components.net: Write failed")
	EmptyServersErr = errors.New("components.Connector: Empty servers pool")
	IllegalStateErr = errors.New("components.Connector: Not connected")
	ClientClosedErr = errors.New("components.Receiver: Client is Closed")
)

type Protocol int

var (
	LENGTH_FILED                       = 4
	PROTOCOL_PACKET_LENGTH_FIELD_BASED = Protocol(0)
	PROTOCOL_PACKET_CUSTOM             = Protocol(100)
)

const (
	LIMIT_MAX_PACKET_LENGTH = 1024 * 1024 // packet max size 1MB, otherwise close connection
	TIMEOUT_READ            = time.Second * 30
	TIMEOUT_WRITE           = time.Second * 30
)

type Encoder interface {
	Encode(message interface{}) []byte
}

type Decoder interface {
	Decode(data []byte) interface{}
}

type Conf struct {
	limitMaxPacketLength int

	timeoutRead  time.Duration // read from EndPoint Timeout
	timeoutWrite time.Duration // Write from EndPoint Timeout

	optionProtocol Protocol

	heartbeat []byte // when write timeout, keep the connection

	decoder Decoder
	encoder Encoder
}
