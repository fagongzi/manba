package uuid

import (
	"crypto/rand"
	"encoding/binary"
	"net"
	"time"
)

type stamp [10]byte

var (
	mac      []byte
	requests chan bool
	answers  chan stamp
)

const gregorianUnix = 122192928000000000 // nanoseconds between gregorion zero and unix zero

func init() {
	mac = make([]byte, 6)
	rand.Read(mac)
	requests = make(chan bool)
	answers = make(chan stamp)
	go unique()
	i, err := net.Interfaces()
	if err != nil {
		return
	}
	for _, d := range i {
		if len(d.HardwareAddr) == 6 {
			mac = d.HardwareAddr[:6]
			return
		}
	}
}

// NewV1 creates a new UUID with variant 1 as described in RFC 4122.
// Variant 1 is based on hosts MAC address and actual timestamp (as count of 100-nanosecond intervals since
// 00:00:00.00, 15 October 1582 (the date of Gregorian reform to the Christian calendar).
func NewV1() *UUID {
	var uuid UUID
	requests <- true
	s := <-answers
	copy(uuid[:4], s[4:])
	copy(uuid[4:6], s[2:4])
	copy(uuid[6:8], s[:2])
	uuid[6] = (uuid[6] & 0x0f) | 0x10
	copy(uuid[8:10], s[8:])
	copy(uuid[10:], mac)
	uuid.variantRFC4122()
	return &uuid
}

func unique() {
	var (
		lastNanoTicks uint64
		clockSequence [2]byte
	)
	rand.Read(clockSequence[:])

	for range requests {
		var s stamp
		nanoTicks := uint64((time.Now().UTC().UnixNano() / 100) + gregorianUnix)
		if nanoTicks < lastNanoTicks {
			lastNanoTicks = nanoTicks
			rand.Read(clockSequence[:])
		} else if nanoTicks == lastNanoTicks {
			lastNanoTicks = nanoTicks + 1
		} else {
			lastNanoTicks = nanoTicks
		}
		binary.BigEndian.PutUint64(s[:], lastNanoTicks)
		copy(s[8:], clockSequence[:])
		answers <- s
	}
}
