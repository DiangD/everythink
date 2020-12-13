package xclient

import (
	"errors"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

//负载均衡算法
const (
	//随机选=选择策略
	RandomSelect SelectMode = iota
	//轮询
	RoundRobinSelect
)

//Discovery 服务发现接口
type Discovery interface {
	Refresh() error
	Update(servers []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}

type MultiServersDiscovery struct {
	r       *rand.Rand //随机函数
	mu      sync.RWMutex
	servers []string //注册的服务
	index   int      //保存当前轮询的位置
}

func NewMultiServersDiscovery(servers []string) *MultiServersDiscovery {
	discovery := &MultiServersDiscovery{
		servers: servers,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	discovery.index = rand.Intn(math.MaxInt32 - 1)
	return discovery
}

func (d *MultiServersDiscovery) Refresh() error {
	return nil
}

func (d *MultiServersDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.servers = servers
	return nil
}

//Get 获取服务
func (d *MultiServersDiscovery) Get(mode SelectMode) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	seed := len(d.servers)
	if seed == 0 {
		return "", errors.New("rpc discovery: no available servers")
	}
	switch mode {
	case RandomSelect:
		return d.servers[d.r.Intn(seed)], nil
	case RoundRobinSelect:
		s := d.servers[d.index%seed]
		d.index = (d.index + 1) % seed
		return s, nil
	default:
		return "", errors.New("rpc discovery: not supported select mode")
	}
}

func (d *MultiServersDiscovery) GetAll() ([]string, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	tmp := make([]string, len(d.servers), len(d.servers))
	copy(tmp, d.servers)
	return tmp, nil
}
