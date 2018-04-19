package goetty

import "time"

const (
	// DefaultSessionBucketSize default bucket size of session map
	DefaultSessionBucketSize = 64
	// DefaultReadBuf read buf size
	DefaultReadBuf = 256
	// DefaultWriteBuf write buf size
	DefaultWriteBuf = 256
)

var (
	defaultEncoder = NewEmptyEncoder()
	defaultDecoder = NewEmptyDecoder()
)

// ServerOption option of server side
type ServerOption func(*serverOptions)

type serverOptions struct {
	decoder                   Decoder
	encoder                   Encoder
	generator                 IDGenerator
	readBufSize, writeBufSize int
	sessionBucketSize         int
	useSyncProtocol           bool
	middlewares               []Middleware
}

func (opts *serverOptions) adjust() {
	if opts.sessionBucketSize == 0 {
		opts.sessionBucketSize = DefaultSessionBucketSize
	}

	if opts.readBufSize == 0 {
		opts.readBufSize = DefaultReadBuf
	}

	if opts.writeBufSize == 0 {
		opts.writeBufSize = DefaultWriteBuf
	}

	if opts.generator == nil {
		opts.generator = NewInt64IDGenerator()
	}

	if opts.encoder == nil {
		opts.encoder = defaultEncoder
	}

	if opts.decoder == nil {
		opts.decoder = defaultDecoder
	}
}

// WithServerDecoder option of server's decoder
func WithServerDecoder(decoder Decoder) ServerOption {
	return func(opts *serverOptions) {
		opts.decoder = decoder
	}
}

// WithServerEncoder option of server's encoder
func WithServerEncoder(encoder Encoder) ServerOption {
	return func(opts *serverOptions) {
		opts.encoder = encoder
	}
}

// WithServerReadBufSize option of server's read buf size
func WithServerReadBufSize(readBufSize int) ServerOption {
	return func(opts *serverOptions) {
		opts.readBufSize = readBufSize
	}
}

// WithServerWriteBufSize option of server's write buf size
func WithServerWriteBufSize(writeBufSize int) ServerOption {
	return func(opts *serverOptions) {
		opts.writeBufSize = writeBufSize
	}
}

// WithServerIDGenerator option of server's id generator
func WithServerIDGenerator(generator IDGenerator) ServerOption {
	return func(opts *serverOptions) {
		opts.generator = generator
	}
}

// WithServerMiddleware option of handle write timeout
func WithServerMiddleware(middlewares ...Middleware) ServerOption {
	return func(opts *serverOptions) {
		opts.middlewares = append(opts.middlewares, middlewares...)
	}
}

// ClientOption option of client side
type ClientOption func(*clientOptions)

type clientOptions struct {
	decoder                   Decoder
	encoder                   Encoder
	readBufSize, writeBufSize int
	connectTimeout            time.Duration
	writeTimeout              time.Duration
	writeTimeoutHandler       func(string, IOSession)
	timeWheel                 *TimeoutWheel
	middlewares               []Middleware
}

func (opts *clientOptions) adjust() {
	if opts.readBufSize == 0 {
		opts.readBufSize = DefaultReadBuf
	}

	if opts.writeBufSize == 0 {
		opts.writeBufSize = DefaultWriteBuf
	}

	if opts.encoder == nil {
		opts.encoder = defaultEncoder
	}

	if opts.decoder == nil {
		opts.decoder = defaultDecoder
	}
}

// WithClientDecoder option of client's decoder
func WithClientDecoder(decoder Decoder) ClientOption {
	return func(opts *clientOptions) {
		opts.decoder = decoder
	}
}

// WithClientEncoder option of client's encoder
func WithClientEncoder(encoder Encoder) ClientOption {
	return func(opts *clientOptions) {
		opts.encoder = encoder
	}
}

// WithClientReadBufSize option of client's read buf size
func WithClientReadBufSize(readBufSize int) ClientOption {
	return func(opts *clientOptions) {
		opts.readBufSize = readBufSize
	}
}

// WithClientWriteBufSize option of client's write buf size
func WithClientWriteBufSize(writeBufSize int) ClientOption {
	return func(opts *clientOptions) {
		opts.writeBufSize = writeBufSize
	}
}

// WithClientConnectTimeout option of timeout to connect to server
func WithClientConnectTimeout(timeout time.Duration) ClientOption {
	return func(opts *clientOptions) {
		opts.connectTimeout = timeout
	}
}

// WithClientWriteTimeoutHandler option of handle write timeout
func WithClientWriteTimeoutHandler(timeout time.Duration, handler func(string, IOSession), timeWheel *TimeoutWheel) ClientOption {
	return func(opts *clientOptions) {
		opts.writeTimeout = timeout
		opts.writeTimeoutHandler = handler
		opts.timeWheel = timeWheel
	}
}

// WithClientMiddleware option of handle write timeout
func WithClientMiddleware(middlewares ...Middleware) ClientOption {
	return func(opts *clientOptions) {
		opts.middlewares = append(opts.middlewares, middlewares...)
	}
}
