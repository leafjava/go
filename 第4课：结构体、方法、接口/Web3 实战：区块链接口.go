package main

import "fmt"

type Blockchain interface {
	GetBalance(address string) (float64, error)
	SendTransaction(from, to string, amount float64) (string, error)
	GetBlockNumber() (int64, error)
}

type Ethereum struct {
	NodeURL string
}

func (e *Ethereum) GetBalance(address string) (float64, error) {
	fmt.Println("查询以太坊余额:", address)
	return 1.5, nil
}

func (e *Ethereum) SendTransaction(from, to string, amount float64) (string, error) {
	fmt.Printf("发送 %.2f ETH 从 %s 到 %s\n", amount, from, to)
	return "0xabc123...", nil
}

func (e *Ethereum) GetBlockNumber() (int64, error) {
	return 18500, nil
}

type TON struct {
	NodeURL string
}

func (t *TON) GetBalance(address string) (float64, error) {
	fmt.Println("查询 TON 余额:", address)
	return 100.0, nil
}

func (t *TON) SendTransaction(from, to string, amount float64) (string, error) {
	fmt.Printf("发送 %.2f TON 从 %s 到 %s\n", amount, from, to)
	return "abc123def456", nil
}

func (t *TON) GetBlockNumber() (int64, error) {
	return 350000, nil
}

func queryBlockchain(bc Blockchain, address string) {
	balance, err := bc.GetBalance(address)
	if err != nil {
		fmt.Println("查询失败:", err)
		return
	}

	blockNum, _ := bc.GetBlockNumber()
	fmt.Printf("余额:%.2f,区块高度:%d\n", balance, blockNum)
}

func main() {
	eth := &Ethereum{NodeURL: "https://eth.llamarpc.com"}
	ton := &TON{NodeURL: "https://toncenter.com/api/v2"}

	queryBlockchain(eth, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
	fmt.Println("---")
	queryBlockchain(ton, "EQD...")
}
