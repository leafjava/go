package main

import (
	"fmt"
	"time"
)

func queryWithTimeout(address string, timeout time.Duration) (float64, error) {
	result := make(chan float64)

	go func() {
		time.Sleep(500 * time.Millisecond)
		result <- 10.5
	}()

	select {
	case balance := <-result:
		return balance, nil
	case <-time.After(timeout):
		return 0, fmt.Errorf("查询超时")
	}
}

func main() {
	balance, err := queryWithTimeout("0x742d35Cc...", 1*time.Second)
	if err != nil {
		fmt.Println("错误:", err)
	} else {
		fmt.Printf("余额:%.2f ETH\n", balance)
	}

	balance, err = queryWithTimeout("0x742d35Cc...", 1*time.Millisecond)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("余额:%.2f ETH\n", balance)
	}
}
