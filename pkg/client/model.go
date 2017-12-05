package client

import (
	"github.com/fagongzi/gateway/pkg/lb"
	"github.com/fagongzi/gateway/pkg/model"
)

// LB load balance
type LB string

const (
	// LBRR ROUNDROBIN
	LBRR = LB(lb.ROUNDROBIN)
)

// NewCluster returns a cluster
func NewCluster(name string, lb LB) *model.Cluster {
	return &model.Cluster{
		Name:   name,
		LbName: string(lb),
	}
}

// BuildOption build option
type BuildOption func(interface{})

// NewServer returns a server
func NewServer(addr string, options ...BuildOption) *model.Server {
	svr := &model.Server{}
	svr.Addr = addr

	for _, opt := range options {
		opt(svr)
	}

	return svr
}

// // CheckPath begin with / checkpath, expect return CheckResponsedBody.
// CheckPath string `json:"checkPath,omitempty"`
// // CheckResponsedBody check url responsed http body, if not set, not check body
// CheckResponsedBody string `json:"checkResponsedBody"`
// // CheckDuration check interval, unit second
// CheckDuration int `json:"checkDuration,omitempty"`
// // CheckTimeout timeout to check server
// CheckTimeout int `json:"checkTimeout,omitempty"`

// func HealthCheckOption(checkURL string, checkInterval int) BuildOption {

// }
