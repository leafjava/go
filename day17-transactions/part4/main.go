package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ### 并发发送多笔交易
type TransferRequest struct {
	To     string
	Amount *big.Int
}

type TransferResult struct {
	Index int
	Hash  string
	Error error
}

func BatchTransfer(
	client *ethclient.Client,
	privateKeyHex string,
	requests []TransferRequest,
) ([]TransferResult, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return nil, fmt.Errorf("私钥解析失败: %w", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	// 获取初始 nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return nil, fmt.Errorf("获取 nonce 失败: %w", err)
	}

	chainID, _ := client.NetworkID(context.Background())
	gasTipCap, _ := client.SuggestGasTipCap(context.Background())
	gasFeeCap, _ := client.SuggestGasPrice(context.Background())

	results := make([]TransferResult, len(requests))
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 并发构建和签名交易（注意 nonce 的顺序）
	signedTxs := make([]*types.Transaction, len(requests))

	for i, req := range requests {
		toAddr := common.HexToAddress(req.To)
		tx := types.NewTx(&types.DynamicFeeTx{
			ChainID:   chainID,
			Nonce:     nonce + uint64(i), // nonce 递增
			GasTipCap: gasTipCap,
			GasFeeCap: gasFeeCap,
			Gas:       21000,
			To:        &toAddr,
			Value:     req.Amount,
			Data:      nil,
		})

		signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)

		if err != nil {
			results[i] = TransferResult{Index: i, Error: err}
			continue
		}
		signedTxs[i] = signedTx
	}

	// 并发送交易
	for i, signedTx := range signedTxs {
		if signedTx == nil {
			continue
		}

		wg.Add(1)
		go func(idx int, tx *types.Transaction) {
			defer wg.Done()

			err := client.SendTransaction(context.Background(), tx)
			mu.Lock()
			if err != nil {
				results[idx] = TransferResult{Index: idx, Error: err}
			} else {
				results[idx] = TransferResult{Index: idx, Hash: tx.Hash().Hex()}
			}
			mu.Unlock()
		}(i, signedTx)
	}

	wg.Wait()
	return results, nil
}

func main() {
	fmt.Println("BatchTransfer — 并发批量转账示例")
	fmt.Println("核心功能：构建多笔 nonce 递增的交易，并发广播到链上")
	fmt.Println("使用前需填入实际的 RPC URL、私钥和转账列表")
}
