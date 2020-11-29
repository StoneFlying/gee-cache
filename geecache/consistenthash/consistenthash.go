package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 哈希函数
type Hash func(data []byte) uint32

type Map struct {
	hash     Hash   // 哈希函数
	replicas int    // 每个真实节点对应多少个虚拟节点
	keys     []int  // 所有真时节点和虚拟节点，已排好序 
	hashMap  map[int]string  // 虚拟节点与真实节点的映射关系
}

// 新建一个map示例
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE  // 如果hash函数为空，则默认使用crc32.ChecksumIEEE哈希函数
	}
	return m 
}

// 添加节点到哈希环
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {  // 每个真实节点对应replicas个虚拟节点
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))  // 对key进行hash计算
			m.keys = append(m.keys, hash)   // 添加节点hash值到哈希环中
			m.hashMap[hash] = key    // 在hashMap中存储虚拟节点和真实节点的关系
		}
	}
	sort.Ints(m.keys)  // 进行排序
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {   // 不存在节点直接返回
		return ""
	}

	hash := int(m.hash([]byte(key)))   // 对键进行hash计算
	idx := sort.Search(len(m.keys), func(i int) bool {  // 顺时针寻找距离键最近的节点
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]   // 返回真实节点
}
