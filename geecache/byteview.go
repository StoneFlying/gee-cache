package geecache

// 字节切片
type ByteView struct {
	b []byte
}

// 返回切片大小
func (v ByteView) Len() int {
	return len(v.b)
}

// 返回一个切片副本
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// 将切片转换为string返回
func (v ByteView) String() string {
	return string(v.b)
}

// 复制切片内容
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
