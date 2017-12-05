package adjust

// Int returns adjust if value is 0
func Int(value, adjust int) int {
	if value == 0 {
		return adjust
	}

	return value
}

// Int8 returns adjust if value is 0
func Int8(value, adjust int8) int8 {
	if value == 0 {
		return adjust
	}

	return value
}

// Int16 returns adjust if value is 0
func Int16(value, adjust int16) int16 {
	if value == 0 {
		return adjust
	}

	return value
}

// Int32 returns adjust if value is 0
func Int32(value, adjust int32) int32 {
	if value == 0 {
		return adjust
	}

	return value
}

// Int64 returns adjust if value is 0
func Int64(value, adjust int64) int64 {
	if value == 0 {
		return adjust
	}

	return value
}

// UInt returns adjust if value is 0
func UInt(value, adjust uint) uint {
	if value == 0 {
		return adjust
	}

	return value
}

// UInt8 returns adjust if value is 0
func UInt8(value, adjust uint8) uint8 {
	if value == 0 {
		return adjust
	}

	return value
}

// UInt16 returns adjust if value is 0
func UInt16(value, adjust uint16) uint16 {
	if value == 0 {
		return adjust
	}

	return value
}

// UInt32 returns adjust if value is 0
func UInt32(value, adjust uint32) uint32 {
	if value == 0 {
		return adjust
	}

	return value
}

// UInt64 returns adjust if value is 0
func UInt64(value, adjust uint64) uint64 {
	if value == 0 {
		return adjust
	}

	return value
}

// String returns adjust if value is nil
func String(value, adjust string) string {
	if value == "" {
		return adjust
	}

	return value
}
