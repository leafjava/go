package main

import (
	"fmt"
	"log"

	"part7/services"
)

func main() {
	// 创建服务
	svc, err := services.NewEthereumService("https://ethereum.publicnode.com")
	if err != nil {
		log.Fatal(err)
	}
	defer svc.Close()

	// 查询 ETH 余额
	address := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
	balance, err := svc.GetBalance(address)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("地址: %s\n", address)
	fmt.Printf("ETH 余额: %s wei\n", balance.String())
}
