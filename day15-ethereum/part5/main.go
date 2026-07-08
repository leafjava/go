package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ERC20 ABI（简化版）
const erc20ABI = `[
    {
        "constant": true,
        "inputs": [{"name": "_owner", "type": "address"}],
        "name": "balanceOf",
        "outputs": [{"name": "balance", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [],
        "name": "decimals",
        "outputs": [{"name": "", "type": "uint8"}],
        "type": "function"
    }
]`

func main() {
	client, err := ethclient.Dial("https://ethereum.publicnode.com")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// USDT 合约地址
	tokenAddress := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")

	// 用户地址
	userAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")

	// 解析 ABI
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		log.Fatal(err)
	}

	// 编码 balanceOf 调用
	data, err := parsedABI.Pack("balanceOf", userAddress)
	if err != nil {
		log.Fatal(err)
	}

	// 调用合约
	msg := ethereum.CallMsg{
		To:   &tokenAddress,
		Data: data,
	}

	result, err := client.CallContract(context.Background(), msg, nil)
	if err != nil {
		log.Fatal(err)
	}

	// 解码结果
	balance := new(big.Int)
	balance.SetBytes(result)

	fmt.Printf("USDT 余额: %s\n", balance.String())
}
