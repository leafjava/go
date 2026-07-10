package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

func main() {
	client, err := ethclient.Dial("https://eth.llamarpc.com")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 获取当前的 Gas 费用建议
	tipCap, _ := client.SuggestGasTipCap(context.Background())

	feeCap, _ := client.SuggestGasPrice(context.Background())

	chainID, _ := client.NetworkID(context.Background())

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     0,
		To:        &common.Address{},
		Value:     big.NewInt(1000000000000000000),
		Gas:       21000,
		GasFeeCap: feeCap, // maxFeePerGas（最高愿意支付的单价）
		GasTipCap: tipCap, // maxPriorityFeePerGas（给矿工的小费）
		Data:      nil,
	})

	fmt.Printf("交易类型: DynamicFee (EIP-1559)\n")
	fmt.Printf("MaxFeePerGas: %s Wei\n", feeCap.String())
	fmt.Printf("MaxPriorityFeePerGas: %s Wei\n", tipCap.String())
	fmt.Printf("链 ID: %d\n", chainID)
}
