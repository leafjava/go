package main

import (
	"errors"
	"fmt"
)

// TODO: 定义 DeFiProtocol 接口
type DeFiProtocol interface {
	Deposit(amount float64) error
	Withdraw(amount float64) error
	GetAPY() float64
	GetTVL() float64
	GetName() string
}

// TODO: 实现 Uniswap 结构体（模拟）
type Uniswap struct {
	name    string
	balance float64
	tvl     float64
	apy     float64
}

func NewUniswap() *Uniswap {
	return &Uniswap{
		name:    "Uniswap",
		balance: 0,
		tvl:     1000000,
		apy:     5.5,
	}
}

func (u *Uniswap) Deposit(amount float64) error {
	if amount <= 0 {
		return errors.New("存款金额必须大于0")
	}
	u.balance += amount
	u.tvl += amount
	fmt.Printf("[%s] 存款成功: %.2f, 当前余额: %.2f\n", u.name, amount, u.balance)
	return nil
}

func (u *Uniswap) Withdraw(amount float64) error {
	if amount <= 0 {
		return errors.New("取款金额必须大于0")
	}
	if amount > u.balance {
		return errors.New("余额不足")
	}
	u.balance -= amount
	u.tvl -= amount
	fmt.Printf("[%s] 取款成功: %.2f, 当前余额: %.2f\n", u.name, amount, u.balance)
	return nil
}

func (u *Uniswap) GetAPY() float64 {
	return u.apy
}

func (u *Uniswap) GetTVL() float64 {
	return u.tvl
}

func (u *Uniswap) GetName() string {
	return u.name
}

// TODO: 实现 Aave 结构体（模拟）
type Aave struct {
	name    string
	balance float64
	tvl     float64
	apy     float64
}

func NewAave() *Aave {
	return &Aave{
		name:    "Aave",
		balance: 0,
		tvl:     2000000,
		apy:     7.2,
	}
}

func (a *Aave) Deposit(amount float64) error {
	if amount <= 0 {
		return errors.New("存款金额必须大于0")
	}
	a.balance += amount
	a.tvl += amount
	fmt.Printf("[%s] 存款成功: %.2f, 当前余额: %.2f\n", a.name, amount, a.balance)
	return nil
}

func (a *Aave) Withdraw(amount float64) error {
	if amount <= 0 {
		return errors.New("取款金额必须大于0")
	}
	if amount > a.balance {
		return errors.New("余额不足")
	}
	a.balance -= amount
	a.tvl -= amount
	fmt.Printf("[%s] 取款成功: %.2f, 当前余额: %.2f\n", a.name, amount, a.balance)
	return nil
}

func (a *Aave) GetAPY() float64 {
	return a.apy
}

func (a *Aave) GetTVL() float64 {
	return a.tvl
}

func (a *Aave) GetName() string {
	return a.name
}

// TODO: 实现 Compound 结构体（模拟）
type Compound struct {
	name    string
	balance float64
	tvl     float64
	apy     float64
}

func NewCompound() *Compound {
	return &Compound{
		name:    "Compound",
		balance: 0,
		tvl:     1500000,
		apy:     6.8,
	}
}

func (c *Compound) Deposit(amount float64) error {
	if amount <= 0 {
		return errors.New("存款金额必须大于0")
	}
	c.balance += amount
	c.tvl += amount
	fmt.Printf("[%s] 存款成功: %.2f, 当前余额: %.2f\n", c.name, amount, c.balance)
	return nil
}

func (c *Compound) Withdraw(amount float64) error {
	if amount <= 0 {
		return errors.New("取款金额必须大于0")
	}
	if amount > c.balance {
		return errors.New("余额不足")
	}
	c.balance -= amount
	c.tvl -= amount
	fmt.Printf("[%s] 取款成功: %.2f, 当前余额: %.2f\n", c.name, amount, c.balance)
	return nil
}

func (c *Compound) GetAPY() float64 {
	return c.apy
}

func (c *Compound) GetTVL() float64 {
	return c.tvl
}

func (c *Compound) GetName() string {
	return c.name
}

// TODO: 实现聚合器函数
func findBestAPY(protocols []DeFiProtocol) DeFiProtocol {
	// 找到 APY 最高的协议
	if len(protocols) == 0 {
		return nil
	}

	bestProtocol := protocols[0]
	for _, protocol := range protocols[1:] {
		if protocol.GetAPY() > bestProtocol.GetAPY() {
			bestProtocol = protocol
		}
	}
	return bestProtocol
}

func main() {
	// 测试 DeFi 协议
	fmt.Println("=== DeFi 协议接口测试 ===\n")

	// 创建协议实例
	uniswap := NewUniswap()
	aave := NewAave()
	compound := NewCompound()

	// 显示初始信息
	fmt.Println("--- 协议信息 ---")
	protocols := []DeFiProtocol{uniswap, aave, compound}
	for _, p := range protocols {
		fmt.Printf("%s - APY: %.2f%%, TVL: $%.2f\n", p.GetName(), p.GetAPY(), p.GetTVL())
	}
	fmt.Println()

	// 测试存款
	fmt.Println("--- 测试存款 ---")
	uniswap.Deposit(1000)
	aave.Deposit(2000)
	compound.Deposit(1500)
	fmt.Println()

	// 测试取款
	fmt.Println("--- 测试取款 ---")
	uniswap.Withdraw(500)
	aave.Withdraw(1000)
	fmt.Println()

	// 测试错误情况
	fmt.Println("--- 测试错误情况 ---")
	err := compound.Withdraw(2000)
	if err != nil {
		fmt.Printf("错误: %v\n", err)
	}
	fmt.Println()

	// 找到最佳 APY
	fmt.Println("--- 寻找最佳 APY ---")
	best := findBestAPY(protocols)
	if best != nil {
		fmt.Printf("最佳协议: %s, APY: %.2f%%\n", best.GetName(), best.GetAPY())
	}
}
