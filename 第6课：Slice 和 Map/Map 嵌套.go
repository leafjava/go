package main

import "fmt"

func main() {
	// 用户 -> (代币 -> 余额)
	userBalances := make(map[string]map[string]float64)

	// 初始化用户
	user1 := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
	userBalances[user1] = make(map[string]float64)
	userBalances[user1]["ETH"] = 10.0
	userBalances[user1]["USDT"] = 1000.0
	userBalances[user1]["USDC"] = 500.0

	// 查询
	if tokens, exists := userBalances[user1]; exists {
		for token, balance := range tokens {
			fmt.Printf("%s: %.2f\n", token, balance)
		}
	}
}
