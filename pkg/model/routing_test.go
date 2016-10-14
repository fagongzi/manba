package model

import (
	"testing"

	"github.com/valyala/fasthttp"
)

func TestParse(t *testing.T) {
	r, err := newRoutingItem("$header_abc_!= == abc== asd ")

	if err != nil {
		t.Error("parse error.")
	}

	if r.targetValue != "abc== asd" {
		t.Error("value parse error.")
	}

	if r.attrName != "abc_!=" {
		t.Error("attr parse error.")
	}
}

func TestParseError(t *testing.T) {
	_, err := newRoutingItem("$header_abc != abc")

	if err == nil {
		t.Error("parse error.")
	}
}

func TestParseHeader(t *testing.T) {
	r, err := newRoutingItem("$header_abc == abc")

	if err != nil {
		t.Error("parse error.")
	}

	req := &fasthttp.Request{}
	req.SetRequestURI("/abc")
	req.Header.Add("abc", "abc")

	if r.sourceValueFun(req) != "abc" {
		t.Error("parse header error")
	}
}

func TestParseCookie(t *testing.T) {
	r, err := newRoutingItem("$cookie_abc == abc")

	if err != nil {
		t.Error("parse error.")
	}

	req := &fasthttp.Request{}
	req.SetRequestURI("/abc")
	req.Header.Add("cookie", "abc=abc")

	if r.sourceValueFun(req) != "abc" {
		t.Error("parse cookie error")
	}
}

func TestParseQuery(t *testing.T) {
	r, err := newRoutingItem("$query_abc == abc")

	if err != nil {
		t.Error("parse error.")
	}

	req := &fasthttp.Request{}
	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=abc")

	if r.sourceValueFun(req) != "abc" {
		t.Error("parse cookie error")
	}
}

func TestMatchesEq(t *testing.T) {
	r, err := newRoutingItem("$query_abc == abc")

	if err != nil {
		t.Error("parse error.")
	}

	req := &fasthttp.Request{}
	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=abc")

	if !r.matches(req) {
		t.Error("matches op eq error")
	}
}

func TestMatchesLt(t *testing.T) {
	r, err := newRoutingItem("$query_abc < 100")

	if err != nil {
		t.Error("parse error.")
	}

	req := &fasthttp.Request{}
	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=1")

	if !r.matches(req) {
		t.Error("matches op lt error")
	}
}

func TestMatchesLe(t *testing.T) {
	r, err := newRoutingItem("$query_abc <= 100")

	if err != nil {
		t.Error("parse error.")
	}

	req := &fasthttp.Request{}
	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=100")

	if !r.matches(req) {
		t.Error("matches op le error")
	}
}

func TestMatchesGt(t *testing.T) {
	r, err := newRoutingItem("$query_abc > 100")

	if err != nil {
		t.Error("parse error.")
	}

	req := &fasthttp.Request{}
	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=101")

	if !r.matches(req) {
		t.Error("matches op gt error")
	}
}

func TestMatchesGe(t *testing.T) {
	r, err := newRoutingItem("$query_abc >= 100")

	if err != nil {
		t.Error("parse error.")
	}

	req := &fasthttp.Request{}
	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=100")

	if !r.matches(req) {
		t.Error("matches op ge error")
	}
}

func TestMatchesIn(t *testing.T) {
	r, err := newRoutingItem("$query_abc in 100")

	if err != nil {
		t.Error("parse error.")
	}

	req := &fasthttp.Request{}
	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=11001")

	if !r.matches(req) {
		t.Error("matches op in error")
	}
}

func TestMatchesReg(t *testing.T) {
	r, err := newRoutingItem("$query_abc ~ ^1100")

	if err != nil {
		t.Error("parse error.")
	}

	req := &fasthttp.Request{}
	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=11001a")

	if !r.matches(req) {
		t.Error("matches op reg error")
	}
}

func TestMatchesRouting(t *testing.T) {
	data := `desc = "test";
	deadline = 100;
	rule = ["$query_abc == abc"];
	`

	r, err := NewRouting(data, "cluster", "/abc*")

	if err != nil {
		t.Error("parse error.")
	}

	req := &fasthttp.Request{}
	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=abc")

	if !r.Matches(req) {
		t.Error("matches routing error")
	}
}

func TestNotMatchesRouting(t *testing.T) {
	data := `desc = "test";
	deadline = 100;
	rule = ["$query_abc == 10"];
	`

	r, err := NewRouting(data, "cluster", "/abc*")

	if err != nil {
		t.Error("parse error.")
	}

	req := &fasthttp.Request{}
	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=20")

	if r.Matches(req) {
		t.Error("not matches routing error")
	}
}

func TestMatchesRoutingAndLogic(t *testing.T) {
	data := `desc = "test";
	deadline = 100;
	rule = ["$query_abc == 10", "$query_123 == 20"];
	`
	r, err := NewRouting(data, "cluster", "/abc*")

	if err != nil {
		t.Error("parse error.")
	}

	req := &fasthttp.Request{}
	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=10&123=20")

	if !r.Matches(req) {
		t.Error("matches and error")
	}
}

func TestNotMatchesRoutingAndLogic(t *testing.T) {
	data := `desc = "test";
	deadline = 100;
	rule = ["$query_abc == 10","$query_123 == 20"];
	`

	r, err := NewRouting(data, "cluster", "/abc*")

	if err != nil {
		t.Error("parse error.")
	}

	req := &fasthttp.Request{}
	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=10&123=30")

	if r.Matches(req) {
		t.Error("matches and error")
	}
}

func TestMatchesRoutingAllLogic(t *testing.T) {
	data := `desc = "test";
	deadline = 100;
	rule = ["$query_abc == 10","$query_123 == 20"];
	or = ["$query_or1 == 30", "$query_or2 == 40"];
	`

	r, err := NewRouting(data, "cluster", "/abc*")

	if err != nil {
		t.Error("parse error.")
	}

	req := &fasthttp.Request{}
	req.SetRequestURI("http://127.0.0.1:8080/abc?abc=10&123=10&or2=40")

	if !r.Matches(req) {
		t.Error("matches and error")
	}
}
