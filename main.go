package main

import (
	"log"
	"net"
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
	server := core.NewServer()
	_ = server.Register(&f)
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	server.Accept(l)
}
func foo(xc *core.XClient, typ, serviceMethod string, args *Args) {
	var reply int
	var err error

	switch typ {
	case "call":
		err = xc.Call(serviceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: %d + %d = %d", typ, serviceMethod, args.Num1, args.Num2, reply)
	}
}

func call(addr1, addr2 string) {
	d := core.NewMulitServerDisCovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := core.NewXClient(d, core.RandomSelect)
	defer func() {
		_ = xc.Close()
	}()

	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			foo(xc, "call", "Foo.Sum", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func broadCast(addr1, addr2 string) {
	d := core.NewMulitServerDisCovery([]string{"tcp@" + addr1, "tcp@" + addr2})
	xc := core.NewXClient(d, core.RandomSelect)
	defer func() { _ = xc.Close() }()
	wg := sync.WaitGroup{}
	for i := 0; i < 1; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			log.Println("go foo", i)
			foo(xc, "broadcast", "Foo.Sum", &Args{Num1: i, Num2: i * i})
		}(i)
	}
	wg.Wait()
}

func main() {
	ch1 := make(chan string)
	ch2 := make(chan string)

	go ServerRun(ch1)
	go ServerRun(ch2)

	addr1 := <-ch1
	addr2 := <-ch2

	time.Sleep(time.Second)

	call(addr1, addr2)
	// broadCast(addr1, addr2)

	log.Println("done")
}

// func add(i, j int) {
// 	log.Println(j + i)
// }

// func call(methodName string, arg1, arg2 interface{}) {
// 	var f func(j, i int)
// 	switch methodName {
// 	case "add":
// 		f = add
// 	}
// 	fv := reflect.ValueOf(f)
// 	param1 := reflect.ValueOf(arg1)
// 	param2 := reflect.ValueOf(arg2)
// 	fv.Call([]reflect.Value{param1, param2})
// }
