package cache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	cacheMap map[string]cacheEntry
	mu       sync.Mutex
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cacheMap[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry, ok := c.cacheMap[key]
	return entry.val, ok
}

func (c *Cache) reapLoop(interval time.Duration, stop chan struct{}) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case t := <-ticker.C:
			c.mu.Lock()
			for key, value := range c.cacheMap {
				if t.Sub(value.createdAt) >= interval {
					delete(c.cacheMap, key)
				}
			}
			c.mu.Unlock()
		}
	}
}

func NewCache(interval time.Duration, stop chan struct{}) *Cache {
	cache := &Cache{
		cacheMap: make(map[string]cacheEntry),
	}

	go cache.reapLoop(interval, stop)
	return cache
}
