package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	count := int32(0)

	f := func(i int, wg *sync.WaitGroup) {
		time.Sleep(1 * time.Second)
		fmt.Println("this is ", i)
		atomic.AddInt32(&count, 1)
		wg.Done()
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go f(i, &wg)
	}

	wg.Wait()
	fmt.Println("yes, count=", atomic.LoadInt32(&count))
}
