package model

import (
	"github.com/fagongzi/gateway/pkg/conf"
)

// ProxyInfo proxy info
type ProxyInfo struct {
	Conf *conf.Conf `json:"conf,omitempty"`
}
