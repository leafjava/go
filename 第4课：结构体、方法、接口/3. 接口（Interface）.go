package main

import "fmt"

type PaymentMethod interface {
	Pay(amount float64) error
	GetName() string
}

type EthereumPayment struct {
	WalletAddress string
}

func (e *EthereumPayment) Pay(amount float64) error {
	fmt.Printf("使用以太坊支付 %.2f ETH\n", amount)
	return nil
}

func (e *EthereumPayment) GetName() string {
	return "Ethereum"
}

type TONPayment struct {
	walletAddress string
}

func (t *TONPayment) Pay(amount float64) error {
	fmt.Printf("使用 TON 支付 %.2f TON\n", amount)
	return nil
}

func (t *TONPayment) GetName() string {
	return "TON"
}

func processPayment(pm PaymentMethod, amount float64) {
	fmt.Printf("支付方式:%s\n", pm.GetName())
	if err := pm.Pay(amount); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("支付成功!")
	}
}

func main() {
	ethPayment := &EthereumPayment{
		WalletAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
	}

	tonPayment := &TONPayment{
		walletAddress: "EQD...",
	}

	processPayment(ethPayment, 1.5)
	fmt.Println("---")
	processPayment(tonPayment, 1.5)
}
