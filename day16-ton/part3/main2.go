package main

import (
	"context"
	"fmt"
	"log"

	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
)

//### 使用 HTTP API（更简单的方式）

func main() {
	client := liteclient.NewConnectionPool()

	configURL := "https://ton.org/global.config.json"
	err := client.AddConnectionsFromConfigUrl(context.Background(), configURL)
	if err != nil {
		log.Fatal("连接 TON 节点失败:", err)
	}

	api := ton.NewAPIClient(client)

	info, err := api.GetMasterchainInfo(context.Background())
	if err != nil {
		log.Fatal("获取主链信息失败:", err)
	}

	fmt.Printf("最新区块 Seqno: %d\n", info.SeqNo)
	fmt.Printf("最新区块 Hash: %x\n", info.RootHash)
}
