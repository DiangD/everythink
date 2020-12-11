package eggrpc

import (
	"bufio"
	"context"
	"eggrpc/codec"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Call struct {
	Seq           uint64
	ServiceMethod string
	Args          interface{}
	reply         interface{}
	Error         error
	Done          chan *Call
}

func (call *Call) done() {
	select {
	case call.Done <- call:
	default:
		log.Println("rpc: discarding Call reply due to insufficient Done chan capacity")
	}
}

type Client struct {
	cc       codec.Codec
	opt      *Option
	sending  sync.Mutex
	header   codec.Header
	mu       sync.Mutex
	seq      uint64
	pending  map[uint64]*Call
	closing  bool
	shutdown bool
}

//断言，可以在编译器确定Client是否实现Closer接口，提前定位问题
var _ io.Closer = (*Client)(nil)

//NewClient 新建客户端
func NewClient(conn net.Conn, opt *Option) (*Client, error) {
	f, ok := codec.NewCodecFuncMap[opt.CodecType]
	if !ok {
		err := fmt.Errorf("invalid codec type %s", opt.CodecType)
		log.Println("rpc client: codec error:", err)
		return nil, err
	}
	if err := json.NewEncoder(conn).Encode(opt); err != nil {
		log.Println("rpc client: options error: ", err)
		_ = conn.Close()
		return nil, err
	}
	return newClientCodec(f(conn), opt), nil
}

//newClientCodec 新建客户端编码器
func newClientCodec(cc codec.Codec, opt *Option) *Client {
	client := &Client{
		seq:     1, //初始seq=1，错误或无效seq为0
		opt:     opt,
		cc:      cc,
		pending: make(map[uint64]*Call),
	}
	//开启异步处理请求
	go client.receive()
	return client
}

var ErrShutDown = errors.New("connection was shut down")

//Close 关闭资源
func (client *Client) Close() error {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing {
		return ErrShutDown
	}
	client.closing = true
	return client.cc.Close()
}

//IsAvailable 客户端是否可用
func (client *Client) IsAvailable() bool {
	client.mu.Lock()
	defer client.mu.Unlock()
	return !client.closing && !client.shutdown
}

//registerCall 注册调用
func (client *Client) registerCall(call *Call) (uint64, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.closing || client.shutdown {
		return 0, ErrShutDown
	}
	call.Seq = client.seq
	client.pending[call.Seq] = call
	client.seq++
	return call.Seq, nil
}

//removeCall 移除调用
func (client *Client) removeCall(seq uint64) *Call {
	client.mu.Lock()
	defer client.mu.Unlock()
	call := client.pending[seq]
	delete(client.pending, seq)
	return call
}

//terminateCalls 终止调用
func (client *Client) terminateCalls(err error) {
	client.sending.Lock()
	defer client.sending.Unlock()
	client.mu.Lock()
	defer client.mu.Unlock()
	if err != nil {
		client.shutdown = true
		for _, call := range client.pending {
			call.Error = err
			call.done()
		}
	}
}

func (client *Client) receive() {
	var err error
	var h codec.Header
	for err == nil {
		h = codec.Header{}
		err = client.cc.ReadHeader(&h)
		log.Printf("rpc client header:%+v\n", h)
		if err != nil {
			break
		}
		call := client.removeCall(h.Seq)
		log.Printf("rpc client call:%+v\n", call)
		switch {
		case call == nil:
			log.Printf("rpc client call == nil header:%+v\n", h)
			err = client.cc.ReadBody(nil)
		case h.Error != "":
			call.Error = fmt.Errorf(h.Error)
			err = client.cc.ReadBody(nil)
			call.done()
		default:
			err = client.cc.ReadBody(call.reply)
			if err != nil {
				call.Error = errors.New("reading body " + err.Error())
			}
			call.done()
		}
	}
	client.terminateCalls(err)
}

func parseOption(opts ...*Option) (*Option, error) {
	if len(opts) == 0 || opts[0] == nil {
		return DefaultOption, nil
	}
	if len(opts) > 1 {
		return nil, errors.New("length of options is exceed,must be 1 or less")
	}
	opt := opts[0]
	opt.MagicNumber = DefaultOption.MagicNumber
	if _, ok := codec.NewCodecFuncMap[opt.CodecType]; !ok {
		opt.CodecType = DefaultOption.CodecType
	}
	return opt, nil
}

type newClientFunc func(conn net.Conn, opt *Option) (client *Client, err error)

type clientRes struct {
	client *Client
	err    error
}

func dialTimeout(function newClientFunc, network, address string, opts ...*Option) (client *Client, err error) {
	opt, err := parseOption(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout(network, address, opt.ConnectTimeout)
	if err != nil {
		return nil, err
	}
	defer func() {
		if client == nil {
			_ = conn.Close()
		}
	}()
	ch := make(chan clientRes)
	go func() {
		client, err := function(conn, opt)
		ch <- clientRes{
			client: client,
			err:    err,
		}
	}()
	if opt.ConnectTimeout == 0 {
		res := <-ch
		return res.client, res.err
	}
	select {
	case <-time.After(opt.ConnectTimeout):
		return nil, fmt.Errorf("rpc client: connect timeout: expect within %s", opt.ConnectTimeout)
	case res := <-ch:
		return res.client, res.err
	}
}

func Dial(network, address string, opts ...*Option) (client *Client, err error) {
	return dialTimeout(NewClient, network, address, opts...)
}

func (client *Client) send(call *Call) {
	client.sending.Lock()
	defer client.sending.Unlock()

	//注册调用，seq++
	seq, err := client.registerCall(call)
	if err != nil {
		call.Error = err
		call.done()
		return
	}
	client.header.ServiceMethod = call.ServiceMethod
	client.header.Seq = seq
	client.header.Error = ""

	if err := client.cc.Write(&client.header, call.Args); err != nil {
		call := client.removeCall(seq)
		if call != nil {
			call.Error = err
			call.done()
		}
	}
}

//异步
func (client *Client) Go(serviceMethod string, args, replay interface{}, done chan *Call) *Call {
	if done == nil {
		done = make(chan *Call, 10)
	} else if cap(done) == 0 {
		log.Panic("rpc client: done channel is unbuffered")
	}

	call := &Call{
		ServiceMethod: serviceMethod,
		Args:          args,
		reply:         replay,
		Done:          done,
	}

	client.send(call)
	return call
}

//一个同步调用
func (client *Client) Call(ctx context.Context, serviceMethod string, args, replay interface{}) error {
	//阻塞直到从chan接收返回值
	call := client.Go(serviceMethod, args, replay, make(chan *Call, 1))
	select {
	case <-ctx.Done():
		client.removeCall(call.Seq)
		return errors.New("rpc client: call failed: " + ctx.Err().Error())
	case call := <-call.Done:
		return call.Error
	}
}

func NewHttpClient(conn net.Conn, opt *Option) (*Client, error) {
	_, _ = io.WriteString(conn, fmt.Sprintf("CONNECT %s HTTP/1.0\n\n", defaultRPCPath))
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: http.MethodConnect})
	if err == nil && resp.Status == connected {
		return NewClient(conn, opt)
	}
	if err == nil {
		err = errors.New("unexpected HTTP response: " + resp.Status)
	}
	return nil, err
}

func DialHTTP(network, address string, opts ...*Option) (*Client, error) {
	return dialTimeout(NewHttpClient, network, address, opts...)
}

//XDial protocol@addr
func XDial(rpcAddr string, opts ...*Option) (*Client, error) {
	parts := strings.Split(rpcAddr, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("rpc client err: wrong format '%s', expect protocol@addr", rpcAddr)
	}
	proto, addr := parts[0], parts[1]
	switch proto {
	case "http":
		return DialHTTP("tcp", addr, opts...)
	default:
		return Dial(proto, addr, opts...)
	}
}
