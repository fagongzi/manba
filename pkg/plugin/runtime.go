package plugin

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/log"
	"github.com/robertkrimen/otto"
)

const (
	// PluginConstructor every plugin needs a constructor
	PluginConstructor = "NewPlugin"
	// PluginPre filter pre method
	PluginPre = "pre"
	// PluginPost filter post method
	PluginPost = "post"
	// PluginPostErr filter post error method
	PluginPostErr = "postErr"

	// PluginReturnCodeFieldName code field name in return json object
	PluginReturnCodeFieldName = "code"
	// PluginReturnErrorFieldName error field name in return json object
	PluginReturnErrorFieldName = "error"
)

// Runtime plugin runtime
type Runtime struct {
	filter.BaseFilter

	meta                           *metapb.Plugin
	vm                             *otto.Otto
	this                           otto.Value
	preFunc, postFunc, postErrFunc otto.Value
}

// NewRuntime return a runtime
func NewRuntime(meta *metapb.Plugin) (*Runtime, error) {
	vm := otto.New()
	// add require for using go module
	vm.Set("require", Require)
	vm.Set("BreakFilterChainCode", filter.BreakFilterChainCode)
	vm.Set("UsingResponse", filter.UsingResponse)

	_, err := vm.Run(string(meta.Content))
	if err != nil {
		return nil, err
	}

	// exec constructor
	plugin, err := vm.Get(PluginConstructor)
	if err != nil {
		return nil, err
	}
	if !plugin.IsFunction() {
		return nil, fmt.Errorf("plugin constructor must be a function")
	}

	this, err := plugin.Call(plugin, string(meta.Cfg))
	if err != nil {
		return nil, err
	}
	if !this.IsObject() {
		return nil, fmt.Errorf("plugin constructor must be return an object")
	}

	// fetch plugin methods
	preFunc, err := this.Object().Get(PluginPre)
	if err != nil {
		return nil, err
	}
	if preFunc.IsDefined() && !preFunc.IsFunction() {
		return nil, fmt.Errorf("pre must function")
	}

	postFunc, err := this.Object().Get(PluginPost)
	if err != nil {
		return nil, err
	}
	if postFunc.IsDefined() && !postFunc.IsFunction() {
		return nil, fmt.Errorf("post must function")
	}

	postErrFunc, err := this.Object().Get(PluginPostErr)
	if err != nil {
		return nil, err
	}
	if postErrFunc.IsDefined() && !postErrFunc.IsFunction() {
		return nil, fmt.Errorf("postErr must function")
	}

	return &Runtime{
		meta:        meta,
		vm:          vm,
		this:        this,
		preFunc:     preFunc,
		postFunc:    postFunc,
		postErrFunc: postErrFunc,
	}, nil
}

// Pre filter pre method
func (rt *Runtime) Pre(c *Ctx) (int, error) {
	if rt.preFunc.IsUndefined() {
		return rt.BaseFilter.Pre(c.delegate)
	}

	value, err := rt.preFunc.Call(rt.this, c)
	if err != nil {
		log.Errorf("plugin %d/%s:%d plugin pre func failed with %+v",
			rt.meta.ID,
			rt.meta.Name,
			rt.meta.Version,
			err)
		return http.StatusInternalServerError, err
	}

	if !value.IsObject() {
		return http.StatusInternalServerError, fmt.Errorf("unexpect js plugin returned %+v", value)
	}

	return parsePluginReturn(value.Object())
}

// Post filter post method
func (rt *Runtime) Post(c *Ctx) (int, error) {
	if rt.postFunc.IsUndefined() {
		return rt.BaseFilter.Post(c.delegate)
	}

	value, err := rt.postFunc.Call(rt.this, c)
	if err != nil {
		log.Errorf("plugin %d/%s:%d plugin post func failed with %+v",
			rt.meta.ID,
			rt.meta.Name,
			rt.meta.Version,
			err)
		return http.StatusInternalServerError, err
	}

	if !value.IsObject() {
		return http.StatusInternalServerError, fmt.Errorf("unexpect js plugin returned %+v", value)
	}

	return parsePluginReturn(value.Object())
}

// PostErr filter post error method
func (rt *Runtime) PostErr(c *Ctx) {
	if rt.postErrFunc.IsUndefined() {
		rt.BaseFilter.PostErr(c.delegate)
		return
	}

	_, err := rt.postErrFunc.Call(rt.this, c)
	if err != nil {
		log.Errorf("plugin %d/%s:%d plugin post error func failed with %+v",
			rt.meta.ID,
			rt.meta.Name,
			rt.meta.Version,
			err)
	}
}

func parsePluginReturn(value *otto.Object) (int, error) {
	code, err := value.Get(PluginReturnCodeFieldName)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if !code.IsNumber() {
		return http.StatusInternalServerError, fmt.Errorf("js plugin result code must be number")
	}
	statusCode, err := code.ToInteger()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	e, err := value.Get(PluginReturnErrorFieldName)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	if e.IsDefined() && !e.IsString() {
		return http.StatusInternalServerError, fmt.Errorf("js plugin result error must be string")
	}

	if e.IsUndefined() {
		return int(statusCode), nil
	}

	errMsg, err := e.ToString()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return int(statusCode), errors.New(errMsg)
}
