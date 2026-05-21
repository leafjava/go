package main

import (
	"errors"
	"fmt"
	"strings"
)

// TODO: 实现交易验证函数
// 返回：是否有效、错误信息
func validateTransaction(from, to string, amount float64) (bool, error) {
	// 1. 验证发送地址（42字符，0x开头）
	if len(from) != 42 {
		return false, errors.New("发送地址长度必须为42个字符")
	}
	if !strings.HasPrefix(from, "0x") {
		return false, errors.New("发送地址必须以0x开头")
	}

	// 2. 验证接收地址
	if len(to) != 42 {
		return false, errors.New("接收地址长度必须为42个字符")
	}
	if !strings.HasPrefix(to, "0x") {
		return false, errors.New("接收地址必须以0x开头")
	}

	// 3. 验证金额（必须大于0）
	if amount <= 0 {
		return false, errors.New("金额必须大于0")
	}

	// 如果全部通过，返回 true, nil
	return true, nil
}

// TODO: 实现 Gas 估算函数
// 返回：预估 Gas、ETH 费用、USD 费用、错误
func estimateGas(txType string, dataSize int) (gasLimit int64, ethCost float64, usdCost float64, err error) {
	// txType: "transfer" (21000 gas), "contract" (50000 + dataSize*68 gas)
	if txType == "transfer" {
		gasLimit = 21000
	} else if txType == "contract" {
		gasLimit = 50000 + int64(dataSize)*68
	} else {
		err = errors.New("无效的交易类型")
		return
	}

	// gasPrice: 50 Gwei
	gasPrice := 50.0
	ethCost = float64(gasLimit) * gasPrice / 1e9

	// ethPrice: $2000
	ethPrice := 2000.0
	usdCost = ethCost * ethPrice

	return
}

func main() {
	// 测试交易验证
	valid, err := validateTransaction(
		"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
		"0x8ba1f109551bD432803012645Ac136ddd64DBA72",
		1.5,
	)

	if err != nil {
		fmt.Println("验证失败:", err)
	} else if valid {
		fmt.Println("✓ 交易有效")
	}

	// 测试 Gas 估算
	gas, eth, usd, err := estimateGas("transfer", 0)
	if err != nil {
		fmt.Println("估算失败:", err)
	} else {
		fmt.Printf("Gas: %d, ETH: %.6f, USD: $%.2f\n", gas, eth, usd)
	}
}
