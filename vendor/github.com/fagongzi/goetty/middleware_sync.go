package goetty

import (
	"net"
	"sync"
	"sync/atomic"
)

type syncClientMiddleware struct {
	BaseMiddleware

	bizDecoder                Decoder
	bizEncoder                Encoder
	writer                    func(IOSession, interface{}) error
	cached                    *simpleQueue
	localOffset, serverOffset uint64
	syncing                   bool
	maxReadTimeouts           int
	timeouts                  int
}

// NewSyncProtocolClientMiddleware return a middleware to process sync protocol
func NewSyncProtocolClientMiddleware(bizDecoder Decoder, bizEncoder Encoder, writer func(IOSession, interface{}) error, maxReadTimeouts int) Middleware {
	return &syncClientMiddleware{
		cached:          newSimpleQueue(),
		writer:          writer,
		bizDecoder:      bizDecoder,
		bizEncoder:      bizEncoder,
		maxReadTimeouts: maxReadTimeouts,
	}
}

func (sm *syncClientMiddleware) PreWrite(msg interface{}, conn IOSession) (bool, interface{}, error) {
	// The client side can only send notifySync,notifyRaw,notifyHB msg to the server,
	// wrap the raw biz msg to notifyRaw
	_, isSync := msg.(*notifySync)
	if !isSync {
		_, isSync = msg.(*notifyHB)
	}

	if !isSync {
		m := acquireNotifyRaw()
		err := sm.bizEncoder.Encode(msg, m.buf)
		if err != nil {
			return false, nil, err
		}

		return true, m, nil
	}

	return sm.BaseMiddleware.PreWrite(msg, conn)
}

func (sm *syncClientMiddleware) PreRead(conn IOSession) (bool, interface{}, error) {
	// If there is any biz msg in the queue, returned it and cancel read option
	if sm.cached.len() > 0 {
		return false, sm.cached.pop(), nil
	}

	return true, nil, nil
}

func (sm *syncClientMiddleware) PostRead(msg interface{}, conn IOSession) (bool, interface{}, error) {
	sm.timeouts = 0

	// if read notify msg from server, sync msg with server,
	// the client read option will block until sync complete and read the raw biz msg
	if nt, ok := msg.(*notify); ok {
		if nt.offset == 0 || nt.offset > sm.getLocalOffset() {
			sm.resetServerOffset(nt.offset)
			releaseNotify(nt)

			err := sm.sync(conn)
			if err != nil {
				return false, nil, err
			}
		}

		return false, nil, nil
	} else if rsp, ok := msg.(*notifySyncRsp); ok {
		sm.syncing = false
		sm.resetLocalOffset(rsp.offset)
		for i := byte(0); i < rsp.count; i++ {
			c, rawMsg, err := sm.bizDecoder.Decode(rsp.buf)
			if err != nil {
				return false, nil, err
			}
			if !c {
				panic("data cann't be enough")
			}

			sm.cached.push(rawMsg)
		}
		releaseNotifySyncRsp(rsp)

		if sm.getLocalOffset() < sm.getServerOffset() {
			err := sm.sync(conn)
			if err != nil {
				return false, nil, err
			}
		}

		return false, nil, nil
	}

	return true, msg, nil
}

func (sm *syncClientMiddleware) Closed(conn IOSession) {
	sm.cached = nil
	sm.localOffset = 0
	sm.serverOffset = 0
	sm.syncing = false
	sm.timeouts = 0
}

func (sm *syncClientMiddleware) Connected(conn IOSession) {
	sm.cached = newSimpleQueue()
	sm.localOffset = 0
	sm.serverOffset = 0
	sm.syncing = false
	sm.timeouts = 0
}

func (sm *syncClientMiddleware) ReadError(err error, conn IOSession) error {
	if netErr, ok := err.(*net.OpError); ok &&
		netErr.Timeout() &&
		sm.timeouts < sm.maxReadTimeouts {
		sm.timeouts++
		sm.syncing = false
		return sm.writer(conn, &notifyHB{
			offset: sm.getLocalOffset(),
		})
	}

	return err
}

func (sm *syncClientMiddleware) getLocalOffset() uint64 {
	return atomic.LoadUint64(&sm.localOffset)
}

