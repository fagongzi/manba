package plugin

import (
	"github.com/robertkrimen/otto"
)

// Engine plugin engine
type Engine struct {
	vm *otto.Otto
}

// NewEngine returns a plugin engine
func NewEngine() (*Engine, error) {
	eng := &Engine{
		vm: otto.New(),
	}

	err := eng.init()
	if err != nil {
		return nil, err
	}

	return eng, nil
}

func (eng *Engine) init() error {
	return nil
}
