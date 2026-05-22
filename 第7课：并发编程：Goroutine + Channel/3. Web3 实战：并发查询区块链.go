package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Balance struct {
	Address string
	Amount  float64
	Error   error
}

// 模拟查询余额（耗时操作）
func queryBalance(address string) (float64, error) {
	time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
	return rand.Float64() * 10, nil
}

// 并发查询
func queryBalancesConcurrent(addresses []string) []Balance {
	results := make(chan Balance, len(addresses))

	// 启动多个 goroutine
	for _, addr := range addresses {
		go func(address string) {
			amount, err := queryBalance(address)
			results <- Balance{
				Address: address,
				Amount:  amount,
				Error:   err,
			}
		}(addr)
	}

	// 收集结果
	balances := make([]Balance, 0, len(addresses))
	for i := 0; i < len(addresses); i++ {
		balances = append(balances, <-results)
	}

	return balances
}

func main() {
	addresses := []string{
		"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
		"0x8ba1f109551bD432803012645Ac136ddd64DBA72",
		"0x1234567890123456789012345678901234567890",
	}

	start := time.Now()
	balances := queryBalancesConcurrent(addresses)
	fmt.Printf("查询完成，耗时: %v\n", time.Since(start))

	for _, b := range balances {
		fmt.Printf("%s: %.2f ETH\n", b.Address, b.Amount)
	}
}
