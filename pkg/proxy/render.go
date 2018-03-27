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
	template *metapb.RenderTemplate
	wg       *sync.WaitGroup
	nodes    []*dispathNode
	doRender func(*fasthttp.RequestCtx)
}

func newRender(nodes []*dispathNode, template *metapb.RenderTemplate) *render {
	rd := &render{
		template: template,
		nodes:    nodes,
	}

	rd.doRender = rd.renderSingle

	if len(nodes) > 1 {
		rd.wg = &sync.WaitGroup{}
		rd.wg.Add(len(nodes))
		rd.doRender = rd.renderMulti
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
		ctx.SetStatusCode(dn.code)
		dn.release()
		return
	}

	if rd.template == nil {
		rd.renderRaw(ctx, dn)
		return
	}

	src := jsoniter.ParseBytes(json, dn.res.Body()).ReadAny()
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

	for _, result := range rd.nodes {
		if hasError {
			result.release()
			continue
		}

		if result.err != nil ||
			result.code >= fasthttp.StatusBadRequest {
			hasError = true
			code = result.code
			err = result.err
			result.release()
			continue
		}

		for _, h := range MultiResultsRemoveHeaders {
			result.res.Header.Del(h)
		}
		result.res.Header.CopyTo(&ctx.Response.Header)

		value[result.node.meta.AttrName] = jsoniter.ParseBytes(json, result.res.Body()).ReadAny()
		result.release()
	}

	if hasError {
		log.Errorf("render: render failed, code=<%d>, errors:\n%+v",
			code,
			err)
		ctx.SetStatusCode(code)
		return
	}

	str, _ := jsoniter.MarshalToString(value)
	if rd.template == nil {
		ctx.WriteString(str)
		return
	}

	any := jsoniter.ParseString(json, str).ReadAny()
	str, _ = json.MarshalToString(rd.extract(any))
	ctx.WriteString(str)
	return
}

func (rd *render) renderRaw(ctx *fasthttp.RequestCtx, dn *dispathNode) {
	ctx.Response.Header.SetContentTypeBytes(dn.res.Header.ContentType())
	ctx.Write(dn.res.Body())
	dn.release()
}

func (rd *render) extract(src jsoniter.Any) map[string]interface{} {
	ret := make(map[string]interface{})
	for _, obj := range rd.template.Objects {
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
