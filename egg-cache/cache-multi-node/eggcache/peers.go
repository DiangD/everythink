package eggcache

//PeerPicker 定位key所在的节点
type PeerPicker interface {
	PickPeer(key string) (PeerGetter, bool)
}

//PeerGetter 调用远程节点获取缓存接口，每一个peer都必须实现。
type PeerGetter interface {
	Get(group, key string) ([]byte, error)
}
