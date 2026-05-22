package main

import (
	"fmt"
	"time"
)

func main() {
	// 匿名函数 + goroutine
	go func() {
		fmt.Println("匿名 goroutine 执行")
	}()

	// 带参数的匿名 goroutine
	name := "leaf"
	go func(n string) {
		fmt.Println("Hello", n)
	}(name)

	time.Sleep(100 * time.Millisecond)
}
