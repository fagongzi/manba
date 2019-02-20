package plugin

import (
	"fmt"

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
)

// Runtime plugin runtime
type Runtime struct {
	id                             uint64
	vm                             *otto.Otto
	this                           otto.Value
	preFunc, postFunc, postErrFunc otto.Value
}

// NewRuntime return a runtime
func NewRuntime(src, cfg string) (*Runtime, error) {
	vm := otto.New()
	_, err := vm.Run(src)
	if err != nil {
		return nil, err
	}

	// add require for using go module
	vm.Set("require", Require)

	// exec constructor
	plugin, err := vm.Get(PluginConstructor)
	if err != nil {
		return nil, err
	}
	if !plugin.IsFunction() {
		return nil, fmt.Errorf("plugin constructor must be a function")
	}

	this, err := plugin.Call(plugin, cfg)
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

	postFunc, err := this.Object().Get(PluginPost)
	if err != nil {
		return nil, err
	}

	postErrFunc, err := this.Object().Get(PluginPostErr)
	if err != nil {
		return nil, err
	}

	return &Runtime{
		vm:          vm,
		this:        this,
		preFunc:     preFunc,
		postFunc:    postFunc,
		postErrFunc: postErrFunc,
	}, nil
}
