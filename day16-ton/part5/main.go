package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
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

	// 从助记词恢复钱包
	words := strings.Split("your mnemonic words here", " ")
	w, err := wallet.FromSeed(api, words, wallet.V4R2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("发送方地址: %s\n", w.Address().String())

	// 接收方地址
	toAddr := address.MustParseAddr("EQCkR1cGZxYbG6QrGqKJ8HxSPfI3NZ9sNjCpHlLHyVTGc5Gq")

	// 转账 0.05 TON
	// 注意：转账时附带 comment（备注）
	err = w.Send(context.Background(), &wallet.Message{
		Mode: wallet.MsgWithRemainingValue,
		InternalMessage: &wallet.InternalMessage{
			IHRDisabled: false,
			Bounce:      false,
			Bounced:     false,
			SrcAddr:     w.WalletAddress(),
			DstAddr:     toAddr,
			Amount:      ton.MustToNano("0.05"),
			StateInit:   nil,
			Bode:        wallet.Comment("转账测试"),
		},
	}, true)

	if err != nil {
		log.Fatal("转账失败:", err)
	}

	fmt.Println("转账成功!")

	// 获取当前余额
	balance, err := w.GetBalance(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("当前余额: %.9f TON\n", float64(balance.Nano().Uint64())/1e9)
}
