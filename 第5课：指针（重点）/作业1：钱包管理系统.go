package main

import (
	"errors"
	"fmt"
)

type Wallet struct {
	Address string  // 钱包地址（区块链地址）
	Balance float64 // 余额
}

// Deposit 存款：向钱包存入指定金额
// 使用指针接收者 (w *Wallet)，因为需要修改 Balance
func (w *Wallet) Deposit(amount float64) error {
	if amount <= 0 { // 校验：金额必须大于0
		return errors.New("存款金额必须大于0")
	}
	w.Balance += amount   // 通过指针直接修改原始钱包的余额
	return nil            // 成功返回 nil（Go 的 null）
}

// Withdraw 取款：从钱包取出指定金额
func (w *Wallet) Withdraw(amount float64) error {
	if amount <= 0 { // 校验：金额必须大于0
		return errors.New("取款金额必须大于0")
	}
	if w.Balance < amount { // 校验：余额不足
		return errors.New("余额不足")
	}
	w.Balance -= amount // 通过指针修改余额
	return nil
}

// Transfer 转账：从当前钱包转钱给目标钱包
// to *Wallet —— 目标钱包的指针，修改它才能生效
func (w *Wallet) Transfer(to *Wallet, amount float64) error {
	if to == nil { // 防御：目标钱包不能为空指针
		return errors.New("目标钱包不存在")
	}
	if err := w.Withdraw(amount); err != nil { // 先扣钱
		return err // 扣钱失败直接返回错误
	}
	to.Balance += amount // 对方加钱（指针修改，直接生效）
	return nil
}

// BatchTransfer 批量转账：从 from 向每个 recipient 各转 amount
func BatchTransfer(from *Wallet, recipients []*Wallet, amount float64) error {
	// 从 from 向每个 recipient 转账 amount
	// 如果任何一笔失败，回滚所有已完成的转账

	var transferred []*Wallet               // 记录已成功转账的收款人，用于回滚
	for i, to := range recipients {         // 遍历所有收款人
		if err := from.Transfer(to, amount); err != nil { // 执行单笔转账
			// 转账失败 → 回滚所有已完成的转账
			for _, r := range transferred {   // 遍历已转账的收款人
				r.Balance -= amount           // 把钱从收款人那里退回来
				from.Balance += amount        // 还给付款人
			}
			return fmt.Errorf("第 %d 笔转账失败: %w", i+1, err) // 返回带序号的错误
		}
		transferred = append(transferred, to) // 记录本次转账成功，加入回滚列表
	}
	return nil
}

func main() {
	// —————— 1. 创建钱包 ——————
	alice := &Wallet{Address: "0xAlice001", Balance: 100.0} // & 取地址，拿到指针
	bob := &Wallet{Address: "0xBob002", Balance: 50.0}
	charlie := &Wallet{Address: "0xCharlie003", Balance: 30.0}

	fmt.Println("===== 初始余额 =====")
	fmt.Printf("Alice:  %.2f ETH\n", alice.Balance)  // %.2f 保留两位小数
	fmt.Printf("Bob:    %.2f ETH\n", bob.Balance)
	fmt.Printf("Charlie: %.2f ETH\n", charlie.Balance)

	// —————— 2. 测试存款 ——————
	fmt.Println("\n===== 测试存款 =====")
	if err := alice.Deposit(50); err != nil { // err := 是 Go 的异常处理惯用法
		fmt.Println("存款失败:", err)
	} else {
		fmt.Printf("Alice 存入 50 → 余额: %.2f\n", alice.Balance)
	}

	// —————— 3. 测试存款负数 ——————
	fmt.Println("\n===== 测试存款负数 =====")
	if err := bob.Deposit(-10); err != nil { // 预期失败
		fmt.Println("存款负数:", err)
	}

	// —————— 4. 测试取款 ——————
	fmt.Println("\n===== 测试取款 =====")
	if err := bob.Withdraw(20); err != nil {
		fmt.Println("取款失败:", err)
	} else {
		fmt.Printf("Bob 取出 20 → 余额: %.2f\n", bob.Balance)
	}

	// —————— 5. 测试余额不足 ——————
	fmt.Println("\n===== 测试余额不足 =====")
	if err := charlie.Withdraw(999); err != nil { // 预期失败
		fmt.Println("超额取款:", err)
	}

	// —————— 6. 测试单笔转账 ——————
	fmt.Println("\n===== 测试转账 =====")
	if err := alice.Transfer(bob, 30); err != nil {
		fmt.Println("转账失败:", err)
	} else {
		fmt.Printf("Alice → Bob 30 ETH\n")
		fmt.Printf("  Alice: %.2f\n", alice.Balance)
		fmt.Printf("  Bob:   %.2f\n", bob.Balance)
	}

	// —————— 7. 测试批量转账 ——————
	fmt.Println("\n===== 测试批量转账 =====")
	recipients := []*Wallet{bob, charlie}           // 收款人切片，装的是指针
	err := BatchTransfer(alice, recipients, 10)      // 每人转10
	if err != nil {
		fmt.Println("批量转账失败:", err)
	}
	fmt.Printf("Alice:  %.2f (扣了 %d 人的钱)\n", alice.Balance, len(recipients))
	fmt.Printf("Bob:    %.2f\n", bob.Balance)
	fmt.Printf("Charlie: %.2f\n", charlie.Balance)

	// —————— 8. 测试批量转账回滚 ——————
	fmt.Println("\n===== 测试批量转账回滚 =====")
	fmt.Println("转账前:")
	fmt.Printf("  Alice: %.2f | Bob: %.2f | Charlie: %.2f\n",
		alice.Balance, bob.Balance, charlie.Balance)

	// 故意让 alice 没有足够余额完成所有转账，触发回滚
	err = BatchTransfer(alice, []*Wallet{bob, charlie}, 1000)
	if err != nil {
		fmt.Println("批量转账失败（预期）:", err)
	}
	fmt.Println("转账后（应和转账前一致）:")
	fmt.Printf("  Alice: %.2f | Bob: %.2f | Charlie: %.2f\n",
		alice.Balance, bob.Balance, charlie.Balance)

	// —————— 9. 测试目标钱包为 nil ——————
	fmt.Println("\n===== 测试 nil 钱包 =====")
	if err := alice.Transfer(nil, 10); err != nil { // 目标为 nil 指针
		fmt.Println("nil 目标:", err)
	}
}
