package main

import "fmt"

func main() {
	// 地址 -> 余额
	balances := make(map[string]float64)

	// 添加余额
	balances["0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"] = 10.5
	balances["0x8ba1f109551bD432803012645Ac136ddd64DBA72"] = 5.2

	// 查询余额
	addr := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"

	if balance, exists := balances[addr]; exists {
		fmt.Printf("地址 %s 余额: %.2f ETH\n", addr, balance)
	}

	// 转账
	from := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
	to := "0x8ba1f109551bD432803012645Ac136ddd64DBA72"
	amount := 2.0

	if balances[from] >= amount {
		balances[from] -= amount
		balances[to] += amount
		fmt.Println("转账成功")
	}

	// 统计总余额
	total := 0.0
	for _, balance := range balances {
		total += balance
	}
	fmt.Printf("总余额: %.2f ETH\n", total)
}
