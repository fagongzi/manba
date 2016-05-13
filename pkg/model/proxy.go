package model

import (
	"encoding/json"
	"github.com/fagongzi/gateway/conf"
)

type ProxyInfo struct {
	Conf *conf.Conf `json:"conf,omitempty"`
}

func UnMarshalProxyInfo(data []byte) *ProxyInfo {
	v := &ProxyInfo{}
	json.Unmarshal(data, v)

	return v
}

func (self *ProxyInfo) Marshal() string {
	v, _ := json.Marshal(self)
	return string(v)
}
