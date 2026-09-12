package main

import "fmt"

func main() {
	gasLimit := 21000
	gasPrice := 50.0
	ethPrice := 2000.0

	gasFeeETH := float64(gasLimit) * gasPrice / 1e9
	gasFeeUSD := gasFeeETH * ethPrice

	fmt.Println("Gas Limit:", gasLimit)
	fmt.Println("Gas Price:", gasPrice, "Gwei")
	fmt.Printf("Gas 费用: %.4f ETH\n", gasFeeETH)
	fmt.Printf("Gas 费用: $%.4f\n", gasFeeUSD)
}
