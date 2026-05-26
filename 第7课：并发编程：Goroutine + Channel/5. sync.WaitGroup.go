package main

import (
	"fmt"
	"sync"
	"time"
)

func processTransaction(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Printf("处理交易 %d\n", id)
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("交易 %d 完成\n", id)
}

func main() {
	var wg sync.WaitGroup

	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go processTransaction(i, &wg)
	}

	wg.Wait()
	fmt.Println("所有交易处理完成")
}
