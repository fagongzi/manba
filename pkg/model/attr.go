package model

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/valyala/fasthttp"
)

// Source where attr get from
type Source int

const (
	// QueryString from query
	QueryString = Source(0)
	// FormData from form data
	FormData = Source(1)
	// JSONBody from json body
	JSONBody = Source(2)
	// RequestHeader from request header
	RequestHeader = Source(3)
	// Cookie from cookie
	Cookie = Source(4)
)

// Attr attr
type Attr struct {
	Name   string `json:"name"`
	Source Source `json:"source"`
}

// Value reutrn value form source
func (attr *Attr) Value(req *fasthttp.Request) string {
	switch attr.Source {
	case QueryString:
		return attr.getQueryValue(req)
	case FormData:
		return attr.getFormValue(req)
	case JSONBody:
		return attr.getJSONBodyValue(req)
	case RequestHeader:
		return attr.getHeaderValue(req)
	case Cookie:
		return attr.getCookieValue(req)
	default:
		return ""
	}
}

func (attr *Attr) getCookieValue(req *fasthttp.Request) string {
	return string(req.Header.Cookie(attr.Name))
}

func (attr *Attr) getHeaderValue(req *fasthttp.Request) string {
	return string(req.Header.Peek(attr.Name))
}

func (attr *Attr) getQueryValue(req *fasthttp.Request) string {
	v, _ := url.QueryUnescape(string(req.URI().QueryArgs().Peek(attr.Name)))
	return v
}

func (attr *Attr) getFormValue(req *fasthttp.Request) string {
	return string(req.PostArgs().Peek(attr.Name))
}

func (attr *Attr) getJSONBodyValue(req *fasthttp.Request) string {
	data := make(map[string]interface{})
	err := json.Unmarshal(req.Body(), &data)
	if err != nil {
		return ""
	}

	var value interface{}
	for _, name := range strings.Split(attr.Name, ".") {
		if value == nil {
			value = data[name]
			continue
		}

		if m, ok := value.(map[string]interface{}); ok {
			value = m[name]
		} else {
			return ""
		}
	}

	if ret, ok := value.(string); ok {
		return ret
	}

	ret, _ := json.Marshal(&value)
	return string(ret)
}
