package core

import (
	"io"
	"log"
	"reflect"
	"sync"

	"rpc.example/core/codec"
)

type XClient struct {
	mu        sync.Mutex
	clientMap map[string]*Client
	d         DisCovery
	mode      SelectMode
	opt       *codec.Option
}

func NewXClient(d DisCovery, mode SelectMode) *XClient {
	return &XClient{
		d:         d,
		clientMap: map[string]*Client{},
		mode:      mode,
	}
}

var _ io.Closer = (*XClient)(nil)

func (x *XClient) Close() error {
	x.mu.Lock()
	defer x.mu.Unlock()

	for key, client := range x.clientMap {
		log.Println("Xclient close")
		_ = client.Close()
		delete(x.clientMap, key)
	}

	return nil
}

func (x *XClient) call(addr string, serviceMethod string, args, reply interface{}) error {
	client, err := x.Dial(addr)
	if err != nil {
		return err
	}
	return client.Call(serviceMethod, args, reply)
}

func (x *XClient) Call(serviceMethod string, args, reply interface{}) error {
	addr, err := x.d.Get(x.mode)
	if err != nil {
		return err
	}
	return x.call(addr, serviceMethod, args, reply)
}

func (x *XClient) Dial(rpcAddr string) (*Client, error) {
	x.mu.Lock()
	defer x.mu.Unlock()
	client, ok := x.clientMap[rpcAddr]
	if ok && !client.IsAvailable() {
		_ = client.Close()
		delete(x.clientMap, rpcAddr)
		client = nil
	}
	if client == nil {
		var err error
		client, err = XDial(rpcAddr, x.opt)
		if err != nil {
			return nil, err
		}
		x.clientMap[rpcAddr] = client
	}
	return client, nil
}

func (x *XClient) Broadcast(serviceMethod string, args, reply interface{}) error {
	addrs, err := x.d.GetAll()
	if err != nil {
		return err
	}

	// wg := sync.WaitGroup{}
	var mu sync.Mutex
	var e error
	replyDone := reply == nil

	for _, addr := range addrs {
		// wg.Add(1)
		// go func(rpcaddr string) {
		// 	defer wg.Done()
		// 	var replyClone interface{}
		// 	if reply != nil {
		// 		replyClone = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
		// 	}
		// 	err := x.call(rpcaddr, serviceMethod, args, replyClone)
		// 	mu.Lock()
		// 	if !replyDone && err != nil && e == nil {
		// 		e = err
		// 	}
		// 	if err == nil && !replyDone {
		// 		replyDone = true
		// 		reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(replyClone).Elem())
		// 	}
		// 	mu.Unlock()
		// }(addr)
		var replyClone interface{}
		if reply != nil {
			replyClone = reflect.New(reflect.ValueOf(reply).Elem().Type()).Interface()
		}
		err := x.call(addr, serviceMethod, args, replyClone)
		mu.Lock()
		if !replyDone && err != nil && e == nil {
			e = err
		}
		if err == nil && !replyDone {
			replyDone = true
			reflect.ValueOf(reply).Elem().Set(reflect.ValueOf(replyClone).Elem())
		}
		mu.Unlock()
	}

	// wg.Wait()
	return e
}
