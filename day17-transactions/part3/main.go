package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ERC-20 transfer 函数的 ABI 定义
const erc20ABI = `[
    {
        "constant": false,
        "inputs": [
            {"name": "_to", "type": "address"},
            {"name": "_value", "type": "uint256"}
        ],
        "name": "transfer",
        "outputs": [{"name": "success", "type": "bool"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [{"name": "_owner", "type": "address"}],
        "name": "balanceOf",
        "outputs": [{"name": "balance", "type": "uint256"}],
        "type": "function"
    }
]`

func SendERC20Token(
	client *ethclient.Client,
	privateKeyHex string,
	tokenContractAddress string,
	toAddress string,
	amount *big.Int,
) (string, error) {
	// 1. 解析私钥
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("私钥解析失败: %w", err)
	}

	// 2. 获取发送方地址
	publicKey := privateKey.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 3. 解析 ABI
	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return "", fmt.Errorf("ABI 解析失败: %w", err)
	}

	// 4. 编码 transfer 函数调用
	toAddr := common.HexToAddress(toAddress)
	data, err := parsedABI.Pack("transfer", toAddr, amount)
	if err != nil {
		return "", fmt.Errorf("ABI 编码失败: %w", err)
	}

	// 5. 获取 nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", fmt.Errorf("获取 nonce 失败: %w", err)
	}

	// 6. 估算 Gas
	tokenAddr := common.HexToAddress(tokenContractAddress)
	msg := ethereum.CallMsg{
		From: fromAddress,
		To:   &tokenAddr,
		Data: data,
	}
	gasLimit, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		return "", fmt.Errorf("估算 Gas 失败: %w", err)
	}
	gasLimit = gasLimit * 12 / 10

	// 7. 获取 Gas 费用
	gasTipCap, _ := client.SuggestGasTipCap(context.Background())
	gasFeeCap, _ := client.SuggestGasPrice(context.Background())

	// 8. 获取链 ID
	chainID, _ := client.NetworkID(context.Background())

	// 9. 构建交易（to 是代币合约地址，value 为 0，data 是 transfer 调用）
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &tokenAddr,
		Value:     big.NewInt(0), // ETH 转账金额为 0
		Data:      data,
	})

	// 10. 签名
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}

	// 11. 发送
	err = client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", fmt.Errorf("发送失败: %w", err)
	}

	return signedTx.Hash().Hex(), nil
}

func main() {
	// 示例：发送 ERC-20 代币转账
	// 使用前请填入实际的 RPC URL、私钥、代币合约地址和目标地址
	fmt.Println("SendERC20Token — ERC-20 代币转账示例")
	fmt.Println("")
	fmt.Println("使用方式：")
	fmt.Println("  1. 连接以太坊节点（主网/测试网/Sepolia/Holesky）")
	fmt.Println("  2. 准备一个有 ETH（付 Gas）和代币的钱包私钥")
	fmt.Println("  3. 将下方占位符替换为实际值后运行")
	fmt.Println("")
	fmt.Println("  RPC_URL              = \"https://sepolia.infura.io/v3/YOUR-KEY\"")
	fmt.Println("  PRIVATE_KEY          = \"你的私钥（十六进制，去掉 0x 前缀）\"")
	fmt.Println("  TOKEN_CONTRACT       = \"0x...（ERC-20 代币合约地址）\"")
	fmt.Println("  TO_ADDRESS           = \"0x...（接收方地址）\"")
	fmt.Println("  AMOUNT               = 1000000000000000000（1 个代币，18 位精度）")
}
