package main

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/xssnick/tonutils-go/address"
	"github.com/xssnick/tonutils-go/liteclient"
	"github.com/xssnick/tonutils-go/tlb"
	"github.com/xssnick/tonutils-go/ton"
	"github.com/xssnick/tonutils-go/ton/jetton"
	"github.com/xssnick/tonutils-go/ton/wallet"
)

func main() {
	client := liteclient.NewConnectionPool()
	err := client.AddConnectionsFromConfigUrl(context.Background(), "https://ton.org/global.config.json")
	if err != nil {
		log.Fatal(err)
	}

	api := ton.NewAPIClient(client)

	// 用户钱包地址
	userAddr, err := address.ParseAddr("EQC9bWZd29foipyPOGWlVNVCQzpGAjvi1rGWF7EbNcSVClpA")
	if err != nil {
		log.Fatal("解析用户地址失败:", err)
	}

	// Jetton Master 合约地址（测试代币）
	jettonMaster, err := address.ParseAddr("EQAbMQzuuGiCne0R7QEj9nrXsjM7gNjeVmrlBZouyC-SCLlO")
	if err != nil {
		log.Fatal("解析 Jetton Master 地址失败:", err)
	}

	// 创建 Jetton 客户端
	jettonClient := jetton.NewJettonMasterClient(api, jettonMaster)

	// 获取用户的 Jetton 钱包
	jettonWallet, err := jettonClient.GetJettonWallet(context.Background(), userAddr)
	if err != nil {
		log.Fatal("获取 Jetton 钱包失败:", err)
	}

	// 查询 Jetton 余额
	balance, err := jettonWallet.GetBalance(context.Background())
	if err != nil {
		log.Fatal("查询余额失败:", err)
	}

	fmt.Printf("用户钱包地址: %s\n", userAddr.String())
	fmt.Printf("Jetton 钱包地址: %s\n", jettonWallet.Address().String())
	fmt.Printf("Jetton 余额: %s\n", balance.String())

}

func TransferJetton(
	ctx context.Context,
	api ton.APIClientWrapped,
	w *wallet.Wallet,
	jettonMaster *address.Address,
	toAddr *address.Address,
	amount *big.Int,
) error {
	// 构建 Jetton 转账 payload
	amountCoins := tlb.MustFromNano(amount, 0)
	transferBody, err := jetton.BuildTransferPayload(
		toAddr,
		nil,               // responseTo：无需回复
		amountCoins,        // Jetton 数量
		tlb.MustFromTON("0.05"), // forward TON amount
		nil,               // payloadForward
		nil,               // customPayload
	)
	if err != nil {
		return fmt.Errorf("构建转账消息体失败: %w", err)
	}

	// 获取发送方的 Jetton 钱包地址
	jc := jetton.NewJettonMasterClient(api, jettonMaster)
	jettonWallet, err := jc.GetJettonWallet(ctx, w.WalletAddress())
	if err != nil {
		return fmt.Errorf("获取 Jetton 钱包地址失败: %w", err)
	}

	// 发送转账消息到 Jetton 钱包
	err = w.Send(ctx, &wallet.Message{
		Mode: wallet.PayGasSeparately + wallet.IgnoreErrors,
		InternalMessage: &tlb.InternalMessage{
			IHRDisabled: false,
			Bounce:      true,
			SrcAddr:     w.WalletAddress(),
			DstAddr:     jettonWallet.Address(),
			Amount:      tlb.MustFromTON("0.05"), // Gas 费（TON）
			Body:        transferBody,
		},
	}, true)

	return err
}
