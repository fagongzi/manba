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
	ClusterName string        `json:"clusterName, omitempty"`
	Rewrite     string        `json:"rewrite, omitempty"`
	AttrName    string        `json:"attrName, omitempty"`
	Validations []*Validation `json:"validations, omitempty"`
}

// Mock mock
type Mock struct {
	Value         string             `json:"value"`
	ContentType   string             `json:"contentType, omitempty"`
	Headers       []*MockHeader      `json:"headers, omitempty"`
	Cookies       []string           `json:"cookies, omitempty"`
	ParsedCookies []*fasthttp.Cookie `json:"-"`
}

// MockHeader header
type MockHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// API a api define
type API struct {
	Name    string         `json:"name, omitempty"`
	URL     string         `json:"url"`
	Method  string         `json:"method"`
	Mock    *Mock          `json:"mock, omitempty"`
	Nodes   []*Node        `json:"nodes"`
	Desc    string         `json:"desc, omitempty"`
	Pattern *regexp.Regexp `json:"-"`
}

// UnMarshalAPI unmarshal
func UnMarshalAPI(data []byte) *API {
	v := &API{}
	json.Unmarshal(data, v)

	if v.Mock != nil && v.Mock.Value == "" {
		v.Mock = nil
	}

	return v
}

// UnMarshalAPIFromReader unmarshal from reader
func UnMarshalAPIFromReader(r io.Reader) (*API, error) {
	v := &API{}

	decoder := json.NewDecoder(r)
	err := decoder.Decode(v)

	if v.Mock != nil && v.Mock.Value == "" {
		v.Mock = nil
	}

	return v, err
}

// NewAPI create a API
func NewAPI(url string, nodes []*Node) *API {
	return &API{
		URL:   url,
		Nodes: nodes,
	}
}

// Parse parse
func (a *API) Parse() {
	a.Pattern = regexp.MustCompile(a.URL)
	for _, n := range a.Nodes {
		if nil != n.Validations {
			for _, v := range n.Validations {
				v.ParseValidation()
			}
		}
	}

	if nil != a.Mock && nil != a.Mock.Cookies && len(a.Mock.Cookies) > 0 {
		a.Mock.ParsedCookies = make([]*fasthttp.Cookie, len(a.Mock.Cookies))
		for index, c := range a.Mock.Cookies {
			ck := &fasthttp.Cookie{}
			ck.Parse(c)
			a.Mock.ParsedCookies[index] = ck
		}
	}
}

// RenderMock dender mock response
func (a *API) RenderMock(ctx *fasthttp.RequestCtx) {
	if a.Mock == nil {
		return
	}

	ctx.Response.Header.SetContentType(a.Mock.ContentType)

	if a.Mock.Headers != nil && len(a.Mock.Headers) > 0 {
		for _, header := range a.Mock.Headers {
			ctx.Response.Header.Add(header.Name, header.Value)
		}
	}

	if a.Mock.ParsedCookies != nil && len(a.Mock.ParsedCookies) > 0 {
		for _, ck := range a.Mock.ParsedCookies {
			ctx.Response.Header.SetCookie(ck)
		}
	}

	ctx.WriteString(a.Mock.Value)
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
