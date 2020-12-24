package eggcache

import (
	"eggcache/lru"
	"sync"
)

type cache struct {
	lru        *lru.Cache //lru
	cacheBytes int64      // max size
	mu         sync.Mutex //并发安全
}

//save 插入或更新缓存
func (c *cache) save(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Save(key, value)
}

//get 获取缓存
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.lru == nil {
		return
	}
	if v, ok := c.lru.Get(key); ok {
		return v.(ByteView), ok
	}
	return
}
