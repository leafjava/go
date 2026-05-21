package main

import (
	"errors"
	"fmt"
	"strings"
)

// 验证以太坊地址并返回格式化地址
func validateEthAddress(address string) (string, error) {
	// 检查长度
	if len(address) != 42 {
		return "", errors.New("地址长度必须为42个字符")
	}

	// 检查前缀
	if !strings.HasPrefix(address, "0x") {
		return "", errors.New("地址必须以0x开头")
	}

	// 转换为小写（标准格式）
	normalized := strings.ToLower(address)
	return normalized, nil
}

// 计算 Gas 费用
func calculateGasFee(gasLimit int64, gasPriceGwei float64) (ethCost float64, usdCost float64, err error) {
	if gasLimit <= 0 {
		err = errors.New("Gas Limit 必须大于0")
		return
	}

	if gasPriceGwei <= 0 {
		err = errors.New("Gas Price 必须大于0")
		return
	}

	// 计算 ETH 费用
	ethCost = float64(gasLimit) * gasPriceGwei / 1e9

	// 假设 ETH 价格 $2000
	ethPrice := 2000.0
	usdCost = ethCost * ethPrice

	return
}

func main() {
	// 测试地址验证
	addr, err := validateEthAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	if err != nil {
		fmt.Println("地址验证失败:", err)
	} else {
		fmt.Println("有效地址:", addr)
	}

	// 测试 Gas 计算
	ethCost, usdCost, err := calculateGasFee(21000, 50)
	if err != nil {
		fmt.Println("计算失败:", err)
	} else {
		fmt.Printf("Gas 费用: %.6f ETH ($%.2f)\n", ethCost, usdCost)
	}
}
