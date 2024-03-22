package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"reflect"
	"time"

	"rpc.example/codec"
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

	conn, _ := net.Dial("tcp", <-addr)
	defer func() { _ = conn.Close() }()

	time.Sleep(time.Second)

	_ = json.NewEncoder(conn).Encode(codec.DefaultOption)

	cc := codec.NewJsonCodec(conn)
	for i := 0; i < 5; i++ {
		header := codec.Header{
			ServiceMethod: "service.method",
			Seq:           uint64(i),
		}
		body := reflect.ValueOf(fmt.Sprintf("req i:%v", i))
		_ = cc.Write(&header, body.Interface())

		_ = cc.ReadHeader(&header)
		var reply string
		_ = cc.ReadBody(&reply)
		log.Println("reply:", reply)
	}
}
