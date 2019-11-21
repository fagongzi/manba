package expr

import (
	"testing"

	"github.com/valyala/fasthttp"
)

func TestParse(t *testing.T) {
	value := []byte("abc$(")
	_, err := Parse(value)
	if err == nil {
		t.Errorf("expect syntax error: %+v", err)
	}

	value = []byte("abc$(abc)")
	_, err = Parse(value)
	if err == nil {
		t.Errorf("expect syntax error: %+v", err)
	}

	value = []byte("abc$(origin.)")
	_, err = Parse(value)
	if err == nil {
		t.Errorf("expect syntax error: %+v", err)
	}

	value = []byte("abc$(origin.path)(")
	_, err = Parse(value)
	if err == nil {
		t.Errorf("expect syntax error: %+v", err)
	}

	value = []byte("abc$(origin.path))")
	_, err = Parse(value)
	if err == nil {
		t.Errorf("expect syntax error: %+v", err)
	}

	value = []byte("abc")
	exprs, err := Parse(value)
	if err != nil {
		t.Errorf("parse const error: %+v", err)
	}
	if len(exprs) != 1 || exprs[0].Name() != "const-expr" {
		t.Errorf("parse const error")
	}

	value = []byte("$(origin.path)")
	exprs, err = Parse(value)
	if err != nil {
		t.Errorf("parse origin-path error: %+v", err)
	}
	if len(exprs) != 1 || exprs[0].Name() != "origin-path-expr" {
		t.Errorf("parse origin-path error, %+v", exprs)
	}

	value = []byte("$(origin.query)")
	exprs, err = Parse(value)
	if err != nil {
		t.Errorf("parse origin-query error: %+v", err)
	}
	if len(exprs) != 1 || exprs[0].Name() != "origin-query-expr" {
		t.Errorf("parse origin-query error, %+v", exprs)
	}

	value = []byte("$(origin.query.abc)")
	exprs, err = Parse(value)
	if err != nil {
		t.Errorf("parse origin-query-param error: %+v", err)
	}
	if len(exprs) != 1 || exprs[0].Name() != "origin-query-param-expr" {
		t.Errorf("parse origin-query-param error, %+v", exprs)
	}

	value = []byte("$(origin.cookie.abc)")
	exprs, err = Parse(value)
	if err != nil {
		t.Errorf("parse origin-cookie error: %+v", err)
	}
	if len(exprs) != 1 || exprs[0].Name() != "origin-cookie-expr" {
		t.Errorf("parse origin-cookie error, %+v", exprs)
	}

	value = []byte("$(origin.header.abc)")
	exprs, err = Parse(value)
	if err != nil {
		t.Errorf("parse origin-header error: %+v", err)
	}
	if len(exprs) != 1 || exprs[0].Name() != "origin-header-expr" {
		t.Errorf("parse origin-header error, %+v", exprs)
	}

	value = []byte("$(origin.body.abc.abc.abc)")
	exprs, err = Parse(value)
	if err != nil {
		t.Errorf("parse origin-body error: %+v", err)
	}
	if len(exprs) != 1 || exprs[0].Name() != "origin-body-expr" {
		t.Errorf("parse origin-body error, %+v", exprs)
	}

	value = []byte("$(depend.abc.abc.abc)")
	exprs, err = Parse(value)
	if err != nil {
		t.Errorf("parse depend error: %+v", err)
	}
	if len(exprs) != 1 || exprs[0].Name() != "depend-expr" {
		t.Errorf("parse depend error, %+v", exprs)
	}

	value = []byte("$(param.abc)")
	exprs, err = Parse(value)
	if err != nil {
		t.Errorf("parse param error: %+v", err)
	}
	if len(exprs) != 1 || exprs[0].Name() != "param-expr" {
		t.Errorf("parse param error, %+v", exprs)
	}

	value = []byte("/$(origin.path)?id=$(param.abc)&value=$(origin.body.abc.abc)&value2=$(depend.abc.abc.abc)&value3=$(origin.header.abc)&value4=4")
	exprs, err = Parse(value)
	if err != nil {
		t.Errorf("parse param error: %+v", err)
	}
	if len(exprs) != 11 {
		t.Errorf("expect 11 expers but %d", len(exprs))
	}
}

func TestExec(t *testing.T) {
	exprs, err := Parse([]byte("$(origin.query.names)"))
	if err != nil {
		t.Errorf("expect syntax error: %+v", err)
	}

	req := fasthttp.AcquireRequest()
	req.SetRequestURI("http://127.0.0.1/path?names=abc&names=abc2")
	value := Exec(&Ctx{
		Origin: req,
	}, exprs...)

	if string(value) != "abc" {
		t.Errorf("expect but %s", value)
	}
}
