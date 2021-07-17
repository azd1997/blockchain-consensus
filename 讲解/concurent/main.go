package main

import (
	"fmt"
	"time"
)

func goRoutine(id int) {
	go func(id int) {
		time.Sleep(5*time.Second)
		fmt.Printf("我是goroutine #%d\n", id)
	}(id)
}

func readGoroutine(c1,c2 chan string) {
	go func(c1,c2 chan string) {
		//for shared == 0 {}
		//fmt.Println(shared)
		for {
			select {
			case msg1 := <-c1:
				fmt.Printf("接收端1：读到数据：%s\n", msg1)
			case msg2 := <-c2:
				fmt.Printf("接收端2：读到数据：%s\n", msg2)
			}
		}
	}(c1,c2)
}


func writeGoroutine1(c chan string) {
	go func(c chan string) {
		//shared++
		c <- "我是发送端1"
		fmt.Printf("发送端1：我已发送数据\n")
	}(c)
}

func writeGoroutine2(c chan string) {
	go func(c chan string) {
		//shared++
		c <- "我是发送端2"
		fmt.Printf("发送端2：我已发送数据\n")
	}(c)
}

//var shared = 0

func main() {
	//for i:=0; i<5; i++ {
	//	goRoutine(i)
	//}
	//
	//time.Sleep(10 * time.Second)

	msgC1 := make(chan string)	// 无缓冲channel，是阻塞的
	// make(chan string, 5)	// 有缓冲的
	msgC2 := make(chan string)


	readGoroutine(msgC1, msgC2)

	writeGoroutine1(msgC1)
	writeGoroutine2(msgC2)

	time.Sleep(1 * time.Second)
}

// CSP 基于 消息传递 而非 共享内存 的goroutine间通信