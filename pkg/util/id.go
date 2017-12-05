package util

import (
	"strings"

	"github.com/fagongzi/goetty"
)

var (
	replacer = strings.NewReplacer("-", "")
)

// NewID returns a uuid string id
func NewID() string {
	id := goetty.NewKey()
	return replacer.Replace(id)
}
