package goetty

import (
	"testing"
	"time"
)

var (
	baseCode   = &StringCodec{}
	serverAddr = "127.0.0.1:11111"
	decoder    = NewIntLengthFieldBasedDecoder(baseCode)
	encoder    = NewIntLengthFieldBasedEncoder(baseCode)
)

func TestServerStart(t *testing.T) {
	server := NewServer(serverAddr,
		WithServerDecoder(decoder),
		WithServerEncoder(encoder))

	go func() {
		<-server.Started()
		server.Stop()
	}()

	err := server.Start(func(session IOSession) error { return nil })

	if err != nil {
		t.Error(err)
	}
}

func TestReceivedMsg(t *testing.T) {
	server := NewServer(serverAddr,
		WithServerDecoder(decoder),
		WithServerEncoder(encoder))

	go func() {
		<-server.Started()

		conn := NewConnector(serverAddr,
			WithClientDecoder(decoder),
			WithClientEncoder(encoder))
		_, err := conn.Connect()
		if err != nil {
			server.Stop()
			t.Error(err)
		} else {
			conn.WriteAndFlush("hello")
		}
	}()

	err := server.Start(func(session IOSession) error {
		defer server.Stop()

		msg, err := session.ReadTimeout(time.Second)
		if err != nil {
			t.Error(err)
			return err
		}

		s, ok := msg.(string)
		if !ok {
			t.Error("received err, not string")
		} else {
			if s != "hello" {
				t.Error("received not match")
			}
		}

		return nil
	})

	if err != nil {
		t.Error(err)
	}
}
