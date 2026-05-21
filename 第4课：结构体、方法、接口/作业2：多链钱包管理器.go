package main

import (
	"errors"
	"fmt"
)

// TODO: 定义 ChainWallet 接口
type ChainWallet interface {
	// GetBalance() float64
	// Deposit(amount float64)
	// Withdraw(amount float64) error
	// GetChainName() string
	GetBalance() float64
	Deposit(amount float64)
	Withdraw(amount float64) error
	GetChainName() string
}

// TODO: 实现 EthWallet 结构体和方法

type EthWallet struct {
	Address string
	Balance float64
}

func (e *EthWallet) GetBalance() float64 { return e.Balance }
func (e *EthWallet) Deposit(amount float64) { e.Balance += amount }
func (e *EthWallet) Withdraw(amount float64) error {
	if amount > e.Balance {
		return errors.New("余额不足")
	}
	e.Balance -= amount
	return nil
}
func (e *EthWallet) GetChainName() string { return "Ethereum" }

// TODO: 实现 TONWallet 结构体和方法

type TONWallet struct {
	Address string
	Balance float64
}

func (t *TONWallet) GetBalance() float64 { return t.Balance }
func (t *TONWallet) Deposit(amount float64) { t.Balance += amount }
func (t *TONWallet) Withdraw(amount float64) error {
	if amount > t.Balance {
		return errors.New("余额不足")
	}
	t.Balance -= amount
	return nil
}
func (t *TONWallet) GetChainName() string { return "TON" }

// TODO: 实现 MultiChainWallet 结构体
type MultiChainWallet struct {
	// wallets map[string]ChainWallet
	wallets map[string]ChainWallet
}

// TODO: 实现方法
// 1. AddWallet(chain string, wallet ChainWallet)
// 2. GetTotalBalance() float64
// 3. GetWallet(chain string) (ChainWallet, error)

func NewMultiChainWallet() *MultiChainWallet {
	return &MultiChainWallet{wallets: make(map[string]ChainWallet)}
}

func (m *MultiChainWallet) AddWallet(chain string, wallet ChainWallet) {
	m.wallets[chain] = wallet
}

func (m *MultiChainWallet) GetTotalBalance() float64 {
	var total float64
	for _, w := range m.wallets {
		total += w.GetBalance()
	}
	return total
}

func (m *MultiChainWallet) GetWallet(chain string) (ChainWallet, error) {
	w, ok := m.wallets[chain]
	if !ok {
		return nil, errors.New("wallet not found")
	}
	return w, nil
}

func main() {
	// 测试多链钱包
	mc := NewMultiChainWallet()

	eth := &EthWallet{Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb", Balance: 1.5}
	ton := &TONWallet{Address: "EQD...", Balance: 100.0}

	mc.AddWallet(eth.GetChainName(), eth)
	mc.AddWallet(ton.GetChainName(), ton)

	fmt.Printf("总余额: %.2f\n", mc.GetTotalBalance())

	// 操作示例
	if w, err := mc.GetWallet("Ethereum"); err == nil {
		w.Deposit(0.5)
		fmt.Printf("Ethereum 余额: %.2f\n", w.GetBalance())
	}

	if w, err := mc.GetWallet("TON"); err == nil {
		if err := w.Withdraw(10); err != nil {
			fmt.Println("TON 提现失败:", err)
		} else {
			fmt.Printf("TON 余额: %.2f\n", w.GetBalance())
		}
	}

	fmt.Printf("操作后总余额: %.2f\n", mc.GetTotalBalance())
}
