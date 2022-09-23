package giCache

import "sync"

type Cache struct {
	mu        sync.RWMutex
	container *LruCache
}

func NewCache(capacity int, evictedFunc EvictedFunc) *Cache {
	return &Cache{
		mu:        sync.RWMutex{},
		container: NewLruCache(capacity, evictedFunc),
	}
}

func (c *Cache) Get(key string) (val Value, ok bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.container.Get(key)
}

func (c *Cache) Set(key string, val Value) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.container.Set(key, val)
}
