package proxy

import (
	"strings"
	"sync"

	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/log"
	"github.com/json-iterator/go"
	"github.com/valyala/fasthttp"
)

type render struct {
	multi    bool
	wg       *sync.WaitGroup
	api      *apiRuntime
	nodes    []*dispathNode
	doRender func(*fasthttp.RequestCtx)
}

func newRender(api *apiRuntime, nodes []*dispathNode) *render {
	rd := &render{
		nodes: nodes,
		api:   api,
	}

	rd.doRender = rd.renderSingle

	if len(nodes) > 1 {
		rd.wg = &sync.WaitGroup{}
		rd.wg.Add(len(nodes))
		rd.doRender = rd.renderMulti
	} else if len(nodes) == 0 {
		rd.doRender = rd.renderDefault
	}

	return rd
}

func (rd *render) render(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetContentType(MultiResultsContentType)
	ctx.SetStatusCode(fasthttp.StatusOK)

	rd.doRender(ctx)
}

func (rd *render) renderSingle(ctx *fasthttp.RequestCtx) {
	dn := rd.nodes[0]

	if dn.err != nil ||
		dn.code >= fasthttp.StatusBadRequest {
		log.Errorf("render: render failed, code=<%d>, errors:\n%+v",
			dn.code,
			dn.err)

		if rd.api.meta.DefaultValue != nil {
			rd.renderDefault(ctx)
			dn.release()
			return
		}

		ctx.SetStatusCode(dn.code)
		dn.release()
		return
	}

	if rd.api.meta.RenderTemplate == nil {
		rd.renderRaw(ctx, dn)
		return
	}

	src := jsoniter.ParseBytes(json, dn.getResponseBody()).ReadAny()
	dn.release()

	v, err := json.MarshalToString(rd.extract(src))
	if err != nil {
		log.Errorf("render: render failed, code=<%d>, errors:\n%+v",
			dn.code,
			err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}

	ctx.WriteString(v)
}

func (rd *render) renderMulti(ctx *fasthttp.RequestCtx) {
	rd.wg.Wait()

	var err error
	var hasError bool
	code := fasthttp.StatusInternalServerError
	value := make(map[string]interface{})

	for _, dn := range rd.nodes {
		if hasError {
			dn.release()
			continue
		}

		if dn.hasError() &&
			!dn.hasDefaultValue() {
			hasError = true
			code = dn.code
			err = dn.err
			dn.release()
			continue
		}

		dn.copyHeaderTo(ctx)
		value[dn.node.meta.AttrName] = jsoniter.ParseBytes(json, dn.getResponseBody()).ReadAny()
		dn.release()
	}

	if hasError {
		log.Errorf("render: render failed, code=<%d>, errors:\n%+v",
			code,
			err)

		if rd.api.meta.DefaultValue != nil {
			rd.renderDefault(ctx)
			return
		}

		ctx.SetStatusCode(code)
		return
	}

	str, _ := jsoniter.MarshalToString(value)
	if rd.api.meta.RenderTemplate == nil {
		ctx.WriteString(str)
		return
	}

	any := jsoniter.ParseString(json, str).ReadAny()
	str, _ = json.MarshalToString(rd.extract(any))
	ctx.WriteString(str)
	return
}

func (rd *render) renderRaw(ctx *fasthttp.RequestCtx, dn *dispathNode) {
	ctx.Response.Header.SetContentTypeBytes(dn.getResponseContentType())
	ctx.Write(dn.getResponseBody())
	dn.release()
}

func (rd *render) renderDefault(ctx *fasthttp.RequestCtx) {
	if rd.api.meta.DefaultValue == nil {
		return
	}

	header := &ctx.Response.Header

	for _, h := range rd.api.meta.DefaultValue.Headers {
		if h.Name == "Content-Type" {
			header.SetContentType(h.Value)
		} else {
			header.Add(h.Name, h.Value)
		}
	}

	for _, ck := range rd.api.defaultCookies {
		header.SetCookie(ck)
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Write(rd.api.meta.DefaultValue.Body)
}

func (rd *render) extract(src jsoniter.Any) map[string]interface{} {
	ret := make(map[string]interface{})
	for _, obj := range rd.api.meta.RenderTemplate.Objects {
		dest := ret

		if !obj.FlatAttrs {
			dest = make(map[string]interface{})
			ret[obj.Name] = dest
		}

		for _, attr := range obj.Attrs {
			extractValue(attr, src, dest)
		}
	}

	return ret
}

func extractValue(attr *metapb.RenderAttr, src jsoniter.Any, obj map[string]interface{}) {
	exps := strings.Split(attr.ExtractExp, ",")

	if len(exps) == 1 {
		obj[attr.Name] = src.Get(getPaths(attr.ExtractExp)...).GetInterface()
		return
	}

	sub := make(map[string]interface{})
	obj[attr.Name] = sub
	for _, exp := range exps {
		paths := getPaths(exp)
		sub[paths[len(paths)-1].(string)] = src.Get(paths...).GetInterface()
	}
}

func getPaths(attr string) []interface{} {
	var ret []interface{}

	for _, path := range strings.Split(attr, ".") {
		ret = append(ret, path)
	}

	return ret
}
