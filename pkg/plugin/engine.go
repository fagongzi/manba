package plugin

import (
	"fmt"
	"net/http"
	"time"

	"github.com/fagongzi/gateway/pkg/filter"
	"github.com/fagongzi/gateway/pkg/pb/metapb"
	"github.com/fagongzi/log"
)

// Engine plugin engine
type Engine struct {
	filter.BaseFilter

	enable     bool
	applied    []*Runtime
	lastActive time.Time
}

// NewEngine returns a plugin engine
func NewEngine(enable bool) *Engine {
	return &Engine{
		enable: enable,
	}
}

// LastActive returns the time that last used
func (eng *Engine) LastActive() time.Time {
	return eng.lastActive
}

// Destroy destory all applied plugins
func (eng *Engine) Destroy() {
	for _, rt := range eng.applied {
		rt.destroy()
	}
}

// UpdatePlugin update plugin
func (eng *Engine) UpdatePlugin(plugin *metapb.Plugin) error {
	target := -1
	for idx, rt := range eng.applied {
		if rt.meta.ID == plugin.ID {
			target = idx
			break
		}
	}

	if target == -1 {
		return fmt.Errorf("plugin not found")
	}

	rt, err := NewRuntime(plugin)
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
		rt, err := NewRuntime(plugin)
		if err != nil {
			return err
		}

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
	if !eng.enable {
		return eng.BaseFilter.Pre(c)
	}

	eng.lastActive = time.Now()

	if len(eng.applied) == 0 {
		return eng.BaseFilter.Pre(c)
	}

	rc := acquireContext()
	rc.delegate = c
	for _, rt := range eng.applied {
		statusCode, err := rt.Pre(rc)
		if nil != err {
			releaseContext(rc)
			return statusCode, err
		}

		if statusCode == filter.BreakFilterChainCode {
			releaseContext(rc)
			return statusCode, err
		}
	}

	releaseContext(rc)
	return http.StatusOK, nil
}

// Post filter post method
func (eng *Engine) Post(c filter.Context) (int, error) {
	if !eng.enable {
		return eng.BaseFilter.Post(c)
	}

	eng.lastActive = time.Now()

	if len(eng.applied) == 0 {
		return eng.BaseFilter.Post(c)
	}

	rc := acquireContext()
	rc.delegate = c

	l := len(eng.applied)
	for i := l - 1; i >= 0; i-- {
		rt := eng.applied[i]

		statusCode, err := rt.Post(rc)
		if nil != err {
			releaseContext(rc)
			return statusCode, err
		}

		if statusCode == filter.BreakFilterChainCode {
			releaseContext(rc)
			return statusCode, err
		}
	}

	releaseContext(rc)
	return http.StatusOK, nil
}

// PostErr filter post error method
func (eng *Engine) PostErr(c filter.Context, code int, err error) {
	if !eng.enable {
		eng.BaseFilter.PostErr(c, code, err)
		return
	}

	eng.lastActive = time.Now()

	if len(eng.applied) == 0 {
		eng.BaseFilter.PostErr(c, code, err)
		return
	}

	rc := acquireContext()
	rc.delegate = c

	l := len(eng.applied)
	for i := l - 1; i >= 0; i-- {
		eng.applied[i].PostErr(rc, code, err)
	}

	releaseContext(rc)
}
