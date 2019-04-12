package proxy

import (
	"context"
	"time"

	"github.com/fagongzi/gateway/pkg/plugin"
	"github.com/fagongzi/log"
)

func (p *Proxy) addGCJSEngine(value *plugin.Engine) {
	p.Lock()
	defer p.Unlock()

	p.gcJSEngines = append(p.gcJSEngines, value)
}

func (p *Proxy) readyToGCJSEngine() {
	_, err := p.runner.RunCancelableTask(func(ctx context.Context) {
		t := time.NewTicker(time.Minute)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Info("stop: gc js engine stopped")
				t.Stop()
				return
			case <-t.C:
				now := time.Now()
				p.Lock()
				var values []*plugin.Engine
				for _, eng := range p.gcJSEngines {
					if now.Sub(eng.LastActive()) > time.Hour {
						go eng.Destroy()
					} else {
						values = append(values, eng)
					}
				}
				p.gcJSEngines = values
				p.Unlock()
			}
		}
	})
	if err != nil {
		log.Fatalf("start gc js engine failed, errors:\n%+v", err)
	}
}
