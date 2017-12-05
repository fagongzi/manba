package uuid

import (
	"strings"
)

var (
	replacer = strings.NewReplacer("-", "")
)

// NewID returns a UUID V4 string
func NewID() string {
	return replacer.Replace(NewV4().String())
}
