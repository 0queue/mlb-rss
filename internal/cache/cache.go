package cache

import "sync"

// just an excuse to play with generics
type Cache[T any] struct {
	m    sync.Mutex
	item T
}

func (c *Cache[T]) Get() T {
	c.m.Lock()
	defer c.m.Unlock()

	return c.item
}

func (c *Cache[T]) Set(t T) {
	c.m.Lock()
	defer c.m.Unlock()

	c.item = t
}