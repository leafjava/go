package main

import (
	"errors"
	"fmt"
)

type Wallet struct {
	Address string
	Balance float64
}

// 存款（指针接收者）
func (w *Wallet) Deposit(amount float64) error {
	if amount <= 0 {
		return errors.New("存款金额必须大于0")
	}
	w.Balance += amount
	return nil
}

// 取款（指针接收者）
func (w *Wallet) Withdraw(amount float64) error {
	if amount <= 0 {
		return errors.New("取款金额必须大于0")
	}
	if amount > w.Balance {
		return fmt.Errorf("余额不足:需要%.2f,当前%.2f", amount, w.Balance)
	}
	w.Balance -= amount
	return nil
}

// 转账
func Transfer(from, to *Wallet, amount float64) error {
	if err := from.Withdraw(amount); err != nil {
		return err
	}
	if err := to.Deposit(amount); err != nil {
		from.Deposit(amount)
		return err
	}
	fmt.Printf("转账成功: %.2f 从 %s 到 %s\n", amount, from.Address, to.Address)
	return nil
}

func main() {
	wallet1 := &Wallet{
		Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
		Balance: 10.0,
	}

	wallet2 := &Wallet{
		Address: "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
		Balance: 5.0,
	}

	fmt.Printf("转账前 - 钱包1: %.2f, 钱包2: %.2f\n", wallet1.Balance, wallet2.Balance)

	if err := Transfer(wallet1, wallet2, 3.0); err != nil {
		fmt.Println("转账失败:", err)
	}

	fmt.Printf("转账后 - 钱包1: %.2f, 钱包2: %.2f\n", wallet1.Balance, wallet2.Balance)
}
