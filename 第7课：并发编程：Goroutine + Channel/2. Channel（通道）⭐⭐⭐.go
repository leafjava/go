package main

//func main() {
//	// 创建 channel
//	ch := make(chan int)
//
//	// 发送数据（在 goroutine 中）
//	go func() {
//		ch <- 42 // 发送
//	}()
//
//	// 接收数据
//	value := <-ch
//	fmt.Println("接收到:", value)
//
//	//带缓冲的channel
//	bufferedCh := make(chan string, 3)
//	bufferedCh <- "A"
//	bufferedCh <- "B"
//	bufferedCh <- "C"
//
//	fmt.Println(<-bufferedCh)
//	fmt.Println(<-bufferedCh)
//	fmt.Println(<-bufferedCh)
//}

func main() {
	ch := make(chan int)

	go func() {
		ch <- 42
	}()

}
