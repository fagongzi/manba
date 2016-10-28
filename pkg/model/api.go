package model

import (
	"encoding/json"
	"io"
	"regexp"

	"strings"

	"github.com/valyala/fasthttp"
)

// Node api dispatch node
type Node struct {
	ClusterName string `json:"clusterName, omitempty"`
	Rewrite     string `json:"rewrite, omitempty"`
	AttrName    string `json:"attrName, omitempty"`
}

// API a api define
type API struct {
	URL     string         `json:"url"`
	Method  string         `json:"method"`
	Alias   string         `json:"alias, omitempty"`
	Nodes   []*Node        `json:"nodes"`
	Desc    string         `json:"desc, omitempty"`
	Pattern *regexp.Regexp `json:"-"`
}

// UnMarshalAPI unmarshal
func UnMarshalAPI(data []byte) *API {
	v := &API{}
	json.Unmarshal(data, v)
	return v
}

// UnMarshalAPIFromReader unmarshal from reader
func UnMarshalAPIFromReader(r io.Reader) (*API, error) {
	v := &API{}

	decoder := json.NewDecoder(r)
	err := decoder.Decode(v)

	return v, err
}

// NewAPI create a API
func NewAPI(url string, nodes []*Node) *API {
	return &API{
		URL:   url,
		Nodes: nodes,
	}
}

// Marshal marshal
func (a *API) Marshal() []byte {
	v, _ := json.Marshal(a)
	return v
}

func (a *API) getNodeURL(req *fasthttp.Request, node *Node) string {
	if node.Rewrite == "" {
		return ""
	}

	return a.Pattern.ReplaceAllString(string(req.URI().RequestURI()), node.Rewrite)
}

func (a *API) matches(req *fasthttp.Request) bool {
	return (a.Method == "*" || strings.ToUpper(string(req.Header.Method())) == a.Method) && a.Pattern.Match(req.URI().RequestURI())
}

func (a *API) updateFrom(api *API) {
	a.URL = api.URL
	a.Nodes = api.Nodes
}
