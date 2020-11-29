package geecache

import (
	"geecache/lru"
	"sync"
)

type cache struct {
	mu         sync.Mutex  // 互斥锁
	lru        *lru.Cache  // 缓存实例
	cacheBytes int64   // 最大缓存大小
}

// 添加键值对到缓存中
func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()  // 互斥访问
	defer c.mu.Unlock()
	if c.lru == nil {  // 延迟初始化
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)  // 添加到缓存中
}

// 根据键从缓存中取值
func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()  // 互斥访问
	defer c.mu.Unlock()
	if c.lru == nil { 
		return
	}

	if v, ok := c.lru.Get(key); ok {  // 从缓存中取值
		return v.(ByteView), ok
	}

	return
}
