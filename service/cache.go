package main

import "sync"

type BucketObjectCache struct {
	cache map[string][]byte
	mu    sync.RWMutex
}

func NewCache() *BucketObjectCache {
	return &BucketObjectCache{
		cache: make(map[string][]byte),
	}
}

func (c *BucketObjectCache) Set(bo *BucketObject, data []byte) {
	path := bo.Path()
	c.mu.Lock()
	c.cache[path] = data
	c.mu.Unlock()
}

func (c *BucketObjectCache) Get(bo *BucketObject) (data []byte, ok bool) {
	path := bo.Path()
	c.mu.RLock()
	data, ok = c.cache[path]
	c.mu.RUnlock()
	return data, ok
}

func (c *BucketObjectCache) Exists(bo *BucketObject) bool {
	_, ok := c.Get(bo)
	return ok
}
