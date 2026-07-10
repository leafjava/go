package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

func BuildAndSignTransaction(
	client *ethclient.Client,
	privateKeyHex string,
	toAddress string,
	amountWei *big.Int,
	data []byte,
) (*types.Transaction, error) {
	// 1. 解析私钥
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("私钥解析失败: %w", err)
	}

	// 2. 获取发送方地址
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("公钥类型转换失败")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 3. 获取 nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, fmt.Errorf("获取 nonce 失败: %w", err)
	}

	// 4. 获取链 ID
	chainID, err := client.NetworkID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("获取链 ID 失败: %w", err)
	}

	// 5. 估算 Gas（如果是合约调用）
	gasLimit := uint64(21000) // 普通 ETH 转账固定 21000 Gas
	if len(data) > 0 {
		// HexToAddress 返回值类型，需先赋给变量再取地址
		toCommonAddr := common.HexToAddress(toAddress)
		msg := ethereum.CallMsg{
			From:  fromAddress,
			To:    &toCommonAddr,
			Data:  data,
			Value: amountWei,
		}
		estimated, err := client.EstimateGas(context.Background(), msg)
		if err != nil {
			return nil, fmt.Errorf("估算 Gas 失败: %w", err)
		}
		gasLimit = estimated + estimated/5 // 上浮 20% 确保交易成功
	}

	// 6. 获取 Gas 费用（EIP-1559 动态费用）
	gasTipCap, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return nil, fmt.Errorf("获取 GasTipCap 失败: %w", err)
	}

	// SuggestGasPrice 返回的 gas price 可作为 GasFeeCap（最大费用上限）
	gasFeeCap, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return nil, fmt.Errorf("获取 GasFeeCap 失败: %w", err)
	}

	// 7. 构建交易
	toAddr := common.HexToAddress(toAddress)
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &toAddr,
		Value:     amountWei,
		Data:      data,
	})

	// 8. 签名
	signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
	if err != nil {
		return nil, fmt.Errorf("签名失败: %w", err)
	}
	return signedTx, nil
}

func main() {
	// 示例：构建并签名一笔 EIP-1559 动态费用交易
	fmt.Println("BuildAndSignTransaction — EIP-1559 交易构建与签名示例")
	fmt.Println("请在代码中填入实际的 RPC URL、私钥和目标地址来测试")
}

func OfflineSignTransaction(
	privateKeyHex string,
	toAddress string,
	amountWei *big.Int,
	nonce uint64,
	chainID *big.Int,
	gasLimit uint64,
	gasTipCap *big.Int,
	gasFeeCap *big.Int,
	data []byte,
) (*types.Transaction, error) {
	// 解析私钥（无需连接节点）
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("私钥解析失败: %w", err)
	}

	toAddr := common.HexToAddress(toAddress)

	// 构建交易
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainID,
		Nonce:     nonce,
		GasTipCap: gasTipCap,
		GasFeeCap: gasFeeCap,
		Gas:       gasLimit,
		To:        &toAddr,
		Value:     amountWei,
		Data:      data,
	})

	// 离线签名
	signer := types.LatestSignerForChainID(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return nil, fmt.Errorf("签名失败: %w", err)
	}

	return signedTx, nil

}
