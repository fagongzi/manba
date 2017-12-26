package grpcx

import (
	"github.com/labstack/echo"
)

//ServiceOption service options
type ServiceOption func(*serviceOptions)

type serviceOptions struct {
	httpEntrypoints []*httpEntrypoint
}

// Service is a service define
type Service struct {
	Name     string
	Metadata interface{}
	opts     *serviceOptions
}

// NewService returns a Service
func NewService(name string, metadata interface{}, opts ...ServiceOption) Service {
	sopts := &serviceOptions{}
	for _, opt := range opts {
		opt(sopts)
	}

	service := Service{
		Name:     name,
		Metadata: metadata,
		opts:     sopts,
	}

	return service
}

// WithAddGetHTTPEntrypoint add a http get service metadata
func WithAddGetHTTPEntrypoint(path string, reqFactory func() interface{}, invoker func(interface{}, echo.Context) (interface{}, error)) ServiceOption {
	return withAddHTTPEntrypoint(path, echo.GET, reqFactory, invoker)
}

// WithAddPutHTTPEntrypoint add a http put service metadata
func WithAddPutHTTPEntrypoint(path string, reqFactory func() interface{}, invoker func(interface{}, echo.Context) (interface{}, error)) ServiceOption {
	return withAddHTTPEntrypoint(path, echo.PUT, reqFactory, invoker)
}

// WithAddPostHTTPEntrypoint add a http post service metadata
func WithAddPostHTTPEntrypoint(path string, reqFactory func() interface{}, invoker func(interface{}, echo.Context) (interface{}, error)) ServiceOption {
	return withAddHTTPEntrypoint(path, echo.POST, reqFactory, invoker)
}

// WithAddDeleteHTTPEntrypoint add a http post service metadata
func WithAddDeleteHTTPEntrypoint(path string, reqFactory func() interface{}, invoker func(interface{}, echo.Context) (interface{}, error)) ServiceOption {
	return withAddHTTPEntrypoint(path, echo.DELETE, reqFactory, invoker)
}

func withAddHTTPEntrypoint(path, method string, reqFactory func() interface{}, invoker func(interface{}, echo.Context) (interface{}, error)) ServiceOption {
	return func(opt *serviceOptions) {
		opt.httpEntrypoints = append(opt.httpEntrypoints, &httpEntrypoint{
			path:       path,
			method:     method,
			reqFactory: reqFactory,
			invoker:    invoker,
		})
	}
}
