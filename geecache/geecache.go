package geecache

import (
	"fmt"
	"geecache/singleflight"
	"log"
	"sync"
)

// 一个缓存的命名空间
type Group struct {
	name      string  // 缓存名字
	getter    Getter  // 缓存未命中时执行的回调函数
	mainCache cache   // 缓存实例
	peers     PeerPicker   // 所有节点
	loader *singleflight.Group  // singleflight防止缓存击穿
}

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)  // 所有缓存命名空间
)

// 新建一个Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,  
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// 根据Group名字返回Group
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// 从缓存中获取值
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok { // 从缓存中获取
		log.Println("[GeeCache] hit")
		return v, nil
	}

	return g.load(key)  // 如果缓存中不存在，则从本地或者其他节点后端获取
}

// 注册节点
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) load(key string) (value ByteView, err error) {
	// 通过singleflight防止缓存击穿
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {   // 根据key选择对应的节点
				if value, err = g.getFromPeer(peer, key); err == nil {  // 如果key存在其他节点，则从其他节点获取
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}

		return g.getLocally(key) // 否则从本地后端获取
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)   // 从本地后端获取后，添加到缓存中
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)  // 从本地后端获取
	if err != nil {
		return ByteView{}, err

	}
	value := ByteView{b: cloneBytes(bytes)} 
	g.populateCache(key, value)   // 存储到本地缓存中
	return value, nil
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)   // 从其他节点获取
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}
