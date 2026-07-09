package main

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// ### Legacy 交易（EIP-155 之前）
func main() {
	// Legacy 交易结构
	tx := types.NewTransaction(
		0,                               // nonce（发送方交易计数）
		common.HexToAddress("..."),      // to（接收方地址）
		big.NewInt(1000000000000000000), // value（转账金额，1 ETH = 10^18 Wei）
		21000,                           // gasLimit
		big.NewInt(20000000000),         // gasPrice（20 Gwei）
		nil,                             // data（合约调用数据）
	)

	fmt.Printf("交易哈希: %s\n", tx.Hash().Hex())
	fmt.Printf("交易类型: Legacy\n")
	fmt.Printf("Gas 费用上限: %d\n", tx.Gas())
	fmt.Printf("Gas 价格: %s Wei\n", tx.GasPrice().String())
}
