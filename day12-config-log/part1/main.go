package main

import (
	"fmt"
	"log"
	"part1/config"
)

func main() {
	// 加载配置
	if err := config.LoadConfig("config.yaml"); err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 使用配置
	fmt.Printf("服务器端口: %d\n", config.AppConfig.Server.Port)
	fmt.Printf("数据库: %s\n", config.AppConfig.GetDatabaseDSN())
	fmt.Printf("以太坊 RPC: %s\n", config.AppConfig.Blockchain.Ethereum.RPCURL)
}
