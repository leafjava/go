package main

//错误1：nil 指针解引用
//func main() {
//	var p *int
//
//	// 错误：nil 指针解引用会 panic
//	// fmt.Println(*p)  // panic: runtime error
//
//	// 正确：先检查是否为 nil
//	if p != nil {
//		fmt.Println(*p)
//	} else {
//		fmt.Println("指针为 nil")
//	}
//
//	// 使用 new 初始化
//	p = new(int)
//	*p = 42
//	fmt.Println(*p) // 42
//}

//### 错误2：返回局部变量的指针
// 错误示例（在某些语言中）
//func createWallet() *Wallet {
//	wallet := Wallet{
//		Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
//		Balance: 10.0,
//	}
//	return &wallet  // Go 中是安全的（逃逸分析）
//}
//
//type Wallet struct {
//	Address string
//	Balance float64
//}
//
//func main() {
//	w := createWallet()
//	fmt.Println(w.Balance)  // 10.0（Go 会自动处理）
//}

//### 错误3：指针比较
//type User struct {
//	ID   int
//	Name string
//}
//
//func main() {
//	user1 := &User{ID: 1, Name: "Alice"}
//	user2 := &User{ID: 1, Name: "Alice"}
//
//	// 比较指针地址（不同）
//	fmt.Println(user1 == user2)  // false
//
//	// 比较指针指向的值（相同）
//	fmt.Println(*user1 == *user2)  // true
//}
