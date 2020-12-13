package registry

import (
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

type EggRegistry struct {
	timeout time.Duration
	mu      sync.Mutex
	servers map[string]*ServerItem
}

type ServerItem struct {
	Addr  string
	start time.Time
}

const (
	defaultPath    = "/_eggrpc_/registry"
	defaultTimeout = time.Minute * 5
)

func New(timeout time.Duration) *EggRegistry {
	return &EggRegistry{
		timeout: timeout,
		servers: make(map[string]*ServerItem),
	}
}

var DefaultEggRegistry = New(defaultTimeout)

func (r *EggRegistry) putServer(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	server, ok := r.servers[addr]
	if !ok {
		r.servers[addr] = &ServerItem{
			Addr:  addr,
			start: time.Now(),
		}
	} else {
		server.start = time.Now()
	}
}

func (r *EggRegistry) aliveServers() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	alive := make([]string, 0)
	for addr, server := range r.servers {
		if r.timeout == 0 || server.start.Add(r.timeout).After(time.Now()) {
			alive = append(alive, addr)
		} else {
			delete(r.servers, addr)
		}
	}
	sort.Strings(alive)
	return alive
}

func (r *EggRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodGet:
		w.Header().Set("X-Eggrpc-Servers", strings.Join(r.aliveServers(), ","))
	case http.MethodPost:
		addr := req.Header.Get("X-Eggrpc-Server")
		if addr == "" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.putServer(addr)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (r *EggRegistry) HandleHTTP(registryPath string) {
	http.Handle(registryPath, r)
	log.Println("rpc registry path:", registryPath)
}

func HandleHTTP() {
	DefaultEggRegistry.HandleHTTP(defaultPath)
}

func Heartbeat(registry, addr string, duration time.Duration) {
	if duration == 0 {
		duration = defaultTimeout - time.Minute
	}
	var err error
	err = sendHeartbeat(registry, addr)
	go func() {
		ticker := time.NewTicker(duration)
		for err == nil {
			<-ticker.C
			err = sendHeartbeat(registry, addr)
		}
	}()
}

func sendHeartbeat(registry, addr string) error {
	log.Println(addr, "send heart beat to registry", registry)
	req, _ := http.NewRequest(http.MethodPost, registry, nil)
	req.Header.Set("X-Eggrpc-Server", addr)
	if _, err := http.DefaultClient.Do(req); err != nil {
		log.Println("rpc server: heart beat err:", err)
		return err
	}
	return nil
}
