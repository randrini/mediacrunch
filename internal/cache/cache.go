package cache

import (
	"sync"
	"time"
)

type entry struct {
	value  interface{}
	expiry time.Time
}

type Cache struct {
	mu    sync.RWMutex
	items map[string]entry
	ttl   time.Duration
}

func New(ttl time.Duration) *Cache {
	c := &Cache{
		items: make(map[string]entry),
		ttl:   ttl,
	}
	go c.cleanup()
	return c
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok || time.Now().After(e.expiry) {
		return nil, false
	}
	return e.value, true
}

func (c *Cache) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = entry{value: value, expiry: time.Now().Add(c.ttl)}
}

func (c *Cache) Invalidate(prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if prefix == "" {
		c.items = make(map[string]entry)
		return
	}
	for k := range c.items {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(c.items, k)
		}
	}
}

func (c *Cache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, e := range c.items {
			if now.After(e.expiry) {
				delete(c.items, k)
			}
		}
		c.mu.Unlock()
	}
}
