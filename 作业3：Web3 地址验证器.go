package main

import (
	"fmt"
	"strings"
)

// 验证以太坊地址格式（0x开头，42个字符）
func isValidEthAddress(address string) bool {
	// 1. 检查长度是否为 42
	if len(address) != 42 {
		return false
	}

	// 2. 检查是否以 "0x" 开头
	if !strings.HasPrefix(address, "0x") {
		return false
	}

	// 3. 检查是否只包含十六进制字符（0x后面的40个字符）
	hexPart := address[2:] // 去掉 "0x" 前缀
	for _, char := range hexPart {
		// 检查每个字符是否是 0-9, a-f, A-F
		if !((char >= '0' && char <= '9') || 
		     (char >= 'a' && char <= 'f') || 
		     (char >= 'A' && char <= 'F')) {
			return false
		}
	}

	return true
}

func main() {
	// 测试用例
	testAddresses := []string{
		"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",  // 无效：41字符
		"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb0", // 有效：42字符
		"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEbG", // 无效：包含非十六进制字符G
		"742d35Cc6634C0532925a3b844Bc9e7595f0bEb0",   // 无效：没有0x前缀
		"0x0000000000000000000000000000000000000000", // 有效：零地址
	}

	for _, addr := range testAddresses {
		result := isValidEthAddress(addr)
		fmt.Printf("地址: %s\n验证结果: %v\n\n", addr, result)
	}
}
