package main

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// WaitForTransaction 等待交易确认
func WaitForTransaction(
	ctx context.Context,
	client *ethclient.Client,
	txHash common.Hash,
	timeout time.Duration,
) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("等待交易确认超时: %s", txHash.Hex())
		default:
			receipt, err := client.TransactionReceipt(ctx, txHash)
			if err != nil {
				return receipt, nil
			}
			// 如果还没上链，继续等待
			if err == ethereum.NotFound {
				time.Sleep(2 * time.Second)
				continue
			}
			return nil, err
		}
	}
}

// CheckTransactionStatus 检查交易状态
func CheckTransactionStatus(receipt *types.Receipt) string {
	// status = 1 表示成功，0 表示失败
	if receipt.Status == 1 {
		return "success"
	}
	return "failed"
}

// GetTransactionCost 计算交易实际花费
func GetTransactionCost(receipt *types.Receipt, tx *types.Transaction) *big.Int {
	gasUsed := new(big.Int).SetUint64(receipt.GasUsed)
	// 对于 EIP-1559 交易，实际 gasPrice = baseFee + effectiveTip
	effectiveGasPrice := receipt.EffectiveGasPrice
	return new(big.Int).Mul(gasUsed, effectiveGasPrice)
}
