package main

import "fmt"

type Wallet struct {
	Address string
	Balance float64
}

// 2. 方法 - 接收者在 func 和函数名之间
// (w Wallet) 是接收者（Receiver）
// 这是一个方法，属于 Wallet 类型
// 调用时必须通过 Wallet 实例

// 记忆技巧
// 接收者在前 = "这个方法属于谁"
// 参数在后 = "这个方法需要什么输入"
func (w Wallet) GetBalance() float64 {
	//^^^^^^^^ 接收者（表示这个方法属于谁）
	return w.Balance
}

func (w *Wallet) Deposit(amount float64) {
	w.Balance += amount
}

func (w *Wallet) Withdraw(amount float64) error {
	if amount > w.Balance {
		return fmt.Errorf("余额不足:需要%.2f,当前%.2f", amount, w.Balance)
	}
	w.Balance -= amount
	return nil
}

func (w *Wallet) Transfer(to string, amount float64) error {
	if err := w.Withdraw(amount); err != nil {
		return err
	}

	fmt.Printf("转账 %.2f ETH 到 %s\n", amount, to)
	return nil
}

func main() {
	wallet := &Wallet{
		Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
		Balance: 1000,
	}

	fmt.Println("初始余额:", wallet.GetBalance())

	wallet.Deposit(200)
	fmt.Println("存款后余额:", wallet.GetBalance())

	err := wallet.Withdraw(3.0)
	if err != nil {
		fmt.Println("取款失败:", err)
	} else {
		fmt.Println("取款后余额:", wallet.GetBalance())
	}

	wallet.Transfer("0x8ba1f109551bD432803012645Ac136ddd64DBA72", 2.0)
	fmt.Println("转账后余额:", wallet.GetBalance())
}
