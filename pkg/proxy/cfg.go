package proxy

import (
	"fmt"
	"strings"
	"time"
)

// Option proxy option
type Option struct {
	LimitCountDispatchWorker   uint64
	LimitCountCopyWorker       uint64
	LimitCountHeathCheckWorker int
	LimitCountConn             int
	LimitIntervalHeathCheck    time.Duration
	LimitDurationConnKeepalive time.Duration
	LimitDurationConnIdle      time.Duration
	LimitTimeoutWrite          time.Duration
	LimitTimeoutRead           time.Duration
	LimitBufferRead            int
	LimitBufferWrite           int
	LimitBytesBody             int
}

// Cfg proxy config
type Cfg struct {
	Addr      string
	AddrRPC   string
	AddrStore string
	AddrPPROF string
	Namespace string
	TTLProxy  int64
	Filers    []*FilterSpec

	Option *Option
}

// AddFilter add a filter
func (c *Cfg) AddFilter(filter *FilterSpec) {
	c.Filers = append(c.Filers, filter)
}

// FilterSpec filter spec
type FilterSpec struct {
	Name               string `json:"name"`
	External           bool   `json:"external,omitempty"`
	ExternalPluginFile string `json:"externalPluginFile,omitempty"`
	ExternalCfg        string `json:"externalCfg,omitempty"`
}

// ParseFilter returns a filter
func ParseFilter(filter string) (*FilterSpec, error) {
	specs := strings.Split(filter, ":")

	switch len(specs) {
	case 1:
		return &FilterSpec{Name: specs[0]}, nil
	case 2:
		return &FilterSpec{
			Name:               specs[0],
			External:           true,
			ExternalPluginFile: specs[1]}, nil
	case 3:
		return &FilterSpec{
			Name:               specs[0],
			External:           true,
			ExternalPluginFile: specs[1],
			ExternalCfg:        specs[2]}, nil
	default:
		return nil, fmt.Errorf("error format: %s", filter)
	}
}
