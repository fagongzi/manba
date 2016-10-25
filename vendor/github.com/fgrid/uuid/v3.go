package uuid

import (
	"crypto/md5"
	"hash"
)

// NewV3 creates a new UUID with variant 3 as described in RFC 4122.
// Variant 3 based namespace-uuid and name and MD-5 hash calculation.
func NewV3(namespace *UUID, name []byte) *UUID {
	uuid := newByHash(md5.New(), namespace, name)
	uuid[6] = (uuid[6] & 0x0f) | 0x30
	return uuid
}

func newByHash(hash hash.Hash, namespace *UUID, name []byte) *UUID {
	hash.Write(namespace[:])
	hash.Write(name[:])

	var uuid UUID
	copy(uuid[:], hash.Sum(nil)[:16])
	uuid.variantRFC4122()
	return &uuid
}
