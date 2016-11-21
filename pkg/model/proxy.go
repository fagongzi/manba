package model

import (
	"encoding/json"

	"github.com/fagongzi/gateway/pkg/conf"
)

// ProxyInfo proxy info
type ProxyInfo struct {
	Conf *conf.Conf `json:"conf,omitempty"`
}

// UnMarshalProxyInfo unmarshal
func UnMarshalProxyInfo(data []byte) *ProxyInfo {
	v := &ProxyInfo{}
	json.Unmarshal(data, v)

	return v
}

// Marshal marshal
func (p *ProxyInfo) Marshal() string {
	v, _ := json.Marshal(p)
	return string(v)
}
