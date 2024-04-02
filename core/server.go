package core

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net"
	"reflect"
	"strings"
	"sync"

	"rpc.example/core/codec"
)

// request stores all information of a call
type request struct {
	h            *codec.Header // header of request
	argv, replyv reflect.Value // argv and replyv of request
	mtype        *methodType
	svc          *service
}

// Server represents an RPC Server.
type Server struct {
	serviceMap sync.Map
}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{}
}

// DefaultServer is the default instance of *Server.
var DefaultServer = NewServer()

func Accept(lis net.Listener) {
	DefaultServer.Accept(lis)
}

func Register(rcvr interface{}) error { return DefaultServer.Register(rcvr) }

func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error: \n", err)
			return
		}
		go server.ServeConn(conn)
	}
}

func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	//解析协议

	defer func() { _ = conn.Close() }()

	opt := &codec.Option{}
	err := json.NewDecoder(conn).Decode(opt)
	if err != nil {
		log.Println("rpc server: ServeConn error Decode: \n", err)
		return
	}
	if opt.MagicNumber != codec.MagicNumber {
		log.Println("rpc server: ServeConn  opt.MagicNumber:\n", opt.MagicNumber)
		return
	}
	if opt.CodecType != codec.JsonType {
		log.Println("rpc server: ServeConn  codec.JsonType: \n", codec.JsonType)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	//构造codec
	server.serveCodec(f(conn))
}

// invalidRequest is a placeholder for response argv when error occurs
var invalidRequest = struct{}{}

func (server *Server) serveCodec(cc codec.Codec) {

	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)

	for {
		//读requsest
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				log.Println("req is nil") //为空符合预期

				break // it's not possible to recover, so close the connection
			}
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		server.handleRequest(cc, req, sending, wg)
	}

	wg.Wait()
	log.Println("Server close")
	_ = cc.Close()
}

func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	header := &codec.Header{}
	err := cc.ReadHeader(header)
	if err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("readRequestHeader err: \n", err)
		}
		return nil, err
	}
	return header, nil
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	log.Println("begin server.readRequest")
	header, err := server.readRequestHeader(cc)
	if err != nil {
		log.Println("server.readRequest readRequestHeader err", err)
		return nil, err
	}
	log.Println("server.readRequest 1")
	req := &request{
		h: header,
	}
	req.svc, req.mtype, err = server.findService(header.ServiceMethod)
	if err != nil {
		log.Println("readRequest findService err: \n", err)
		return req, err
	}
	req.argv = req.mtype.newArgv()
	req.replyv = req.mtype.newReplyv()

	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Ptr {
		argvi = req.argv.Addr().Interface()
	}
	if err := cc.ReadBody(argvi); err != nil {
		log.Println("readRequest ReadBody err: \n", err)
		return nil, err
	}
	log.Println("end server.readRequest req", req)
	return req, nil
}

func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	err := cc.Write(h, body)
	if err != nil {
		log.Println("sendResponse Write err: \n", err)
		return
	}
	return
}

func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	err := req.svc.call(req.mtype, req.argv, req.replyv)
	if err != nil {
		req.h.Error = err.Error()
		server.sendResponse(cc, req.h, invalidRequest, sending)
		return
	}
	server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
}

func (server *Server) Register(rcvr interface{}) error {
	s := newService(rcvr)
	if _, dup := server.serviceMap.LoadOrStore(s.name, s); dup {
		return errors.New("rpc: service already defined: " + s.name)
	}
	return nil
}

func (server *Server) findService(serviceMethod string) (svc *service, mtype *methodType, err error) {
	dot := strings.Split(serviceMethod, ".")
	if len(dot) != 2 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := dot[0], dot[1]
	svcAny, ok := server.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}
	svc = svcAny.(*service)
	mtype, ok = svc.method[methodName]
	if !ok {
		err = errors.New("rpc server: can't find method " + methodName)
	}
	return
}
