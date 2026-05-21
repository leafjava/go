package main

import "fmt"

func main() {
	// Gas 参数
	gasLimit := 21000  // Gas 限制
	gasPrice := 50.0   // Gwei
	ethPrice := 2000.0 // USD

	// TODO: 计算以下内容
	// 1. Gas 费用（ETH）= gasLimit * gasPrice / 1e9
	gasFeeETH := float64(gasLimit) * gasPrice / 1e9
	// 2. Gas 费用（USD）= Gas费用(ETH) * ethPrice
	gasFeeUSD := gasFeeETH * ethPrice
	// 3. 输出结果，保留4位小数

	// 示例输出：
	// Gas Limit: 21000
	// Gas Price: 50.0 Gwei
	// Gas 费用: 0.0011 ETH
	// Gas 费用: $2.1000
	fmt.Println("Gas Limit:", gasLimit)
	fmt.Println("Gas Price:", gasPrice, "Gwei")
	fmt.Printf("Gas 费用: %.4f ETH\n", gasFeeETH)
	fmt.Printf("Gas 费用: $%.4f\n", gasFeeUSD)
}
