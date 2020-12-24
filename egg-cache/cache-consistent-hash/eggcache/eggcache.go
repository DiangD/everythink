package eggcache

import (
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
func (g *Group) load(key string) (ByteView, error) {
	return g.getLocally(key)
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
