package main

import "fmt"

//type Wallet struct {
//	Balance float64
//}
//
//func depositValue(w Wallet, amount float64) {
//	w.Balance += amount
//	fmt.Println("函数内余额:", w.Balance)
//}
//
//func main() {
//	wallet := Wallet{Balance: 100.0}
//	depositValue(wallet, 100.0)
//	fmt.Println("函数外余额:", wallet.Balance)
//}

//### 值传递（复制）
//type Wallet struct {
//	Balance float64
//}
//
//func depositValue(w Wallet, amount float64) {
//	w.Balance += amount
//	fmt.Println("函数内余额:", w.Balance)
//}
//
//func main() {
//	wallet := Wallet{Balance: 100.0}
//	depositValue(wallet, 50.0)
//	fmt.Println("函数外余额:", wallet.Balance)
//}

// ### 指针传递（引用）⭐
type Wallet struct {
	Balance float64
}

// 指针传递：会修改原始数据
func depositPointer(w *Wallet, amount float64) {
	w.Balance += amount
	fmt.Println("函数内余额:", w.Balance)
}

func main() {
	wallet := Wallet{Balance: 100.0}

	depositPointer(&wallet, 50.0)
	fmt.Println("函数外余额:", wallet.Balance)
}
