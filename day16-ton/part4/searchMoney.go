package main

import (
	"context"
	"fmt"
	"log"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
)

//### 查询余额

func main() {
	client := liteclient.NewConnectionPool()
	err := client.AddConnectionsFromConfigUrl(context.Background(), "https://ton.org/global.config.json")
	if err != nil {
		log.Fatal(err)
	}

	api := ton.NewAPIClient(client)

	// 解析地址
	addr := address.MustParseAddr("EQA0gvftODUs9itSl5whd01IxERToRZkA_-tsriaZPxuRkS4")

	// 获取主链最新区块
	block, err := api.GetMasterchainInfo(context.Background())
	if err != nil {
		log.Fatal("获取区块信息失败:", err)
	}

	// 获取账户状态
	account, err := api.GetAccount(context.Background(), block, addr)
	if err != nil {
		log.Fatal("查询账户失败:", err)
	}

	fmt.Printf("地址: %s\n", addr.String())

	if !account.IsActive {
		fmt.Println("余额: 0 TON（账户未激活）")
		fmt.Println("状态: 未激活（non-existent）")
	} else {
		balanceTON := float64(account.State.Balance.Nano().Uint64()) / 1e9
		fmt.Printf("余额: %.9f TON\n", balanceTON)
		fmt.Printf("状态: %s\n", account.State.Status)
	}
}
