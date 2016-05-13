package model

import (
	"encoding/json"
	"fmt"
	"io"
)

type Bind struct {
	ClusterName string `json:"clusterName,omitempty"`
	ServerAddr  string `json:"serverAddr,omitempty"`
}

func UnMarshalBindFromReader(r io.Reader) (*Bind, error) {
	v := &Bind{}

	decoder := json.NewDecoder(r)
	err := decoder.Decode(v)

	if nil != err {
		return nil, err
	}

	return v, nil
}

func (self *Bind) ToString() string {
	return fmt.Sprintf("%s-%s", self.ServerAddr, self.ClusterName)
}

func (self *Bind) Marshal() []byte {
	v, _ := json.Marshal(self)
	return v
}
