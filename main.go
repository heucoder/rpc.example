package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

func ServerRun(addr chan string) {
	l, err := net.Listen("tcp", ":17777")
	if err != nil {
		log.Fatal("network error:", err)
	}
	log.Println("start rpc server on", l.Addr())
	addr <- l.Addr().String()
	Accept(l)
}

func main() {
	addr := make(chan string)
	go ServerRun(addr)

	client, _ := Dial("tcp", <-addr)
	defer func() { _ = client.Close() }()
	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var reply string
			err := client.Call("service-method", fmt.Sprintf("req i:%v", i), &reply)
			if err != nil {
				log.Fatalln("main err:", err)
				return
			}
			log.Println("resp:", reply)
		}(i)

	}
	wg.Wait()
}
