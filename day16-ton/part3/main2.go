package main

import (
	"context"
	"fmt"
	"log"

	"github.com/xssnick/tonutils-go/ton"
)

func main() {
	api := ton.NewAPIClient(ton.NewHTTPClient("https://toncenter.com/api/v2", ""))

	info, err := api.GetMasterchainInfo(context.Background())
	if err != nil {
		log.Fatal("获取主链信息失败:", err)
	}

	fmt.Printf("最新区块 Seqno: %d\n", info.Last.Seqno)
	fmt.Printf("最新区块 Hash: %x\n", info.Last.RootHash)
}
