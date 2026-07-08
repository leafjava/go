package main

import (
	"context"
	"fmt"
	"log"

	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
)

func main() {
	client := liteclient.NewConnectionPool()

	configURL := "https://ton.org/global.config.json"
	err := client.AddConnectionsFromConfigUrl(context.Background(), configURL)
	if err != nil {
		log.Fatal("连接 TON 节点失败:", err)
	}

	api := ton.NewAPIClient(client)

	fmt.Println("TON 节点连接成功")
	_ = api
}
