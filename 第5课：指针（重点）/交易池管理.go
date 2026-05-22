package main

import "fmt"

type Transaction struct {
	Hash   string
	From   string
	To     string
	Amount float64
	Status string
}

type TransactionPool struct {
	Transactions []*Transaction // 指针切片
	MaxSize      int
}

// 添加交易
func (tp *TransactionPool) AddTransaction(tx *Transaction) error {
	if len(tp.Transactions) >= tp.MaxSize {
		return fmt.Errorf("交易池已满: %d/%d", len(tp.Transactions), tp.MaxSize)
	}
	tp.Transactions = append(tp.Transactions, tx)
	return nil
}

// 确认交易
func (tp *TransactionPool) ConfirmTransaction(hash string) bool {
	for _, tx := range tp.Transactions {
		if tx.Hash == hash {
			tx.Status = "confirmed" // 直接修改原交易
			return true
		}
	}
	return false
}

// 获取待处理交易
func (tp *TransactionPool) GetPendingTransactions() []*Transaction {
	var pending []*Transaction // 建一个空切片，准备装结果
	//for 索引, 元素 := range 切片 { }
	for _, tx := range tp.Transactions { // 遍历所有交易
		if tx.Status == "pending" { // 状态是 pending 的
			pending = append(pending, tx) // 就加进去
		}
	}
	return pending // 返回过滤后的结果
}

func main() {
	pool := &TransactionPool{
		Transactions: make([]*Transaction, 0),
		MaxSize:      100,
	}

	// 添加交易
	tx1 := &Transaction{
		Hash:   "0xabc123",
		From:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
		To:     "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
		Amount: 1.5,
		Status: "pending",
	}

	pool.AddTransaction(tx1)

	// 确认交易
	pool.ConfirmTransaction("0xabc123")

	fmt.Println("交易状态:", tx1.Status) // confirmed
}
