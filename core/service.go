package core

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

type methodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	numCalls  uint64 //调用次数
}

func (m *methodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

func (m *methodType) newArgv() reflect.Value {
	var argv reflect.Value
	if m.ArgType.Kind() == reflect.Pointer {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

func (m *methodType) newReplyv() reflect.Value {
	reply := reflect.New(m.ReplyType.Elem())
	if reply.Elem().Kind() == reflect.Array {
		reply.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0)) //?m.ReplyType.Elem()
	} else if reply.Elem().Kind() == reflect.Map {
		reply.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	}
	return reply
}

type service struct {
	name   string
	typ    reflect.Type
	rcvr   reflect.Value
	method map[string]*methodType
}

func newService(rcvr interface{}) *service {
	service := new(service)
	service.rcvr = reflect.ValueOf(rcvr)
	service.typ = reflect.TypeOf(rcvr)
	service.name = reflect.Indirect(service.rcvr).Type().Name()
	if !ast.IsExported(service.name) {
		log.Fatalf("rpc server: %s is not a valid service name", service.name)
	}
	service.registerMethods()
	return service
}

func (s *service) registerMethods() {
	s.method = make(map[string]*methodType, 0)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		if method.Type.NumIn() != 3 || method.Type.NumOut() != 1 {
			continue
		}
		if method.Type.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argType, replyType := method.Type.In(1), method.Type.In(2)
		if !isExportedOrBuiltinType(argType) || !isExportedOrBuiltinType(replyType) {
			continue
		}

		m := &methodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		s.method[method.Name] = m
		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}

func isExportedOrBuiltinType(t reflect.Type) bool { //这个不是特别理解
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}

func (s *service) call(m *methodType, argv, replyv reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)
	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.rcvr, argv, replyv})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}
