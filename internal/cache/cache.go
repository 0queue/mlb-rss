package cache

import "sync"

// just an excuse to play with generics
type Cache[T any] struct {
	m    sync.Mutex
	item T
	ok   bool
}

func (c *Cache[T]) Get() (T, bool) {
	c.m.Lock()
	defer c.m.Unlock()

	return c.item, c.ok
}

func (c *Cache[T]) Set(t T) {
	c.m.Lock()
	defer c.m.Unlock()

	c.item = t
	c.ok = true
}