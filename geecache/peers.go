package geecache

// 根据key选择节点
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// 节点调用Get函数从后端获取值
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
