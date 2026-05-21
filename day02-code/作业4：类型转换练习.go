package main

import (
	"fmt"
	"strconv"
)

func main() {
	// 场景1：用户输入的金额（字符串）转换为数字
	amountStr := "1000.50"
	// TODO: 转换为 float64 并计算手续费（0.1%）
	amount, _ := strconv.ParseFloat(amountStr, 64)
	fee := amount * 0.001
	fmt.Printf("金额: %.2f, 手续费: %.2f\n", amount, fee)

	// 场景2：区块高度（int64）转换为字符串
	blockHeight := int64(18500000)
	// TODO: 转换为字符串并输出
	blockHeightStr := strconv.FormatInt(blockHeight, 10)
	fmt.Println("区块高度:", blockHeightStr)

	// 场景3：Gas Price（float64）转换为 Wei（int64）
	gasPriceGwei := 50.5
	// TODO: 转换为 Wei（1 Gwei = 1e9 Wei）
	gasPriceWei := int64(gasPriceGwei * 1e9)
	fmt.Printf("Gas Price: %.1f Gwei = %d Wei\n", gasPriceGwei, gasPriceWei)

	// 场景4：十六进制地址转换
	hexStr := "0x1a2b3c"
	// TODO: 使用 strconv.ParseInt 转换为十进制
	decimalVal, _ := strconv.ParseInt(hexStr[2:], 16, 64)
	fmt.Printf("十六进制: %s = 十进制: %d\n", hexStr, decimalVal)
}
