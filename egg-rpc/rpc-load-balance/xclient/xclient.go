package xclient

import (
	"context"
	. "eggrpc"
	"reflect"
	"sync"
)

type XClient struct {
	d       Discovery
	mode    SelectMode
	mu      sync.Mutex
	opt     *Option
	clients map[string]*Client
}

func (xc *XClient) Close() error {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	for k, client := range xc.clients {
		_ = client.Close()
		delete(xc.clients, k)
	}
	return nil
}

func NewXClient(d Discovery, mode SelectMode, opt *Option) *XClient {
	return &XClient{
		d:       d,
		mode:    mode,
		opt:     opt,
		clients: make(map[string]*Client),
	}
}

func (xc *XClient) dial(rpcAddr string) (*Client, error) {
	xc.mu.Lock()
	defer xc.mu.Unlock()
	client, ok := xc.clients[rpcAddr]
	if ok && !client.IsAvailable() {
		_ = client.Close()
		delete(xc.clients, rpcAddr)
		client = nil
	}
	if client == nil {
		var err error
		client, err = XDial(rpcAddr, xc.opt)
		if err != nil {
			return nil, err
		}
		xc.clients[rpcAddr] = client
	}
	return client, nil
}

func (xc *XClient) call(ctx context.Context, rpcAddr, serviceMethod string, args, reply interface{}) error {
	client, err := xc.dial(rpcAddr)
	if err != nil {
		return err
	}
	return client.Call(ctx, serviceMethod, args, reply)
}

func (xc *XClient) Call(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	rpcAddr, err := xc.d.Get(xc.mode)
	if err != nil {
		return err
	}
	return xc.call(ctx, rpcAddr, serviceMethod, args, reply)
}

func (xc *XClient) Broadcast(ctx context.Context, serviceMethod string, args, reply interface{}) error {
	servers, err := xc.d.GetAll()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	var e error
	var mu sync.Mutex
	replyDone := reply == nil
	ctx, cancel := context.WithCancel(ctx)
	for _, server := range servers {
		wg.Add(1)
		rpcAddr := server
		go func() {
			defer wg.Done()
			var replyTmp interface{}
			if reply != nil {
				replyTmp = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
			}
			err := xc.call(ctx, rpcAddr, serviceMethod, args, reply)
			mu.Lock()
			if err != nil && e == nil {
				e = err
				cancel()
			}
			if err == nil && !replyDone {
				reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(replyTmp).Elem())
				replyDone = true
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	return e
}
