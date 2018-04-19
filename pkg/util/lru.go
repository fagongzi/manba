package util

import (
	"container/list"
	"sync"
)

// Cache is an LRU cache. It is not safe for concurrent access.
type Cache struct {
	sync.RWMutex

	// MaxBytes is the maximum bytes of cache entries before
	// an item is evicted. Zero means no limit.
	MaxBytes uint64
	current  uint64

	// OnEvicted optionally specificies a callback function to be
	// executed when an entry is purged from the cache.
	OnEvicted func(key Key, value interface{})

	ll    *list.List
	cache map[interface{}]*list.Element
}

// A Key may be any value that is comparable. See http://golang.org/ref/spec#Comparison_operators
type Key interface{}

type entry struct {
	key   Key
	value []byte
}

// NewLRUCache creates a new Cache.
// If maxBytes is zero, the cache has no limit and it's assumed
// that eviction is done by the caller.
func NewLRUCache(maxBytes uint64) *Cache {
	return &Cache{
		MaxBytes: maxBytes,
		ll:       list.New(),
		cache:    make(map[interface{}]*list.Element),
	}
}

// Add adds a value to the cache.
func (c *Cache) Add(key Key, value []byte) {
	c.Lock()

	if c.cache == nil {
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}
	if ee, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ee)

		entry := ee.Value.(*entry)
		c.current -= uint64(len(entry.value))
		c.current += uint64(len(value))
		entry.value = value
		c.Unlock()
		return
	}

	c.current += uint64(len(value))
	ele := c.ll.PushFront(&entry{key, value})
	c.cache[key] = ele
	if c.MaxBytes != 0 && c.current > c.MaxBytes {
		c.removeOldest()
	}
	c.Unlock()
}

// Get looks up a key's value from the cache.
func (c *Cache) Get(key Key) (value []byte, ok bool) {
	c.RLock()

	if c.cache == nil {
		c.RUnlock()
		return
	}

	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		c.RUnlock()
		return ele.Value.(*entry).value, true
	}

	c.RUnlock()
	return
}

// Remove removes the provided key from the cache.
func (c *Cache) Remove(key Key) {
	c.Lock()

	if c.cache == nil {
		c.Unlock()
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}

	c.Unlock()
}

func (c *Cache) removeOldest() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	c.current -= uint64(len(kv.value))
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value)
	}
}

// Len returns the number of items in the cache.
func (c *Cache) Len() int {
	c.RLock()
	if c.cache == nil {
		c.RUnlock()
		return 0
	}
	value := c.ll.Len()
	c.RUnlock()
	return value
}

// Clear purges all stored items from the cache.
func (c *Cache) Clear() {
	c.Lock()
	if c.OnEvicted != nil {
		for _, e := range c.cache {
			kv := e.Value.(*entry)
			c.OnEvicted(kv.key, kv.value)
		}
	}
	c.ll = nil
	c.cache = nil
	c.current = 0
	c.Unlock()
}
