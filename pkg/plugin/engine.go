package plugin

import (
	"fmt"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/log"
)

// Engine plugin engine
type Engine struct {
	filter.BaseFilter

	applied []*Runtime
}

// NewEngine returns a plugin engine
func NewEngine() *Engine {
	return &Engine{}
}

// UpdatePlugin update plugin
func (eng *Engine) UpdatePlugin(plugin *metapb.Plugin) error {

	target := -1
	for idx, rt := range eng.applied {
		if rt.id == plugin.ID {
			target = idx
			break
		}
	}

	if target == -1 {
		return fmt.Errorf("plugin not found")
	}

	rt, err := NewRuntime(string(plugin.Content), string(plugin.Cfg))
	if err != nil {
		return err
	}

	eng.applied[target] = rt
	return nil
}

// ApplyPlugins apply plugins
func (eng *Engine) ApplyPlugins(plugins ...*metapb.Plugin) error {
	var applied []*Runtime
	for idx, plugin := range plugins {
		rt, err := NewRuntime(string(plugin.Content), string(plugin.Cfg))
		if err != nil {
			return err
		}

		rt.id = plugin.ID
		applied = append(applied, rt)
		log.Infof("plugin: %d/%s:%d applied with index %d",
			plugin.ID,
			plugin.Name,
			plugin.Version,
			idx)
	}

	eng.applied = applied
	return nil
}

// Name returns filter name
func (eng *Engine) Name() string {
	return "JS-Plugin-Engine"
}

// Init returns error if init failed
func (eng *Engine) Init(cfg string) error {
	return nil
}

// Pre filter pre method
func (eng *Engine) Pre(c filter.Context) (int, error) {
	return eng.BaseFilter.Pre(c)
}

// Post filter post method
func (eng *Engine) Post(c filter.Context) (int, error) {
	return eng.BaseFilter.Post(c)
}

// PostErr filter post error method
func (eng *Engine) PostErr(c filter.Context) {
	eng.BaseFilter.PostErr(c)
}
