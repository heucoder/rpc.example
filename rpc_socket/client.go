package rpc_socket

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func RunClient() error {
	conn, err := net.Dial("tcp", "127.0.0.1:18888")
	if err != nil {
		fmt.Println("err :", err)
		return err
	}
	defer conn.Close() // 关闭连接
	inputReader := bufio.NewReader(os.Stdin)
	for {
		input, _ := inputReader.ReadString('\n') // 读取用户输入
		inputInfo := strings.Trim(input, "\r\n")
		if strings.ToUpper(inputInfo) == "Q" { // 如果输入q就退出
			return nil
		}
		_, err = conn.Write([]byte(inputInfo)) // 发送数据
		if err != nil {
			return err
		}
		buf := [512]byte{}
		n, err := conn.Read(buf[:])
		if err != nil {
			fmt.Println("recv failed, err:", err)
			return err
		}
		fmt.Println(string(buf[:n]))
	}
}
