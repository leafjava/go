package main

import "fmt"

type Transaction struct {
	Hash   string
	Amount float64
	Status string
}

func main() {
	// 创建交易列表
	txs := make([]Transaction, 0, 10)

	// 添加交易
	txs = append(txs, Transaction{
		Hash:   "0xabc123",
		Amount: 1.5,
		Status: "Pending",
	})

	txs = append(txs, Transaction{
		Hash:   "0xdef456",
		Amount: 2.0,
		Status: "confirmed",
	})

	// 遍历
	for i, tx := range txs {
		fmt.Printf("%d: %s - %.2f ETH (%s)\n", i, tx.Hash, tx.Amount, tx.Status)
	}

	// 过滤已确认的交易
	confirmed := make([]Transaction, 0)
	for _, tx := range txs {
		if tx.Status == "confirmed" {
			confirmed = append(confirmed, tx)
		}
	}

	fmt.Println("已确认交易数:", len(confirmed))

}
