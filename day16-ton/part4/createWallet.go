package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

//### 创建钱包

func main() {
	client := liteclient.NewConnectionPool()
	err := client.AddConnectionsFromConfigUrl(context.Background(), "https://ton.org/global.config.json")
	if err != nil {
		log.Fatal(err)
	}

	api := ton.NewAPIClient(client)

	// 生成助记词（新钱包）
	words := wallet.NewSeed()
	fmt.Println("助记词（请安全保存）:")
	fmt.Println(strings.Join(words, " "))

	// 从助记词恢复钱包
	w, err := wallet.FromSeed(api, words, wallet.V4R2)
	if err != nil {
		log.Fatal("创建钱包失败:", err)
	}

	// 获取钱包地址
	address := w.Address()
	fmt.Printf("钱包地址（Bounceable）: %s\n", address.String())
	fmt.Printf("钱包地址（Non-Bounceable）: %s\n", address.Bounce(false).String())
}
