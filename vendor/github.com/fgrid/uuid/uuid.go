package uuid

import (
	"encoding/binary"
	"fmt"
)

// The UUID represents Universally Unique IDentifier (which is 128 bit long).
type UUID [16]byte

var (
	// NIL is defined in RFC 4122 section 4.1.7.
	// The nil UUID is special form of UUID that is specified to have all 128 bits set to zero.
	NIL = &UUID{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	}
	// NameSpaceDNS assume name to be a fully-qualified domain name.
	// Declared in RFC 4122 Appendix C.
	NameSpaceDNS = &UUID{
		0x6b, 0xa7, 0xb8, 0x10, 0x9d, 0xad, 0x11, 0xd1,
		0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8,
	}
	// NameSpaceURL assume name to be a URL.
	// Declared in RFC 4122 Appendix C.
	NameSpaceURL = &UUID{
		0x6b, 0xa7, 0xb8, 0x11, 0x9d, 0xad, 0x11, 0xd1,
		0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8,
	}
	// NameSpaceOID assume name to be an ISO OID.
	// Declared in RFC 4122 Appendix C.
	NameSpaceOID = &UUID{
		0x6b, 0xa7, 0xb8, 0x12, 0x9d, 0xad, 0x11, 0xd1,
		0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8,
	}
	// NameSpaceX500 assume name to be a X.500 DN (in DER or a text output format).
	// Declared in RFC 4122 Appendix C.
	NameSpaceX500 = &UUID{
		0x6b, 0xa7, 0xb8, 0x14, 0x9d, 0xad, 0x11, 0xd1,
		0x80, 0xb4, 0x00, 0xc0, 0x4f, 0xd4, 0x30, 0xc8,
	}
)

// Version of the UUID represents a kind of subtype specifier.
func (u *UUID) Version() int {
	return int(binary.BigEndian.Uint16(u[6:8]) >> 12)
}

// String returns the human readable form of the UUID.
func (u *UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}

func (u *UUID) variantRFC4122() {
	u[8] = (u[8] & 0x3f) | 0x80
}
