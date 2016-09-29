package conf

import "time"

// Conf config struct
type Conf struct {
	LogLevel string `json:"-"`

	Addr    string `json:"addr"`
	MgrAddr string `json:"mgrAddr"`

	EtcdAddrs  []string `json:"etcdAddrs"`
	EtcdPrefix string   `json:"etcdPrefix"`

	Filers []string `json:"filers"`

	// Maximum number of connections which may be established to server
	MaxConns int `json:"maxConns"`
	// MaxConnDuration Keep-alive connections are closed after this duration.
	MaxConnDuration time.Duration `json:"maxConnDuration"`
	// MaxIdleConnDuration Idle keep-alive connections are closed after this duration.
	MaxIdleConnDuration time.Duration `json:"maxIdleConnDuration"`
	// ReadBufferSize Per-connection buffer size for responses' reading.
	ReadBufferSize int `json:"readBufferSize"`
	// WriteBufferSize Per-connection buffer size for requests' writing.
	WriteBufferSize int `json:"writeBufferSize"`
	// ReadTimeout Maximum duration for full response reading (including body).
	ReadTimeout time.Duration `json:"readTimeout"`
	// WriteTimeout Maximum duration for full request writing (including body).
	WriteTimeout time.Duration `json:"writeTimeout"`
	// MaxResponseBodySize Maximum response body size.
	MaxResponseBodySize int `json:"maxResponseBodySize"`

	// EnablePPROF enable pprof
	EnablePPROF bool `json:"enablePPROF"`
	// PPROFAddr pprof addr
	PPROFAddr string `json:"pprofAddr,omitempty"`
}
