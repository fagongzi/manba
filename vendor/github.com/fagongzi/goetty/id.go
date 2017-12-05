package goetty

// NewKey get a new Key
func NewKey() string {
	return NewV4UUID()
}

// NewV1UUID new v1 uuid
func NewV1UUID() string {
	return NewV1().String()
}

// NewV4UUID new v4 uuid
func NewV4UUID() string {
	return NewV4().String()
}

// NewV4Bytes new byte array v4 uuid
func NewV4Bytes() []byte {
	return NewV4().Bytes()
}

// NewV1Bytes new byte array v1 uuid
func NewV1Bytes() []byte {
	return NewV1().Bytes()
}
