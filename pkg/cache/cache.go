package cache

import (
	"github.com/patrickmn/go-cache"
	"time"
)

// Cache is a wrapper for go-cache
var c *Cache
type Cache struct {
	*cache.Cache
}

func GetCache() *Cache {
	if c == nil {
		c = NewCache()
	}
	return c
}
// NewCache 	returns a new Cache
func NewCache() *Cache {
	return &Cache{
		Cache: cache.New(5*time.Minute, 10*time.Minute),
	}
}

// Set sets a value in the cache
func (c *Cache) Set(key string, value interface{}, expiration time.Duration) {
	c.Cache.Set(key, value, expiration)
}

// Get gets a value from the cache
func (c *Cache) Get(key string) (interface{}, bool) {
	return c.Cache.Get(key)
}

// Delete deletes a value from the cache
func (c *Cache) Delete(key string) {
	c.Cache.Delete(key)
}

// Clear clears the cache
func (c *Cache) Clear() {

	c.Cache.Flush()
}

// GetOrSet gets a value from the cache or sets it if it doesn't exist
func (c *Cache) GetOrSet(key string, value interface{}, expiration time.Duration) interface{} {
	if v, ok := c.Get(key); ok {
		return v
	}

	c.Set(key, value, expiration)

	return value
}
