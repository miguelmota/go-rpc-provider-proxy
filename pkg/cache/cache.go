package cache

import (
	"time"

	"github.com/patrickmn/go-cache"
	gocache "github.com/patrickmn/go-cache"
)

// Cache ...
type Cache struct {
	cache *gocache.Cache
}

// NewCache ...
func NewCache() *Cache {
	c := cache.New(5*time.Minute, 10*time.Minute)
	return &Cache{
		cache: c,
	}
}

// Set ...
func (c *Cache) Set(key string, value interface{}, expires time.Duration) {
	c.cache.Set(key, value, expires)
}

// Get ...
func (c *Cache) Get(key string) (interface{}, time.Time, bool) {
	return c.cache.GetWithExpiration(key)
}
