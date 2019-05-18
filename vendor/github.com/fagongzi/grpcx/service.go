package grpcx

//ServiceOption service options
type ServiceOption func(*serviceOptions)

type serviceOptions struct {
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
