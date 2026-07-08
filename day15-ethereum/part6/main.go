package main

import (
	"context"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("wss://eth.llamarpc.com")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 合约地址
	contractAddress := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")

	// 创建过滤器
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}

	// 订阅日志
	logs := make(chan types.Log)
	sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("开始监听事件...")

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case vLog := <-logs:
			fmt.Printf("区块: %d, 交易: %s\n", vLog.BlockNumber, vLog.TxHash.Hex())
		}
	}
}
