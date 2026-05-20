package main

import "fmt"

func isValidEthAddress(address string) bool {
	return false
}

func main() {
	addresses := []string{
		"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
		"0xInvalidAddress",
		"742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
	}

	for _, addr := range addresses {
		if isValidEthAddress(addr) {
			fmt.Printf("Address %s is valid\n", addr)
		} else {
			fmt.Printf("Address %s is invalid\n", addr)
		}
	}
}
