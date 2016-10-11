package goetty

import (
	uuid "github.com/satori/go.uuid"
)

func NewKey() string {
	return NewV4UUID()
}

func NewV1UUID() string {
	return uuid.NewV1().String()
}

func NewV4UUID() string {
	return uuid.NewV4().String()
}

func NewV4Bytes() []byte {
	return uuid.NewV4().Bytes()
}

func NewV1Bytes() []byte {
	return uuid.NewV1().Bytes()
}
