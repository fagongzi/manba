package model

import (
	"fmt"
	"testing"

	"github.com/fagongzi/gateway/pkg/util"
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
		req.SetBody(util.MustMarshal(data))
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

// func TestMatchesLt(t *testing.T) {
// 	r, err := newRoutingItem("$query_abc < 100")

// 	if err != nil {
// 		t.Error("parse error.")
// 	}

// 	req := &fasthttp.Request{}
// 	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=1")

// 	if !r.matches(req) {
// 		t.Error("matches op lt error")
// 	}
// }

// func TestMatchesLe(t *testing.T) {
// 	r, err := newRoutingItem("$query_abc <= 100")

// 	if err != nil {
// 		t.Error("parse error.")
// 	}

// 	req := &fasthttp.Request{}
// 	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=100")

// 	if !r.matches(req) {
// 		t.Error("matches op le error")
// 	}
// }

// func TestMatchesGt(t *testing.T) {
// 	r, err := newRoutingItem("$query_abc > 100")

// 	if err != nil {
// 		t.Error("parse error.")
// 	}

// 	req := &fasthttp.Request{}
// 	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=101")

// 	if !r.matches(req) {
// 		t.Error("matches op gt error")
// 	}
// }

// func TestMatchesGe(t *testing.T) {
// 	r, err := newRoutingItem("$query_abc >= 100")

// 	if err != nil {
// 		t.Error("parse error.")
// 	}

// 	req := &fasthttp.Request{}
// 	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=100")

// 	if !r.matches(req) {
// 		t.Error("matches op ge error")
// 	}
// }

// func TestMatchesIn(t *testing.T) {
// 	r, err := newRoutingItem("$query_abc in 100")

// 	if err != nil {
// 		t.Error("parse error.")
// 	}

// 	req := &fasthttp.Request{}
// 	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=11001")

// 	if !r.matches(req) {
// 		t.Error("matches op in error")
// 	}
// }

// func TestMatchesReg(t *testing.T) {
// 	r, err := newRoutingItem("$query_abc ~ ^1100")

// 	if err != nil {
// 		t.Error("parse error.")
// 	}

// 	req := &fasthttp.Request{}
// 	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=11001a")

// 	if !r.matches(req) {
// 		t.Error("matches op reg error")
// 	}
// }

// func TestMatchesRouting(t *testing.T) {
// 	data := `desc = "test";
// 	deadline = 100;
// 	rule = ["$query_abc == abc"];
// 	`

// 	r, err := NewRouting(data, "cluster", "/abc*")

// 	if err != nil {
// 		t.Error("parse error.")
// 	}

// 	req := &fasthttp.Request{}
// 	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=abc")

// 	if !r.Matches(req) {
// 		t.Error("matches routing error")
// 	}
// }

// func TestNotMatchesRouting(t *testing.T) {
// 	data := `desc = "test";
// 	deadline = 100;
// 	rule = ["$query_abc == 10"];
// 	`

// 	r, err := NewRouting(data, "cluster", "/abc*")

// 	if err != nil {
// 		t.Error("parse error.")
// 	}

// 	req := &fasthttp.Request{}
// 	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=20")

// 	if r.Matches(req) {
// 		t.Error("not matches routing error")
// 	}
// }
