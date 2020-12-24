package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func([]byte) uint32

type Map struct {
	hash     Hash           //hash函数
	replicas int            //虚拟节点的倍数
	keys     []int          //哈希环
	hashmap  map[int]string //虚拟节点hash与名称的映射
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashmap:  make(map[int]string),
	}
	if fn == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

//Add 插入多个真实节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		//建立虚拟节点
		for i := 0; i < m.replicas; i++ {
			//计算hash
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashmap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

//Get 获取节点名称
func (m *Map) Get(key string) string {
	if len(key) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	//按照hash查找节点
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	//将虚拟节点映射到真实节点
	return m.hashmap[m.keys[idx%len(m.keys)]]
}
