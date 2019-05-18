package util

import (
	"fmt"
)

const (
	// MinAddrFormat min addr format
	MinAddrFormat = "000000000000000000000"
	// MaxAddrFormat max addr format
	MaxAddrFormat = "255.255.255.255:99999"
)

// GetAddrFormat returns addr format for sort, padding left by 0
func GetAddrFormat(addr string) string {
	return fmt.Sprintf("%021s", addr)
}

// GetAddrNextFormat returns next addr format for sort, padding left by 0
func GetAddrNextFormat(addr string) string {
	return fmt.Sprintf("%s%c", addr[:len(addr)-1], addr[len(addr)-1]+1)
}
