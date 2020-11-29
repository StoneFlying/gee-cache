package singleflight

import "sync"

type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

type Group struct {
	mu sync.Mutex       // 互斥锁
	m  map[string]*call // lazily initialized
}


func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()   
	if g.m == nil {   // 如果map还未初始化则进行初始化
		g.m = make(map[string]*call)
	}
	if c, ok := g.m[key]; ok {  // 如果键正则访问，则进行等待
		g.mu.Unlock()
		c.wg.Wait()
		return c.val, c.err   // 其他协程访问完毕，直接获取值
	}
	c := new(call)  // 第一次访问，新建一个call结构体表面当前key正在访问
	c.wg.Add(1)  // 计算加1
	g.m[key] = c   // 更新map
	g.mu.Unlock()

	c.val, c.err = fn()  // 进行访问取值
	c.wg.Done()  // 访问完毕，计数减1，其他协程返回

	g.mu.Lock()
	delete(g.m, key)  // 访问完毕，删除g.m中存储的对应的key
	g.mu.Unlock()

	return c.val, c.err
}
