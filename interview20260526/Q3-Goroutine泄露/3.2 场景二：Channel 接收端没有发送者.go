package main

import "fmt"

func watchTransaction(txHash string) {
	confirmCh := make(chan int)

	go func() {
		// 想从某个 RPC 调用拿确认数
		confirmations, err := callRPC(txHash)
		if err != nil {
			return // ⚠️ 这里直接 return 了，confirmCh 永远不会有数据
		}
		confirmCh <- confirmations
	}()

	// 主流程在等 confirmCh
	confirms := <-confirmCh // 上面 return 了，这里永远阻塞
	fmt.Println("确认数:", confirms)
}

func main() {
	watchTransaction("123")
}
