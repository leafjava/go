package main

import (
	"fmt"
	"runtime"
	"time"
)

func leak1() {
	ch := make(chan int)

	go func() {
		ch <- 42
		fmt.Println("永远到不了这里")
	}()
}

func main() {
	fmt.Println("启动前goroutine数：", runtime.NumGoroutine())
	for i := 0; i < 100; i++ {
		leak1()
	}
	time.Sleep(time.Second)
	fmt.Println("启动后goroutine数:", runtime.NumGoroutine())
}
