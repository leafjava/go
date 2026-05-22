package main

import "fmt"

// ### 切片中的指针
//type User struct {
//	ID   int
//	Name string
//}
//
//func main() {
//	// 值切片
//	users1 := []User{
//		{ID: 1, Name: "Alice"},
//		{ID: 2, Name: "Bob"},
//	}
//
//	// 修改切片元素（会修改原数据）
//	users1[0].Name = "Alice"
//	fmt.Println(users1[0].Name)
//
//	// 指针切片（更常用）⭐
//	users2 := []*User{
//		{ID: 1, Name: "Alice"},
//		{ID: 2, Name: "Bob"},
//	}
//
//	// 修改指针指向的数据
//	users2[0].Name = "Bob"
//	fmt.Println(users2[0].Name)
//}

// ### Map 中的指针
type Wallet struct {
	Address string
	Balance float64
}

func main() {
	wallets := make(map[string]*Wallet)

	wallets["user1"] = &Wallet{
		Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
		Balance: 10.0,
	}

	// 直接修改 Map 中的值
	wallets["user1"].Balance += 5.0
	fmt.Println("余额:", wallets["user1"].Balance) // 15.0
}

//make(map[string]*Wallet)
//make — Go 内置函数，用来创建 map、切片、channel（这三种类型必须初始化才能用）
//map[string]*Wallet — 这是一个 map 类型
//string 是 key 的类型（键）
//*Wallet 是 value 的类型（值），指针指向 Wallet 结构体
//make(...) 返回一个已经初始化、可以直接用的 map
//类比 JS：
//
//
//const wallets = new Map()        // make 就相当于 new Map()
//const wallets = {}               // 或者空对象字面量

//wallets["user1"] = &Wallet{...}
//这不就是 JS 里最基础的写法嘛：
//
//
//// Go
//wallets["user1"] = &Wallet{ Address: "...", Balance: 10.0 }
//
//// JS
//wallets["user1"] = { address: "...", balance: 10.0 }
//wallets["user1"] — 用 key "user1" 从 map 里取值/赋值，就是方括号取键
//&Wallet{} — 创建一个 Wallet 并取它的指针（& 是取地址）
//& 存在的理由和上次说的指针一样——如果不用 &，赋值时会拷贝一整份 Wallet，之后通过 wallets["user1"] 修改 Balance 不会影响别处持有的同一个 Wallet。用指针就大家指向同一块内存，改一处到处生效。
