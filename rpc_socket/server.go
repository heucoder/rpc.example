package rpc_socket

import (
	"bufio"
	"fmt"
	"net"
)

func socket(t int) error {
	if t == 1 {
		return RunServer()
	}
	if t == 2 {
		return RunClient()
	}
	fmt.Printf("t:%v is illeagl\n", t)
	return nil
}

func RunServer() error {
	listen, err := net.Listen("tcp", "127.0.0.1:18888")
	if err != nil {
		fmt.Println("listen failed, err:", err)
		return err
	}
	for {
		conn, err := listen.Accept() // 建立连接
		if err != nil {
			fmt.Println("accept failed, err:", err)
			continue
		}
		go process(conn) // 启动一个goroutine处理连接
	}
}

func process(conn net.Conn) {
	defer conn.Close() // 关闭连接
	for {
		reader := bufio.NewReader(conn)
		var buf [128]byte
		n, err := reader.Read(buf[:]) // 读取数据
		if err != nil {
			fmt.Println("read from client failed, err:", err)
			break
		}
		recvStr := string(buf[:n])
		fmt.Println("收到client端发来的数据：", recvStr)
		recvStr += "server:"
		conn.Write([]byte(recvStr)) // 发送数据
	}
}
