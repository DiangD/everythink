package lru

import "container/list"

//lru cache

type Value interface {
	//返回value占用的内存
	Len() int
}

type entry struct {
	key   string
	value Value
}

type Cache struct {
	maxBytes  int64
	nBytes    int64
	dl        *list.List //双向链表
	cache     map[string]*list.Element
	OnEvicted OnEvictedFunc //回调
}

type OnEvictedFunc func(key string, value Value)

func New(maxBytes int64, onEvicted OnEvictedFunc) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		dl:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

//Get 获取缓存
func (c *Cache) Get(key string) (Value, bool) {
	if ele, ok := c.cache[key]; ok {
		c.dl.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, ok
	}
	return nil, false
}

//Remove 删除缓存 最近最少使用
func (c *Cache) Remove() {
	ele := c.dl.Back()
	if ele != nil {
		c.dl.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nBytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

//Save 插入或更新缓存
func (c *Cache) Save(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.dl.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nBytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.dl.PushFront(&entry{
			key:   key,
			value: value,
		})
		c.cache[key] = ele
		c.nBytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes <= 0 || c.maxBytes < c.nBytes {
		c.Remove()
	}
}

//Len 获取缓存key的个数
func (c *Cache) Len() int {
	return c.dl.Len()
}
