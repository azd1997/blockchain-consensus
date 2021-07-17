package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	server, err := net.Listen("tcp", "127.0.0.1:9000")
	if err != nil {
		panic(err)//会退出程序
	}

	for {
		conn, err := server.Accept()	// 阻塞
		if err != nil {
			continue
		}

		go func() {
			res := make([]byte, 1024)
			n, err := conn.Read(res)
			fmt.Println("接收到数据：", string(res[:n]))
			fmt.Println("接收到数据：", res[:n])

			// 回发消息
			_, err = conn.Write([]byte("我是服务端"))
			if err != nil {
				panic(err)
			}

			time.Sleep(10 * time.Second)
		}()

	}
}
// 如果有多个client，现在的写法会导致什么问题？怎么解决？