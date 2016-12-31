package uuid

import "crypto/rand"

// NewV4 creates a new UUID with variant 4 as described in RFC 4122. Variant 4 based on pure random bytes.
func NewV4() *UUID {
	buf := make([]byte, 16)
	rand.Read(buf)
	buf[6] = (buf[6] & 0x0f) | 0x40
	var uuid UUID
	copy(uuid[:], buf[:])
	uuid.variantRFC4122()
	return &uuid
}
