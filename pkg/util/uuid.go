package util

import (
	"github.com/fgrid/uuid"
)

func UUID() string {
	return uuid.NewV4().String()
}
