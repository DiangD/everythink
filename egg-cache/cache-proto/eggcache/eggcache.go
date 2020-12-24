package eggcache

import (
	pb "eggcache/pb"
	"eggcache/singleflight"
	"errors"
	"log"
	"sync"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	getter    Getter //回调
	mainCache cache
	peers     PeerPicker
	loader    *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}
	mu.Lock()
	defer mu.Unlock()
	group := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = group
	return group
}

func GetGroup(name string) *Group {
	mu.RLock()
	group := groups[name]
	mu.RUnlock()
	return group
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

//Get 获取缓存
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, errors.New("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Printf("Eggcache hit,key:%v", key)
		return v, nil
	}
	return g.load(key)
}

//load
func (g *Group) load(key string) (view ByteView, err error) {
	val, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if view, err = g.getFromPeer(key, peer); err == nil {
					return view, nil
				}
				log.Println("[EggCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return val.(ByteView), nil
	}
	return
}

func (g *Group) getFromPeer(key string, peer PeerGetter) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	resp := &pb.Response{}
	err := peer.Get(req, resp)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{bytes: resp.Value}, nil
}

//getLocally 读取本地数据源
func (g *Group) getLocally(key string) (ByteView, error) {
	//读取数据源
	//写法很骚，定义一个函数类型 F，并且实现接口 A 的方法，然后在这个方法中调用自己。这是 Go 语言中将其他函数（参数返回值定义与 F 一致）转换为接口 A 的常用技巧。
	resource, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}
	val := ByteView{bytes: cloneBytes(resource)}
	g.populateCache(key, val)
	return val, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.save(key, value)
}
