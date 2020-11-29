package geecache

import (
	"hash"
	"hash/crc64"
	"hash/fnv"
	"log"
)

type IFilter interface {
	Push([]byte)
	Exists([]byte) bool
	Close() error
	Write() error
	IsEmpty() bool
}

var DefaultHash = []hash.Hash64{fnv.New64(), crc64.New(crc64.MakeTable(crc64.ISO))}   // 哈希函数

type Filter struct {
	Bytes  []byte   // byte切片
	Hashes []hash.Hash64
}

// 向布隆过滤器添加元素
func (f *Filter) Push(str []byte) {
	var byteLen = len(f.Bytes)
	for _, v := range f.Hashes {
		v.Reset()
		_, err := v.Write(str)
		if err != nil {
			log.Println(err.Error())
		}
		var res = v.Sum64()
		var yByte = res % uint64(byteLen)  // 结果在哪一字节
		var yBit = res & 7  // 在字节的哪一位
		var now = f.Bytes[yByte] | 1 << yBit  // 对应字节的对应位置1
		if now != f.Bytes[yByte] {  // 不相等则进行更改
			f.Bytes[yByte] = now
		}

	}
}

// 查看元素是否在布隆过滤器中
func (f *Filter) Exists(str []byte) bool {
	var byteLen = len(f.Bytes)
	for _, v := range f.Hashes {
		v.Reset()
		_, err := v.Write(str)
		if err != nil {
			log.Println(err.Error())
		}
		var res = v.Sum64()
		var yByte = res % uint64(byteLen)   // 结果在哪一字节
		var yBit = res & 7   // 在字节的哪一位
		if f.Bytes[yByte]|1<<yBit != f.Bytes[yByte] {  // 是否相等，相等则可能存在过滤器中
			return false
		}
	}
	return true
}

// 判断布隆过滤器是否为空
func (f *Filter) IsEmpty() bool {
	for i, _ := range f.Bytes {
		if f.Bytes[i] != 0 {
			return false
		}
	}
	return true
}
