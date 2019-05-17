package plugin

import (
	"testing"
	"time"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/gateway/pkg/util"
	"github.com/valyala/fasthttp"
)

type testCtx struct {
	req     *fasthttp.RequestCtx
	forward *fasthttp.Request
	res     *fasthttp.Response
	attrs   map[string]interface{}
}

func newTestCtx() filter.Context {
	return &testCtx{
		req: &fasthttp.RequestCtx{
			Request: *fasthttp.AcquireRequest(),
		},
		forward: fasthttp.AcquireRequest(),
		res:     fasthttp.AcquireResponse(),
		attrs:   make(map[string]interface{}),
	}
}

func (c *testCtx) StartAt() time.Time {
	return time.Now()
}

func (c *testCtx) EndAt() time.Time {
	return time.Now()
}

func (c *testCtx) OriginRequest() *fasthttp.RequestCtx {
	return c.req
}

func (c *testCtx) ForwardRequest() *fasthttp.Request {
	return c.forward
}

func (c *testCtx) Response() *fasthttp.Response {
	return c.res
}

func (c *testCtx) API() *metapb.API {
	return nil
}

func (c *testCtx) DispatchNode() *metapb.DispatchNode {
	return nil
}

func (c *testCtx) Server() *metapb.Server {
	return nil
}

func (c *testCtx) Analysis() *util.Analysis {
	return nil
}

func (c *testCtx) SetAttr(key string, value interface{}) {
	c.attrs[key] = value
}

func (c *testCtx) GetAttr(key string) interface{} {
	return c.attrs[key]
}

func TestNew(t *testing.T) {
	script := `
		function NewPlugin() {
			return {
				"pre": function(c) {
					return {
						"code": 200
					}
				},
				"post": function(c) {
					return {
						"code": 200
					}
				},
				"postErr": function(c) {
					
				}
			}
		}
	`
	meta := &metapb.Plugin{
		ID:      1,
		Name:    "test",
		Author:  "fagongzi",
		Email:   "zhangxu19830126@gmail.com",
		Version: 1,
		Type:    metapb.JavaScript,
		Content: []byte(script),
	}

	_, err := NewRuntime(meta)
	if err != nil {
		t.Errorf("create runtime plugin failed with %+v", err)
	}
}

func TestPre(t *testing.T) {
	c := newTestCtx()
	req := &c.OriginRequest().Request
	req.SetRequestURI("http://127.0.0.1/path?name=abc")
	req.Header.Add("x-error", "test failed11111111")
	req.Header.Add("x-removed", "test failed11111111")

	script := `
		var JSON = require("json");
		var REDIS = require("redis");
		var HTTP = require("http");
		var LOG = require("log");

		function NewPlugin(cfg) {
			return {
				"cfg": JSON.Parse(cfg),
				"pre": function(c) {
					c.SetAttr("x-query-name", c.OriginRequest().Query("name"))
					c.SetAttr("x-pre", "call-pre")
					c.SetAttr("x-cfg-ip", this.cfg.ip)
					c.OriginRequest().SetBody(JSON.Stringify({
						"name":"zhangsan"
					}))
					c.OriginRequest().RemoveHeader("x-removed")

					return {
						"code": 503,
						"error": c.OriginRequest().Header("x-error"),
					}
				}
			}
		}
	`
	meta := &metapb.Plugin{
		ID:      1,
		Name:    "test",
		Author:  "fagongzi",
		Email:   "zhangxu19830126@gmail.com",
		Version: 1,
		Type:    metapb.JavaScript,
		Content: []byte(script),
		Cfg:     []byte(`{"ip":"127.0.0.1"}`),
	}

	rt, err := NewRuntime(meta)
	if err != nil {
		t.Errorf("create runtime plugin failed with %+v", err)
	}

	code, err := rt.Pre(&Ctx{delegate: c})
	if err == nil || err.Error() != "test failed11111111" {
		t.Errorf("expect error for pre: expect x-error but %+v", err)
	}

	if value, ok := c.GetAttr("x-pre").(string); !ok || value != "call-pre" {
		t.Errorf(" expect attr x-pre but %+v", value)
	}

	if len(req.Header.Peek("x-removed")) > 0 {
		t.Errorf("header x-removed expect removed")
	}

	if value, ok := c.GetAttr("x-query-name").(string); !ok || value != "abc" {
		t.Errorf(" expect attr x-query-name but %+v", value)
	}

	if value, ok := c.GetAttr("x-cfg-ip").(string); !ok || value != "127.0.0.1" {
		t.Errorf(" expect attr x-cfg-ip but %+v", value)
	}

	if string(req.Body()) != `{"name":"zhangsan"}` {
		t.Errorf("un expect body %+v", string(req.Body()))
	}

	if code != 503 {
		t.Errorf("expect code 503 but %d", code)
	}
}

func TestPost(t *testing.T) {
	script := `
		function NewPlugin() {
			return {
				"post": function(c) {
					return {
						"code": 503,
						"error": "test failed",
					}
				}
			}
		}
	`
	meta := &metapb.Plugin{
		ID:      1,
		Name:    "test",
		Author:  "fagongzi",
		Email:   "zhangxu19830126@gmail.com",
		Version: 1,
		Type:    metapb.JavaScript,
		Content: []byte(script),
	}

	rt, err := NewRuntime(meta)
	if err != nil {
		t.Errorf("create runtime plugin failed with %+v", err)
	}

	code, err := rt.Post(&Ctx{})
	if err == nil {
		t.Errorf("expect error for pre")
	}

	if code != 503 {
		t.Errorf("expect code 503 but %d", code)
	}
}
