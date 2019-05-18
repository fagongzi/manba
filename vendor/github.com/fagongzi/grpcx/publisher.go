package grpcx

import (
	"google.golang.org/grpc/naming"
)

// Publisher a service publish
type Publisher interface {
	Publish(service string, meta naming.Update) error
}
