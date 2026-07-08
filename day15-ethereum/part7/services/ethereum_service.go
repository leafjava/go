package services

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type EthereumService struct {
	client *ethclient.Client
}

func NewEthereumService(rpcURL string) (*EthereumService, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	return &EthereumService{client: client}, nil
}

func (s *EthereumService) GetBalance(address string) (*big.Int, error) {
	addr := common.HexToAddress(address)
	balance, err := s.client.BalanceAt(context.Background(), addr, nil)
	if err != nil {
		return nil, err
	}
	return balance, nil
}

func (s *EthereumService) SendTransaction(privateKeyHex, toAddress string, amount *big.Int) (string, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", err
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return "", fmt.Errorf("无法转换公钥")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	nonce, err := s.client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return "", err
	}

	gasLimit := uint64(21000)
	gasPrice, err := s.client.SuggestGasPrice(context.Background())
	if err != nil {
		return "", err
	}

	toAddr := common.HexToAddress(toAddress)
	tx := types.NewTransaction(nonce, toAddr, amount, gasLimit, gasPrice, nil)

	chainID, err := s.client.NetworkID(context.Background())
	if err != nil {
		return "", err
	}

	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
	if err != nil {
		return "", err
	}

	err = s.client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return "", err
	}
	return signedTx.Hash().Hex(), nil
}

func (s *EthereumService) Close() {
	s.client.Close()
}
