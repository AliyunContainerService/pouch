package containerio

import (
	"github.com/alibaba/pouch/pkg/collect"
)

// Cache saves the all container's io.
type Cache struct {
	m *collect.SafeMap
}

// NewCache creates a container's io storage.
func NewCache() *Cache {
	return &Cache{
		m: collect.NewSafeMap(),
	}
}

// Put writes a container's io into storage.
func (c *Cache) Put(id string, io *IO) error {
	c.m.Put(id, io)
	return nil
}

// Get reads a container's io by id.
func (c *Cache) Get(id string) *IO {
	obj, ok := c.m.Get(id).Result()
	if !ok {
		return nil
	}

	if io, ok := obj.(*IO); ok {
		return io
	}
	return nil
}

// Remove removes the container's io.
func (c *Cache) Remove(id string) {
	c.m.Remove(id)
}
