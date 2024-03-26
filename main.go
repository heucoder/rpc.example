package main

import (
	"log"
	"net"
	"reflect"
	"sync"
	"time"

	"rpc.example/core"
)

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func ServerRun(addr chan string) {
	var f Foo
	err := core.Register(f)
	if err != nil {
		log.Fatal("Register error:", err)
		panic(err)
	}
	l, err := net.Listen("tcp", ":17777")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	core.Accept(l)
}

func main() {
	addr := make(chan string)
	go ServerRun(addr)

	client, _ := core.Dial("tcp", <-addr)
	defer func() { _ = client.Close() }()
	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var reply int
			args := Args{Num1: i, Num2: i * i}
			err := client.Call("Foo.Sum", args, &reply)
			if err != nil {
				log.Fatalln("main err:", err)
				return
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)

	}
	wg.Wait()
	// call("add", 1, 3)
}

func add(i, j int) {
	log.Println(j + i)
}

func call(methodName string, arg1, arg2 interface{}) {
	var f func(j, i int)
	switch methodName {
	case "add":
		f = add
	}
	fv := reflect.ValueOf(f)
	param1 := reflect.ValueOf(arg1)
	param2 := reflect.ValueOf(arg2)
	fv.Call([]reflect.Value{param1, param2})
}