func (sm *syncClientMiddleware) getServerOffset() uint64 {
	return sm.serverOffset
}

func (sm *syncClientMiddleware) resetLocalOffset(offset uint64) {
	atomic.StoreUint64(&sm.localOffset, offset)
}

func (sm *syncClientMiddleware) resetServerOffset(offset uint64) {
	sm.serverOffset = offset
}

func (sm *syncClientMiddleware) sync(conn IOSession) error {
	if !sm.syncing {
		req := acquireNotifySync()
		req.offset = sm.getLocalOffset()
		err := sm.writer(conn, req)
		if err != nil {
			return err
		}
		sm.syncing = true
	}

	return nil
}

type syncServerMiddleware struct {
	sync.RWMutex
	BaseMiddleware

	bizDecoder     Decoder
	bizEncoder     Encoder
	writer         func(IOSession, interface{}) error
	offsetQueueMap map[interface{}]*OffsetQueue
}

// NewSyncProtocolServerMiddleware return a middleware to process sync protocol
func NewSyncProtocolServerMiddleware(bizDecoder Decoder, bizEncoder Encoder, writer func(IOSession, interface{}) error) Middleware {
	return &syncServerMiddleware{
		bizDecoder:     bizDecoder,
		bizEncoder:     bizEncoder,
		writer:         writer,
		offsetQueueMap: make(map[interface{}]*OffsetQueue),
	}
}

func (sm *syncServerMiddleware) Connected(conn IOSession) {
	sm.Lock()
	sm.offsetQueueMap[conn.ID()] = newOffsetQueue()
	sm.Unlock()
}

func (sm *syncServerMiddleware) Closed(conn IOSession) {
	sm.Lock()
	delete(sm.offsetQueueMap, conn.ID())
	sm.Unlock()
}

func (sm *syncServerMiddleware) PreWrite(msg interface{}, conn IOSession) (bool, interface{}, error) {
	if _, ok := msg.(*notify); ok {
		return sm.BaseMiddleware.PreWrite(msg, conn)
	} else if _, ok := msg.(*notifySyncRsp); ok {
		return sm.BaseMiddleware.PreWrite(msg, conn)
	}

	// add biz msg to the offset queue, and send notify msg to client
	sm.RLock()
	q := sm.offsetQueueMap[conn.ID()]
	sm.RUnlock()

	if q == nil {
		panic("offset queue cann't be nil")
	}

	m := acquireNotify()
	m.offset = q.Add(msg)
	return true, m, nil
}

func (sm *syncServerMiddleware) PostRead(msg interface{}, conn IOSession) (bool, interface{}, error) {
	if m, ok := msg.(*notifySync); ok {
		sm.RLock()
		q := sm.offsetQueueMap[conn.ID()]
		sm.RUnlock()

		if q == nil {
			panic("offset queue cann't be nil")
		}

		// send biz msgs to client
		items, maxOffset := q.Get(m.offset)
		releaseNotifySync(m)

		rsp := acquireNotifySyncRsp()
		rsp.offset = maxOffset
		rsp.count = byte(len(items))
		for _, item := range items {
			err := sm.bizEncoder.Encode(item, rsp.buf)
			if err != nil {
				return false, nil, err
			}
		}
		err := sm.writer(conn, rsp)
		if err != nil {
			return false, nil, err
		}

		return false, nil, nil
	} else if m, ok := msg.(*notifyRaw); ok {
		c, biz, err := sm.bizDecoder.Decode(m.buf)
		releaseNotifyRaw(m)
		if err != nil {
			return false, nil, err
		}
		if !c {
			panic("bug, cann't missing data")
		}

		return true, biz, nil
	} else if m, ok := msg.(*notifyHB); ok {
		sm.RLock()
		q := sm.offsetQueueMap[conn.ID()]
		sm.RUnlock()

		if q == nil {
			panic("offset queue cann't be nil")
		}

		max := q.GetMaxOffset()
		if m.offset < max {
			nt := acquireNotify()
			nt.offset = q.GetMaxOffset()
			err := sm.writer(conn, nt)
			if err != nil {
				return false, nil, err
			}
		}

		return false, nil, nil
	}

	return true, msg, nil
}
