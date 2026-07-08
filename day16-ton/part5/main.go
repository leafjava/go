package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

func main() {
	client := liteclient.NewConnectionPool()
	err := client.AddConnectionsFromConfigUrl(context.Background(), "https://ton.org/global.config.json")

	if err != nil {
		log.Fatal(err)
	}

	api := ton.NewAPIClient(client)

	// 获取主链最新区块
	block, err := api.GetMasterchainInfo(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// 从助记词恢复钱包
	words := strings.Split("your mnemonic words here", " ")
	w, err := wallet.FromSeed(api, words, wallet.V4R2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("发送方地址: %s\n", w.Address().String())

	// 接收方地址
	toAddr := address.MustParseAddr("EQA0gvftODUs9itSl5whd01IxERToRZkA_-tsriaZPxuRkS4")

	// 构造转账 body（备注）
	body, err := wallet.CreateCommentCell("转账测试")
	if err != nil {
		log.Fatal(err)
	}

	// 转账 0.05 TON
	err = w.Send(context.Background(), &wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: false,
			Bounce:      false,
			Bounced:     false,
			SrcAddr:     w.WalletAddress(),
			DstAddr:     toAddr,
			Amount:      tlb.MustFromTON("0.05"),
			StateInit:   nil,
			Body:        body,
		},
	}, true)

	if err != nil {
		log.Fatal("转账失败:", err)
	}

	fmt.Println("转账成功!")

	// 获取当前余额
	balance, err := w.GetBalance(context.Background(), block)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("当前余额: %.9f TON\n", float64(balance.Nano().Uint64())/1e9)
}
