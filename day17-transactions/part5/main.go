package main

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// WaitForTransaction 等待交易确认
// 轮询 TransactionReceipt，直到交易被打包上链或超时
func WaitForTransaction(
	ctx context.Context,
	client *ethclient.Client,
	txHash common.Hash,
	timeout time.Duration,
) (*types.Receipt, error) {
	// 设置超时上下文，防止无限等待
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			// 超时：返回 nil receipt + 超时错误
			return nil, fmt.Errorf("等待交易确认超时: %s", txHash.Hex())
		default:
			receipt, err := client.TransactionReceipt(ctx, txHash)
			if err != nil {
				// 交易尚未被打包（receipt 还不存在），继续等待
				if errors.Is(err, ethereum.NotFound) {
					time.Sleep(2 * time.Second)
					continue
				}
				// 其他错误（网络问题等），直接返回
				return nil, fmt.Errorf("查询收据失败: %w", err)
			}
			// 拿到收据，返回
			return receipt, nil
		}
	}
}

// CheckTransactionStatus 检查交易状态
// Status=1 成功，Status=0 失败（链上回滚）
func CheckTransactionStatus(receipt *types.Receipt) string {
	if receipt.Status == 1 {
		return "success"
	}
	return "failed"
}

// GetTransactionCost 计算交易实际 Gas 花费（单位：Wei）
// GasUsed × EffectiveGasPrice = 总 Gas 费
func GetTransactionCost(receipt *types.Receipt, tx *types.Transaction) *big.Int {
	gasUsed := new(big.Int).SetUint64(receipt.GasUsed)
	// EffectiveGasPrice 是 EIP-1559 交易的实际成交 Gas 价（baseFee + effectiveTip）
	effectiveGasPrice := receipt.EffectiveGasPrice
	return new(big.Int).Mul(gasUsed, effectiveGasPrice)
}

func main() {
	fmt.Println("交易状态监控工具")
	fmt.Println("提供三个核心功能：")
	fmt.Println("  WaitForTransaction   — 轮询等待交易确认")
	fmt.Println("  CheckTransactionStatus — 判断交易成功/失败")
	fmt.Println("  GetTransactionCost   — 计算实际 Gas 花费")
}
