package filter

// NewCachedValue returns a cached value
func NewCachedValue(body, contentType []byte) []byte {
	size := len(contentType) + 4 + len(body)
	data := make([]byte, size, size)
	idx := 0
	int2BytesTo(len(contentType), data[0:4])
	idx += 4
	copy(data[idx:idx+len(contentType)], contentType)
	idx += len(contentType)
	copy(data[idx:], body)
	return data
}

// ParseCachedValue returns cached value as content-type and body value
func ParseCachedValue(data []byte) ([]byte, []byte) {
	size := byte2Int(data[0:4])
	return data[4 : 4+size], data[4+size:]
}

func int2BytesTo(v int, ret []byte) {
	ret[0] = byte(v >> 24)
	ret[1] = byte(v >> 16)
	ret[2] = byte(v >> 8)
	ret[3] = byte(v)
}

func byte2Int(data []byte) int {
	return int((int(data[0])&0xff)<<24 | (int(data[1])&0xff)<<16 | (int(data[2])&0xff)<<8 | (int(data[3]) & 0xff))
}
