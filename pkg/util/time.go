package util

import (
	"time"
)

// NowWithMillisecond returns timestamp with millisecond
func NowWithMillisecond() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
