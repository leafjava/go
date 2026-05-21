package main

import "fmt"

func counter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

// Web3 实战：创建交易计数器
func createTxCounter(initialCount int) func(txHash string) int {
	count := initialCount

	return func(txHash string) int {
		count++
		fmt.Printf("处理交易 #%d: %s\n", count, txHash)
		return count
	}
}

func main() {
	c1 := counter()
	fmt.Println(c1())
	fmt.Println(c1())
	fmt.Println(c1())

	c2 := counter()
	fmt.Println(c2())

	txCounter := createTxCounter(0)
	txCounter("0xabc123")
}
