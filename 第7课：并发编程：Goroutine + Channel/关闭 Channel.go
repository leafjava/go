package main

import "fmt"

func main() {
	ch := make(chan int, 3)

	// 发送数据
	ch <- 1
	ch <- 2
	ch <- 3
	close(ch) // 关闭 channel

	// 接收所有数据
	for value := range ch {
		fmt.Println(value)
	}

	// 检查 channel 是否关闭
	_, ok := <-ch
	if !ok {
		fmt.Println("channel closed")
	}
}
