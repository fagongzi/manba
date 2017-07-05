package goetty

import (
	"testing"
	"time"
)

// StringEncoder string encoder
type StringEncoder struct {
}

// StringDecoder string decoder
type StringDecoder struct {
}

// NewStringEncoder create a string encoder
func NewStringEncoder() Encoder {
	return &StringEncoder{}
}

// Encode encod
func (e StringEncoder) Encode(data interface{}, out *ByteBuf) error {
	msg, _ := data.(string)
	b := []byte(msg)

	out.WriteInt(len(b))
	out.Write(b)

	return nil
}

// NewStringDecoder create a string decoder
func NewStringDecoder() Decoder {
	return &StringDecoder{}
}

// Decode decode
func (d StringDecoder) Decode(in *ByteBuf) (complete bool, msg interface{}, err error) {
	_, data, err := in.ReadMarkedBytes()

	if err != nil {
		return true, nil, err
	}

	return true, string(data), nil
}

var (
	serverAddr = "127.0.0.1:12345"
	decoder    = NewIntLengthFieldBasedDecoder(NewStringDecoder())
	encoder    = NewStringEncoder()
)

func TestServerStart(t *testing.T) {
	server := NewServer(serverAddr, decoder, encoder, NewInt64IDGenerator())

	go func() {
		time.Sleep(time.Second * 2)
		server.Stop()
	}()

	err := server.Start(func(session IOSession) error { return nil })

	if err != nil {
		t.Error(err)
	}
}

func TestReceivedMsg(t *testing.T) {
	server := NewServer(serverAddr, NewIntLengthFieldBasedDecoder(NewStringDecoder()), NewStringEncoder(), NewInt64IDGenerator())

	go func() {
		tw := NewHashedTimeWheel(time.Second, 60, 2)
		tw.Start()

		time.Sleep(time.Second * 2)
		cnf := &Conf{
			Addr:                   serverAddr,
			TimeoutWrite:           time.Second * 2,
			TimeoutConnectToServer: time.Second * 5,
			TimeWheel:              tw,
		}
		conn := NewConnector(cnf, decoder, encoder)
		_, err := conn.Connect()
		if err != nil {
			server.Stop()
			t.Error(err)
		} else {
			conn.Write("hello")
		}
	}()

	err := server.Start(func(session IOSession) error {
		defer server.Stop()

		msg, err := session.Read()
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
