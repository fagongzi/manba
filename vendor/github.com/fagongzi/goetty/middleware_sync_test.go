package goetty

import (
	"fmt"
	"testing"
	"time"
)

var (
	baseCodec   = &StringCodec{}
	bizDecoder  = NewIntLengthFieldBasedDecoder(baseCodec)
	bizEncoder  = NewIntLengthFieldBasedEncoder(baseCodec)
	syncCodec   = &SyncCodec{}
	syncDecoder = NewIntLengthFieldBasedDecoder(syncCodec)
	syncEncoder = NewIntLengthFieldBasedEncoder(syncCodec)
)

func TestSyncClientMiddleware(t *testing.T) {
	conn := NewConnector("")

	writer := func(conn IOSession, msg interface{}) error {
		return nil
	}
	sm := NewSyncProtocolClientMiddleware(bizDecoder, bizEncoder, writer, 3)
	sm.(*syncClientMiddleware).cached.push("hello")
	yes, msg, err := sm.PostRead("hello", conn)
	if err != nil {
		t.Error(err)
		return
	}
	if !yes {
		t.Fail()
		return
	}
	if msg != "hello" {
		t.Fail()
		return
	}

	_, msg, _ = sm.PreRead(conn)
	if msg != "hello" {
		t.Fail()
		return
	}

	yes, msg, err = sm.PostRead(acquireNotify(), conn)
	if err != nil {
		t.Error(err)
		return
	}
	if yes {
		t.Fail()
		return
	}
	if msg != nil {
		t.Fail()
		return
	}

	yes, msg, err = sm.PostRead(acquireNotifySyncRsp(), conn)
	if err != nil {
		t.Error(err)
		return
	}
	if yes {
		t.Fail()
		return
	}
	if msg != nil {
		t.Fail()
		return
	}

	sm.(*syncClientMiddleware).writer = func(conn IOSession, msg interface{}) error {
		if _, ok := msg.(*notifySync); !ok {
			t.Fail()
		}
		return nil
	}
	nt := acquireNotify()
	nt.offset = 1
	sm.PostRead(acquireNotify(), conn)

	sm.(*syncClientMiddleware).writer = writer
	rsp := acquireNotifySyncRsp()
	rsp.offset = 1
	rsp.count = 1
	bizEncoder.Encode("hello2", rsp.buf)
	sm.PostRead(rsp, conn)
	_, msg, _ = sm.PreRead(conn)
	if msg != "hello2" {
		t.Fail()
		return
	}

	yes, msg, err = sm.PreWrite("hello", conn)
	if err != nil {
		t.Error(err)
		return
	}
	if !yes {
		t.Fail()
		return
	}
	if _, ok := msg.(*notifyRaw); !ok {
		t.Fail()
		return
	}

	yes, msg, err = sm.PreWrite(acquireNotifySync(), conn)
	if err != nil {
		t.Error(err)
		return
	}
	if !yes {
		t.Fail()
		return
	}
	if _, ok := msg.(*notifySync); !ok {
		t.Fail()
		return
	}
}

func TestSyncServerMiddleware(t *testing.T) {
	conn := &clientIOSession{
		id:  int64(1),
		in:  NewByteBuf(32),
		out: NewByteBuf(32),
	}

	writer := func(conn IOSession, msg interface{}) error {
		return nil
	}
	sm := NewSyncProtocolServerMiddleware(bizDecoder, bizEncoder, writer)
	sm.Connected(conn)
	if 1 != len(sm.(*syncServerMiddleware).offsetQueueMap) {
		t.Fail()
		return
	}

	sm.Closed(conn)
	if 0 != len(sm.(*syncServerMiddleware).offsetQueueMap) {
		t.Fail()
		return
	}

	sm.Connected(conn)
	yes, msg, err := sm.PreWrite(acquireNotify(), conn)
	if err != nil {
		t.Error(err)
		return
	}
	if !yes {
		t.Fail()
		return
	}
	if _, ok := msg.(*notify); !ok {
		t.Fail()
		return
	}

	yes, msg, err = sm.PreWrite(acquireNotifySyncRsp(), conn)
	if err != nil {
		t.Error(err)
		return
	}
	if !yes {
		t.Fail()
		return
	}
	if _, ok := msg.(*notifySyncRsp); !ok {
		t.Fail()
		return
	}

	yes, msg, err = sm.PreWrite("hello", conn)
	if err != nil {
		t.Error(err)
		return
	}
	if !yes {
		t.Fail()
		return
	}
	if _, ok := msg.(*notify); !ok {
		t.Fail()
		return
	}

	sm.Closed(conn)
	sm.Connected(conn)
	sm.PreWrite("hello", conn)
	sm.(*syncServerMiddleware).writer = func(conn IOSession, msg interface{}) error {
		if _, ok := msg.(*notifySyncRsp); !ok {
			return fmt.Errorf("unexpect msg")
		}
		return nil
	}
	yes, msg, err = sm.PostRead(acquireNotifySync(), conn)
	if err != nil {
		t.Error(err)
		return
	}
	if yes {
		t.Fail()
		return
	}
	if msg != nil {
		t.Fail()
		return
	}

	sm.Closed(conn)
	sm.Connected(conn)
	sm.PreWrite("hello", conn)
	sm.(*syncServerMiddleware).writer = writer
	yes, msg, err = sm.PostRead("hello", conn)
	if err != nil {
		t.Error(err)
		return
	}
	if !yes {
		t.Fail()
		return
	}
	if msg != "hello" {
		t.Fail()
		return
	}
}

func TestSyncMiddleware(t *testing.T) {
	addr := "127.0.0.1:12345"
	svr := NewServer(addr,
		WithServerDecoder(syncDecoder),
		WithServerEncoder(syncEncoder),
		WithServerMiddleware(NewSyncProtocolServerMiddleware(bizDecoder, bizEncoder, func(conn IOSession, msg interface{}) error {
			return conn.WriteAndFlush(msg)
		})))

	go func() {
		<-svr.Started()
		conn := NewConnector(addr,
			WithClientDecoder(syncDecoder),
			WithClientEncoder(syncEncoder),
			WithClientMiddleware(NewSyncProtocolClientMiddleware(bizDecoder, bizEncoder, func(conn IOSession, msg interface{}) error {
				return conn.WriteAndFlush(msg)
			}, 3)))

		_, err := conn.Connect()
		if err != nil {
			svr.Stop()
			t.Error(err)
		} else {
			conn.WriteAndFlush("hello")
		}
	}()

	err := svr.Start(func(session IOSession) error {
		defer svr.Stop()

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
