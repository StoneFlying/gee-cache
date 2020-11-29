package lru

import "container/list"

// 缓存结构体
type Cache struct {
	maxBytes int64       // 最大缓存容量
	nbytes   int64       // 已缓存容量
	ll       *list.List  // 缓存队列
	cache    map[string]*list.Element   // 通过map实现O(1)访问
	OnEvicted func(key string, value Value)  // 删除元素时执行
}

// 元素节点，包括key和value，value需要实现接口len函数返回其占用空间大小
type entry struct { 
	key   string
	value Value
}

// 值占用空间大小
type Value interface {
	Len() int
}

// 新建缓存
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// 添加一个键值对到Cache
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {  // 如果元素存在cache中，则将其移动到对头
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)    
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())  // 更新已使用空间大小，旧值与新值的差
		kv.value = value   // 更新value
	} else {
		ele := c.ll.PushFront(&entry{key, value})  // 如果没有存在cache中，则添加到对头
		c.cache[key] = ele    // 相应的添加到map中
		c.nbytes += int64(len(key)) + int64(value.Len()) // 更新已占用空间为
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {  // 如果超过最大空间，则删除队尾元素
		c.RemoveOldest()
	}
}

// 根据键从缓存中获取值
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {   // 如果元素存在缓存中
		c.ll.MoveToFront(ele)  // 由于是新访问，将其移到对头
		kv := ele.Value.(*entry)  
		return kv.value, true  // 返回键对应的值
	}
	return  // 不存在缓存中则直接返回空
}

// 从缓存中移除最久未访问的元素
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()  // 取得队尾元素
	if ele != nil {
		c.ll.Remove(ele)  // 删除队尾元素
		kv := ele.Value.(*entry) 
		delete(c.cache, kv.key)  // 相应的从map中删除
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())  // 更新已使用空间
		if c.OnEvicted != nil {  // 如果删除元素后需要执行的回调函数不为空，则执行回调函数
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// 获取缓存系统中的元素数目
func (c *Cache) Len() int {
	return c.ll.Len()
}
