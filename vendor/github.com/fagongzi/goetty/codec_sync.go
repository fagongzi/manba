package goetty

import (
	"fmt"
	"sync"
)

const (
	cmdNotify        = byte(1)
	cmdNotifySync    = byte(2)
	cmdNotifySyncRsp = byte(3)
	cmdNotifyRaw     = byte(4)
	cmdNotifyHB      = byte(5)
)

var (
	notifyRawPool     sync.Pool
	notifyPool        sync.Pool
	notifySyncPool    sync.Pool
	notifySyncRspPool sync.Pool
)

type notifyRaw struct {
	buf *ByteBuf
}

type notifyHB struct {
	offset uint64
}

type notify struct {
	offset uint64
}

type notifySync struct {
	offset uint64
}

type notifySyncRsp struct {
	offset uint64
	count  byte
	buf    *ByteBuf
}

func acquireNotifyRaw() *notifyRaw {
	v := notifyRawPool.Get()
	if v == nil {
		v = &notifyRaw{
			buf: NewByteBuf(64),
		}
	}

	return v.(*notifyRaw)
}

func releaseNotifyRaw(value *notifyRaw) {
	value.buf.Clear()
	notifyRawPool.Put(value)
}

func acquireNotify() *notify {
	v := notifyPool.Get()
	if v == nil {
		v = &notify{}
	}

	return v.(*notify)
}

func releaseNotify(value *notify) {
	value.offset = 0
	notifyPool.Put(value)
}

func acquireNotifySync() *notifySync {
	v := notifySyncPool.Get()
	if v == nil {
		v = &notifySync{}
	}

	return v.(*notifySync)
}

func releaseNotifySync(value *notifySync) {
	value.offset = 0
	notifySyncPool.Put(value)
}

func acquireNotifySyncRsp() *notifySyncRsp {
	v := notifySyncRspPool.Get()
	if v == nil {
		v = &notifySyncRsp{
			buf: NewByteBuf(64),
		}
	}

	return v.(*notifySyncRsp)
}

func releaseNotifySyncRsp(value *notifySyncRsp) {
	value.buf.Clear()
	value.offset = 0
	value.count = 0
	notifySyncRspPool.Put(value)
}

// SyncCodec sync protocol dercoder and encoder
type SyncCodec struct{}

// Decode decode with raw byte array
func (codec *SyncCodec) Decode(in *ByteBuf) (bool, interface{}, error) {
	cmd, _ := in.ReadByte()
	if cmd == cmdNotify {
		value := acquireNotify()
		value.offset, _ = in.ReadUInt64()
		return true, value, nil
	} else if cmd == cmdNotifySync {
		value := acquireNotifySync()
		value.offset, _ = in.ReadUInt64()
		return true, value, nil
	} else if cmd == cmdNotifySyncRsp {
		value := acquireNotifySyncRsp()
		value.offset, _ = in.ReadUInt64()
		value.count, _ = in.ReadByte()
		value.buf.Write(in.GetMarkedRemindData())
		in.MarkedBytesReaded()
		return true, value, nil
	} else if cmd == cmdNotifyRaw {
		value := acquireNotifyRaw()
		value.buf.Write(in.GetMarkedRemindData())
		in.MarkedBytesReaded()
		return true, value, nil
	} else if cmd == cmdNotifyHB {
		value := &notifyHB{}
		value.offset, _ = in.ReadUInt64()
		return true, value, nil
	}

	return false, nil, nil
}

// Encode encode sync protocol
func (codec *SyncCodec) Encode(data interface{}, out *ByteBuf) error {
	if msg, ok := data.(*notifyRaw); ok {
		out.WriteByte(cmdNotifyRaw)
		out.WriteByteBuf(msg.buf)
		releaseNotifyRaw(msg)
		return nil
	} else if msg, ok := data.(*notify); ok {
		out.WriteByte(cmdNotify)
		out.WriteUint64(msg.offset)
		releaseNotify(msg)
		return nil
	} else if msg, ok := data.(*notifySync); ok {
		out.WriteByte(cmdNotifySync)
		out.WriteUint64(msg.offset)
		releaseNotifySync(msg)
		return nil
	} else if msg, ok := data.(*notifySyncRsp); ok {
		out.WriteByte(cmdNotifySyncRsp)
		out.WriteUint64(msg.offset)
		out.WriteByte(msg.count)
		out.WriteByteBuf(msg.buf)
		releaseNotifySyncRsp(msg)
		return nil
	} else if msg, ok := data.(*notifyHB); ok {
		out.WriteByte(cmdNotifyHB)
		out.WriteUint64(msg.offset)
		return nil
	}

	return fmt.Errorf("not support msg: %v", data)
}
