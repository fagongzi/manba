package model

import (
	"fmt"
	"testing"

	"github.com/fagongzi/util/json"
	"github.com/valyala/fasthttp"
)

func TestMatch(t *testing.T) {
	testRegCase(QueryString, t)
	testRegCase(FormData, t)
	testRegCase(JSONBody, t)
	testRegCase(RequestHeader, t)
	testRegCase(Cookie, t)

	testNumberCase(QueryString, t)
	testNumberCase(FormData, t)
	testNumberCase(JSONBody, t)
	testNumberCase(RequestHeader, t)
	testNumberCase(Cookie, t)

	testStringCase(QueryString, t)
	testStringCase(FormData, t)
	testStringCase(JSONBody, t)
	testStringCase(RequestHeader, t)
	testStringCase(Cookie, t)
}

func testRegCase(src Source, t *testing.T) {
	attr := "name"
	value := "^[0-1]+$"

	c := &Condition{
		Attr: &Attr{
			Name:   attr,
			Source: src,
		},
		Op:    OpMatch,
		Value: value,
	}

	assertTrue(c.Match(getTestReq(attr, "0", src)), fmt.Sprintf("Match %v failed", c.Op), t)
	assertTrue(c.Match(getTestReq(attr, "00", src)), fmt.Sprintf("Match %v failed", c.Op), t)
	assertFalse(c.Match(getTestReq(attr, "00a", src)), fmt.Sprintf("Match %v failed", c.Op), t)
	assertFalse(c.Match(getTestReq(attr, "aaa", src)), fmt.Sprintf("Match %v failed", c.Op), t)
}

func testStringCase(src Source, t *testing.T) {
	attr := "name"
	value := "hello"

	c := &Condition{
		Attr: &Attr{
			Name:   attr,
			Source: src,
		},
		Op:    OpEQ,
		Value: value,
	}

	assertTrue(c.Match(getTestReq(attr, value, src)), fmt.Sprintf("Match %v failed", c.Op), t)
	assertFalse(c.Match(getTestReq(attr, fmt.Sprintf("%s0", value), src)), fmt.Sprintf("Match %v failed", c.Op), t)

	c.Op = OpIn
	assertTrue(c.Match(getTestReq(attr, value, src)), fmt.Sprintf("Match %v failed", c.Op), t)
	assertTrue(c.Match(getTestReq(attr, value[1:], src)), fmt.Sprintf("Match %v failed", c.Op), t)
	assertFalse(c.Match(getTestReq(attr, fmt.Sprintf("%s0", value), src)), fmt.Sprintf("Match %v failed", c.Op), t)
}

func testNumberCase(src Source, t *testing.T) {
	attr := "age"
	value := 10

	c := &Condition{
		Attr: &Attr{
			Name:   attr,
			Source: src,
		},
		Op:    OpLT,
		Value: fmt.Sprintf("%d", value),
	}

	assertTrue(c.Match(getTestReq(attr, fmt.Sprintf("%d", value-1), src)), fmt.Sprintf("Match %v failed", c.Op), t)
	assertFalse(c.Match(getTestReq(attr, fmt.Sprintf("%d", value), src)), fmt.Sprintf("Match %v failed", c.Op), t)
	assertFalse(c.Match(getTestReq(attr, fmt.Sprintf("%d", value+1), src)), fmt.Sprintf("Match %v failed", c.Op), t)

	c.Op = OpLE
	assertTrue(c.Match(getTestReq(attr, fmt.Sprintf("%d", value-1), src)), fmt.Sprintf("Match %v failed", c.Op), t)
	assertTrue(c.Match(getTestReq(attr, fmt.Sprintf("%d", value), src)), fmt.Sprintf("Match %v failed", c.Op), t)
	assertFalse(c.Match(getTestReq(attr, fmt.Sprintf("%d", value+1), src)), fmt.Sprintf("Match %v failed", c.Op), t)

	c.Op = OpGT
	assertFalse(c.Match(getTestReq(attr, fmt.Sprintf("%d", value-1), src)), fmt.Sprintf("Match %v failed", c.Op), t)
	assertFalse(c.Match(getTestReq(attr, fmt.Sprintf("%d", value), src)), fmt.Sprintf("Match %v failed", c.Op), t)
	assertTrue(c.Match(getTestReq(attr, fmt.Sprintf("%d", value+1), src)), fmt.Sprintf("Match %v failed", c.Op), t)

	c.Op = OpGE
	assertFalse(c.Match(getTestReq(attr, fmt.Sprintf("%d", value-1), src)), fmt.Sprintf("Match %v failed", c.Op), t)
	assertTrue(c.Match(getTestReq(attr, fmt.Sprintf("%d", value), src)), fmt.Sprintf("Match %v failed", c.Op), t)
	assertTrue(c.Match(getTestReq(attr, fmt.Sprintf("%d", value+1), src)), fmt.Sprintf("Match %v failed", c.Op), t)

}

func getTestReq(name string, value string, src Source) *fasthttp.Request {
	req := &fasthttp.Request{}
	switch src {
	case QueryString:
		req.SetRequestURI(fmt.Sprintf("http://127.0.0.1:8080/path?%s=%s", name, value))
	case FormData:
		req.PostArgs().Add(name, value)
	case JSONBody:
		data := make(map[string]interface{})
		data[name] = value
		req.SetBody(json.MustMarshal(data))
	case RequestHeader:
		req.Header.Add(name, value)
	case Cookie:
		req.Header.SetCookie(name, value)
	}

	return req
}

func assertTrue(value bool, msg string, t *testing.T) {
	if !value {
		t.Error(msg)
	}
}

func assertFalse(value bool, msg string, t *testing.T) {
	assertTrue(!value, msg, t)
}
