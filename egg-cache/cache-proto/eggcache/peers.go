package eggcache

import pb "eggcache/pb"

//PeerPicker 定位key所在的节点
type PeerPicker interface {
	PickPeer(key string) (PeerGetter, bool)
}

//PeerGetter 调用远程节点获取缓存接口，每一个peer都必须实现。
type PeerGetter interface {
	Get(req *pb.Request, resp *pb.Response) error
}
