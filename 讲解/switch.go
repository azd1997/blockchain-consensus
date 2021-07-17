package main

import (
	"fmt"
	"math/rand"
)

func testSwitch() {
	names := []string{"tom", "alice", "any"}

	randomName := names[rand.Intn(3) ]

	switch randomName {
	case "tom":
		fmt.Println("我是tom")
	case "alice":
		fmt.Println("我是alice")
	default:
		fmt.Println("我不是tom也不是alice")
	}
}
