package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	// 连接以太坊节点
	client, err := ethclient.Dial("https://ethereum-rpc.publicnode.com")
	if err != nil {
		log.Fatal("连接失败:", err)
	}
	defer client.Close()

	fmt.Println("连接成功")

	// 获取最新区块号
	blockNumber, err := client.BlockNumber(context.Background())
	if err != nil {
		log.Fatal("获取区块号失败:", err)
	}
	fmt.Printf("最新区块: %d\n", blockNumber)
}
