package model

import (
	"encoding/json"
	"io"
	"net/url"
)

type Node struct {
	ClusterName string `json:"clusterName,omitempty"`
	Url         string `json:"url,omitempty"`
	AttrName    string `json:"attrName,omitempty"`
}

type Aggregation struct {
	Url   string  `json:"url"`
	Nodes []*Node `json:"nodes"`
}

func UnMarshalAggregation(data []byte) *Aggregation {
	v := &Aggregation{}
	json.Unmarshal(data, v)
	v.Url, _ = url.QueryUnescape(v.Url)
	return v
}

func UnMarshalAggregationFromReader(r io.Reader) (*Aggregation, error) {
	v := &Aggregation{}

	decoder := json.NewDecoder(r)
	err := decoder.Decode(v)

	return v, err
}

func NewAggregation(url string, nodes []*Node) *Aggregation {
	return &Aggregation{
		Url:   url,
		Nodes: nodes,
	}
}

func (self *Aggregation) Marshal() []byte {
	v, _ := json.Marshal(self)
	return v
}

func (self *Aggregation) updateFrom(ang *Aggregation) {
	self.Url = ang.Url
	self.Nodes = ang.Nodes
}
