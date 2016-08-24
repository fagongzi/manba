package util

import (
	"github.com/fgrid/uuid"
)

// UUID return uuid string
func UUID() string {
	return uuid.NewV4().String()
}
