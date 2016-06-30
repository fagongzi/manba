package uuid

import "crypto/sha1"

// NewV5 creates a new UUID with variant 5 as described in RFC 4122.
// Variant 5 based namespace-uuid and name and SHA-1 hash calculation.
func NewV5(namespaceUUID *UUID, name []byte) *UUID {
	uuid := newByHash(sha1.New(), namespaceUUID, name)
	uuid[6] = (uuid[6] & 0x0f) | 0x50
	return uuid
}

// NewNamespaceUUID creates a namespace UUID by using the namespace name in the NIL name space.
// This is a different approach as the 4 "standard" namespace UUIDs which are timebased UUIDs (V1).
func NewNamespaceUUID(namespace string) *UUID {
	return NewV5(NIL, []byte(namespace))
}
