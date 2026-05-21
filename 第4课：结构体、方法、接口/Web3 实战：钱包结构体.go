package main

import "fmt"

type Wallet struct {
	Address    string
	PrivateKey string
	Balance    float64
	Network    string
}

type Transaction struct {
	Hash      string
	From      string
	To        string
	Amount    float64
	GasUsed   int64
	Status    string
	Timestamp int64
}

type NFT struct {
	TokenID     string
	Name        string
	Description string
	ImageURL    string
	Owner       string
	Contract    string
}

func main() {
	wallet := Wallet{
		Address:    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
		PrivateKey: "0x...",
		Balance:    1000,
		Network:    "Ethereum",
	}

	tx := Transaction{
		Hash:      "0xabc123...",
		From:      wallet.Address,
		To:        "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
		Amount:    1000,
		GasUsed:   21000,
		Status:    "pending",
		Timestamp: 1704067200,
	}

	fmt.Println("钱包:%s,余额:%.2f ETH", wallet.Address, wallet.Balance)
	fmt.Printf("交易: %s, 金额: %.2f ETH\n", tx.Hash, tx.Amount)
}
