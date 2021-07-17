package main

import (
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:9000")
	if err != nil {
		panic(err)//会退出程序
	}
	// 给服务端发消息
	_, err = conn.Write([]byte("我是客户端"))
	if err != nil {
		panic(err)//会退出程序
	}

	res := make([]byte, 1024)
	n, err := conn.Read(res)
	fmt.Println("接收到数据：", string(res[:n]))
}
