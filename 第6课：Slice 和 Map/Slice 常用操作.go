package main

import "fmt"

func main() {
	addresses := []string{
		"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
		"0x8ba1f109551bD432803012645Ac136ddd64DBA72",
		"0x1234567890123456789012345678901234567890",
	}

	// 1. 检查是否包含
	target := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
	found := false
	for _, addr := range addresses {
		if addr == target {
			found = true
			break
		}
	}
	fmt.Println("包含地址:", found)

	// 2. 删除元素（通过切片重组）
	indexToRemove := 1
	addresses = append(addresses[:indexToRemove], addresses[indexToRemove+1:]...)
	fmt.Println("删除后:", addresses)

	// 3. 插入元素
	newAddr := "0xNEWADDRESS"
	insertIndex := 1
	addresses = append(addresses[:insertIndex], append([]string{newAddr}, addresses[insertIndex:]...)...)
	fmt.Println("插入后:", addresses)

	// 4. 复制切片
	copied := make([]string, len(addresses))
	copy(copied, addresses)
	fmt.Println("复制:", copied)
}
