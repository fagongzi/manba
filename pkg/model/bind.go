package model

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// Bind a bind server and cluster
type Bind struct {
	ClusterName string `json:"clusterName,omitempty"`
	ServerAddr  string `json:"serverAddr,omitempty"`
}

// UnMarshalBindFromReader unmarshal
func UnMarshalBindFromReader(r io.Reader) (*Bind, error) {
	v := &Bind{}

	decoder := json.NewDecoder(r)
	err := decoder.Decode(v)

	if nil != err {
		return nil, err
	}

	return v, nil
}

// ToString return a desc string
func (b *Bind) ToString() string {
	return fmt.Sprintf("%s-%s", b.ServerAddr, b.ClusterName)
}

// UnMarshalBindFromString unmarshal
func UnMarshalBindFromString(key string) *Bind {
	info := strings.Split(key, "-")
	return &Bind{
		ServerAddr:  info[0],
		ClusterName: info[1],
	}
}

// Marshal marshal
func (b *Bind) Marshal() []byte {
	v, _ := json.Marshal(b)
	return v
}

// UnMarshalBind unmarshal
func UnMarshalBind(data []byte) *Bind {
	v := &Bind{}
	err := json.Unmarshal(data, v)

	if err != nil {
		return v
	}

	return v
}
