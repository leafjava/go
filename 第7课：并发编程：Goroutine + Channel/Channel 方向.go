package main

import "fmt"

// 只发送 channel
func sendOnly(ch chan<- int) {
	ch <- 100
}

// 只接收 channel
func receiveOnly(ch <-chan int) {
	value := <-ch
	fmt.Println("接收:", value)
}

func main() {
	ch := make(chan int)

	go sendOnly(ch)
	receiveOnly(ch)
}
