package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("https://ethereum-rpc.publicnode.com")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 地址
	address := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")

	// 查询余额（Wei）
	balance, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("余额（Wei）: %s\n", balance.String())

	// 转换为 ETH
	fbalance := new(big.Float)
	fbalance.SetString(balance.String())
	ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))

	fmt.Printf("余额（ETH）: %f\n", ethValue)
}
