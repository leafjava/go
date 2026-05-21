package main

import "fmt"

func main() {
	// 声明钱包变量
	walletAddress := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
	ethBalance := 1.5
	usdtBalance := 1000.00
	isVerified := true
	txCount := 42

	// 输出钱包信息
	fmt.Println("========== 钱包信息 ==========")
	fmt.Println("地址:", walletAddress)
	fmt.Printf("ETH 余额: %.1f ETH\n", ethBalance)
	fmt.Printf("USDT 余额: %.2f USDT\n", usdtBalance)
	fmt.Println("已验证:", isVerified)
	fmt.Println("交易次数:", txCount)
	fmt.Println("==============================")
}
