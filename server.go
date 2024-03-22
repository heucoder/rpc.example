package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"sync"

	"rpc.example/codec"
)

// request stores all information of a call
type request struct {
	h            *codec.Header // header of request
	argv, replyv reflect.Value // argv and replyv of request
}

// Server represents an RPC Server.
type Server struct{}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{}
}

// DefaultServer is the default instance of *Server.
var DefaultServer = NewServer()

func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:%v \n", err)
			return
		}
		go server.ServeConn(conn)
	}
}

func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	//解析协议
	opt := &codec.Option{}
	err := json.NewDecoder(conn).Decode(opt)
	if err != nil {
		log.Println("rpc server: ServeConn error Decode:%v \n", err)
		return
	}
	if opt.MagicNumber != codec.MagicNumber {
		log.Println("rpc server: ServeConn  opt.MagicNumber:%v \n", opt.MagicNumber)
		return
	}
	if opt.CodecType != codec.JsonType {
		log.Println("rpc server: ServeConn  codec.JsonType:%v \n", codec.JsonType)
		return
	}
	//构造codec
	server.serveCodec(codec.NewJsonCodec(conn))
}

// invalidRequest is a placeholder for response argv when error occurs
var invalidRequest = struct{}{}

func (server *Server) serveCodec(cc codec.Codec) {

	sending := new(sync.Mutex)
	wg := sync.WaitGroup{}

	for {
		//读requsest
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				log.Println("req is nil")

				break // it's not possible to recover, so close the connection
			}
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		server.handleRequest(cc, req, sending, &wg)
	}

	wg.Wait()
	_ = cc.Close()
}

func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	header := &codec.Header{}
	err := cc.ReadHeader(header)
	if err != nil {
		log.Println("readRequestHeader err:%v \n", err)
		return nil, err
	}
	return header, nil
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	header, err := server.readRequestHeader(cc)
	if err != nil {
		log.Println("readRequest readRequestHeader err:%v \n", err)
		return nil, err
	}
	argv := reflect.New(reflect.TypeOf(""))
	if err := cc.ReadBody(argv.Interface()); err != nil {
		log.Println("readRequest ReadBody err:%v \n", err)
		return nil, err
	}
	return &request{
		h:    header,
		argv: argv,
	}, nil
}

func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	err := cc.Write(h, body)
	if err != nil {
		log.Println("sendResponse Write err:%v \n", err)
		return
	}
	return
}

func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println(req.h, req.argv.Elem())
	req.replyv = reflect.ValueOf(fmt.Sprintf("geerpc resp %d", req.h.Seq))
	server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}
