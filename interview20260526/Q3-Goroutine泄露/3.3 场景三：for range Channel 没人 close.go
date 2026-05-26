package main

import (
	"fmt"
	"time"
)

func workerLeak() {
	jobs := make(chan int, 5)

	// 生产者
	go func() {
		for i := 0; i < 5; i++ {
			jobs <- i
		}
		// ⚠️ 忘了 close(jobs)
	}()

	// 消费者
	go func() {
		for job := range jobs {
			// for range 在 channel close 之前永远不退出
			fmt.Println("处理任务:", job)
		}
		// 永远到不了这里
	}()

	time.Sleep(time.Second)
}
