package main

import (
	"errors"
	"fmt"
)

// TODO: 创建钱包管理器（使用闭包）
func createWallet(initialBalance float64) (
	deposit func(float64) float64,
	withdraw func(float64) (float64, error),
	getBalance func() float64,
) {
	// 实现存款、取款、查询余额三个闭包函数
	// balance 变量被三个函数共享
	balance := initialBalance

	deposit = func(amount float64) float64 {
		balance += amount
		return balance
	}

	withdraw = func(amount float64) (float64, error) {
		if amount > balance {
			return balance, errors.New("余额不足")
		}
		balance -= amount
		return balance, nil
	}

	getBalance = func() float64 {
		return balance
	}

	return deposit, withdraw, getBalance
}

func main() {
	// 创建钱包
	deposit, withdraw, getBalance := createWallet(100.0)

	// 测试存款
	fmt.Println("存款后余额:", deposit(50))

	// 测试取款
	newBalance, err := withdraw(30)
	if err != nil {
		fmt.Println("取款失败:", err)
	} else {
		fmt.Println("取款后余额:", newBalance)
	}

	// 测试查询
	fmt.Println("当前余额:", getBalance())
}
