package proxy

import (
	"sync"

	"github.com/buger/jsonparser"
	"github.com/fagongzi/log"
	"github.com/fagongzi/util/hack"
)

type multiContext struct {
	sync.RWMutex
	data []byte
}

func (c *multiContext) reset() {
	c.init()
}

func (c *multiContext) init() {
	c.data = emptyObject
}

func (c *multiContext) completePart(attr string, data []byte) {
	c.Lock()
	if len(data) > 0 {
		c.data, _ = jsonparser.Set(c.data, data, attr)
	}
	c.Unlock()
}

func (c *multiContext) getAttr(paths ...string) string {
	c.RLock()
	value, _, _, err := jsonparser.Get(c.data, paths...)
	c.RUnlock()
	if err != nil {
		log.Errorf("extract %+v failed, errors:\n%+v", paths, err)
		return ""
	}

	return hack.SliceToString(value)
}
