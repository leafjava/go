# 第5课：指针（重点）

> 学习时间：2-3小时 | 难度：⭐⭐⭐

## 📋 本课目标

- 理解指针的概念和作用
- 掌握指针的声明和使用
- 理解值传递和指针传递的区别
- 避免常见的指针错误

## 1. 指针基础

### 什么是指针？

```go
package main

import "fmt"

func main() {
    // 普通变量
    age := 23
    fmt.Println("age 的值:", age)
    fmt.Println("age 的地址:", &age)  // & 取地址符
    
    // 指针变量（存储地址）
    var ptr *int = &age  // * 声明指针类型
    fmt.Println("ptr 的值（地址）:", ptr)
    fmt.Println("ptr 指向的值:", *ptr)  // * 解引用
    
    // 通过指针修改值
    *ptr = 24
    fmt.Println("修改后 age 的值:", age)  // 24
}
```

### 指针语法

```go
package main

import "fmt"

func main() {
    // 声明指针
    var p1 *int        // nil 指针
    var p2 *string
    var p3 *float64
    
    fmt.Println(p1, p2, p3)  // <nil> <nil> <nil>
    
    // 创建指针
    name := "林燊"
    p2 = &name
    
    balance := 100.5
    p3 = &balance
    
    // 使用 new 创建指针
    p1 = new(int)
    *p1 = 42
    
    fmt.Println(*p1, *p2, *p3)
}
```

## 2. 值传递 vs 指针传递

### 值传递（复制）

```go
package main

import "fmt"

type Wallet struct {
    Balance float64
}

// 值传递：不会修改原始数据
func depositValue(w Wallet, amount float64) {
    w.Balance += amount
    fmt.Println("函数内余额:", w.Balance)
}

func main() {
    wallet := Wallet{Balance: 100.0}
    
    depositValue(wallet, 50.0)
    fmt.Println("函数外余额:", wallet.Balance)  // 100.0（未改变）
}
```

### 指针传递（引用）⭐

```go
package main

import "fmt"

type Wallet struct {
    Balance float64
}

// 指针传递：会修改原始数据
func depositPointer(w *Wallet, amount float64) {
    w.Balance += amount  // 自动解引用，等价于 (*w).Balance
    fmt.Println("函数内余额:", w.Balance)
}

func main() {
    wallet := Wallet{Balance: 100.0}
    
    depositPointer(&wallet, 50.0)
    fmt.Println("函数外余额:", wallet.Balance)  // 150.0（已改变）
}
```

### Go vs Java 对比

```go
// Java 方式（对象默认是引用传递）
/*
public void deposit(Wallet wallet, double amount) {
    wallet.balance += amount;  // 会修改原对象
}
*/

// Go 方式（需要显式使用指针）
func deposit(w *Wallet, amount float64) {
    w.Balance += amount
}
```

## 3. Web3 实战：指针应用

### 钱包操作

```go
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
        return fmt.Errorf("余额不足: 需要 %.2f, 当前 %.2f", amount, w.Balance)
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
        // 回滚
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
```

### 交易池管理

```go
package main

import "fmt"

type Transaction struct {
    Hash   string
    From   string
    To     string
    Amount float64
    Status string
}

type TransactionPool struct {
    Transactions []*Transaction  // 指针切片
    MaxSize      int
}

// 添加交易
func (tp *TransactionPool) AddTransaction(tx *Transaction) error {
    if len(tp.Transactions) >= tp.MaxSize {
        return fmt.Errorf("交易池已满: %d/%d", len(tp.Transactions), tp.MaxSize)
    }
    tp.Transactions = append(tp.Transactions, tx)
    return nil
}

// 确认交易
func (tp *TransactionPool) ConfirmTransaction(hash string) bool {
    for _, tx := range tp.Transactions {
        if tx.Hash == hash {
            tx.Status = "confirmed"  // 直接修改原交易
            return true
        }
    }
    return false
}

// 获取待处理交易
func (tp *TransactionPool) GetPendingTransactions() []*Transaction {
    var pending []*Transaction
    for _, tx := range tp.Transactions {
        if tx.Status == "pending" {
            pending = append(pending, tx)
        }
    }
    return pending
}

func main() {
    pool := &TransactionPool{
        Transactions: make([]*Transaction, 0),
        MaxSize:      100,
    }
    
    // 添加交易
    tx1 := &Transaction{
        Hash:   "0xabc123",
        From:   "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
        To:     "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
        Amount: 1.5,
        Status: "pending",
    }
    
    pool.AddTransaction(tx1)
    
    // 确认交易
    pool.ConfirmTransaction("0xabc123")
    
    fmt.Println("交易状态:", tx1.Status)  // confirmed
}
```

## 4. 指针和切片、Map

### 切片中的指针

```go
package main

import "fmt"

type User struct {
    ID   int
    Name string
}

func main() {
    // 值切片
    users1 := []User{
        {ID: 1, Name: "Alice"},
        {ID: 2, Name: "Bob"},
    }
    
    // 修改切片元素（会修改原数据）
    users1[0].Name = "Alice Updated"
    fmt.Println(users1[0].Name)  // Alice Updated
    
    // 指针切片（更常用）⭐
    users2 := []*User{
        {ID: 1, Name: "Alice"},
        {ID: 2, Name: "Bob"},
    }
    
    // 修改指针指向的数据
    users2[0].Name = "Alice Updated"
    fmt.Println(users2[0].Name)  // Alice Updated
}
```

### Map 中的指针

```go
package main

import "fmt"

type Wallet struct {
    Address string
    Balance float64
}

func main() {
    // Map 的值是指针
    wallets := make(map[string]*Wallet)
    
    wallets["user1"] = &Wallet{
        Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
        Balance: 10.0,
    }
    
    // 直接修改 Map 中的值
    wallets["user1"].Balance += 5.0
    fmt.Println("余额:", wallets["user1"].Balance)  // 15.0
}
```

## 5. 常见错误和注意事项

### 错误1：nil 指针解引用

```go
package main

import "fmt"

func main() {
    var p *int
    
    // 错误：nil 指针解引用会 panic
    // fmt.Println(*p)  // panic: runtime error
    
    // 正确：先检查是否为 nil
    if p != nil {
        fmt.Println(*p)
    } else {
        fmt.Println("指针为 nil")
    }
    
    // 使用 new 初始化
    p = new(int)
    *p = 42
    fmt.Println(*p)  // 42
}
```

### 错误2：返回局部变量的指针

```go
package main

import "fmt"

// 错误示例（在某些语言中）
func createWallet() *Wallet {
    wallet := Wallet{
        Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
        Balance: 10.0,
    }
    return &wallet  // Go 中是安全的（逃逸分析）
}

type Wallet struct {
    Address string
    Balance float64
}

func main() {
    w := createWallet()
    fmt.Println(w.Balance)  // 10.0（Go 会自动处理）
}
```

### 错误3：指针比较

```go
package main

import "fmt"

type User struct {
    ID   int
    Name string
}

func main() {
    user1 := &User{ID: 1, Name: "Alice"}
    user2 := &User{ID: 1, Name: "Alice"}
    
    // 比较指针地址（不同）
    fmt.Println(user1 == user2)  // false
    
    // 比较指针指向的值（相同）
    fmt.Println(*user1 == *user2)  // true
}
```

## 6. 何时使用指针？⭐

### 使用指针的场景

1. **需要修改数据**
```go
func (w *Wallet) Deposit(amount float64) {
    w.Balance += amount
}
```

2. **避免大结构体复制**
```go
type LargeStruct struct {
    Data [1000000]int
}

func process(ls *LargeStruct) {  // 避免复制 1MB 数据
    // ...
}
```

3. **需要表示"不存在"**
```go
func findUser(id int) *User {
    // 找不到返回 nil
    return nil
}
```

### 不使用指针的场景

1. **小的不可变数据**
```go
func add(a, b int) int {  // int 很小，直接传值
    return a + b
}
```

2. **不需要修改的数据**
```go
func (w Wallet) GetBalance() float64 {  // 只读操作
    return w.Balance
}
```

## 📝 作业

### 作业1：钱包管理系统

创建 `homework/day5/wallet_system.go`：

```go
package main

import "fmt"

type Wallet struct {
    Address string
    Balance float64
}

// TODO: 实现以下方法（使用指针接收者）
// 1. Deposit(amount float64) error
// 2. Withdraw(amount float64) error
// 3. Transfer(to *Wallet, amount float64) error

// TODO: 实现批量转账函数
func BatchTransfer(from *Wallet, recipients []*Wallet, amount float64) error {
    // 从 from 向每个 recipient 转账 amount
    // 如果任何一笔失败，回滚所有已完成的转账
    return nil
}

func main() {
    // 测试你的代码
}
```

### 作业2：交易池优化

创建 `homework/day5/tx_pool.go`：

```go
package main

import "fmt"

type Transaction struct {
    Hash      string
    From      string
    To        string
    Amount    float64
    GasPrice  float64
    Status    string
}

type TxPool struct {
    Pending   []*Transaction
    Confirmed []*Transaction
    Failed    []*Transaction
}

// TODO: 实现以下方法
// 1. AddPending(tx *Transaction)
// 2. ConfirmTransaction(hash string) bool
// 3. FailTransaction(hash string) bool
// 4. GetHighestGasPriceTx() *Transaction
// 5. RemoveConfirmedTransactions()

func main() {
    // 测试交易池
}
```

### 作业3：NFT 所有权转移

创建 `homework/day5/nft_transfer.go`：

```go
package main

import "fmt"

type NFT struct {
    TokenID int
    Name    string
    Owner   *User  // 指针指向所有者
}

type User struct {
    Address string
    NFTs    []*NFT  // 用户拥有的 NFT
}

// TODO: 实现 NFT 转移函数
func TransferNFT(nft *NFT, from, to *User) error {
    // 1. 验证 from 是当前所有者
    // 2. 从 from.NFTs 中移除
    // 3. 添加到 to.NFTs
    // 4. 更新 nft.Owner
    return nil
}

// TODO: 实现批量转移函数
func BatchTransferNFTs(nfts []*NFT, from, to *User) error {
    // 转移多个 NFT
    return nil
}

func main() {
    // 测试 NFT 转移
}
```

### 作业4：内存优化对比

创建 `homework/day5/memory_comparison.go`：

```go
package main

import (
    "fmt"
    "time"
)

type LargeData struct {
    Data [100000]int
}

// 值传递
func processByValue(data LargeData) {
    // 模拟处理
    _ = data.Data[0]
}

// 指针传递
func processByPointer(data *LargeData) {
    // 模拟处理
    _ = data.Data[0]
}

func main() {
    data := LargeData{}
    
    // TODO: 测试值传递的时间
    start := time.Now()
    for i := 0; i < 1000; i++ {
        processByValue(data)
    }
    fmt.Println("值传递耗时:", time.Since(start))
    
    // TODO: 测试指针传递的时间
    start = time.Now()
    for i := 0; i < 1000; i++ {
        processByPointer(&data)
    }
    fmt.Println("指针传递耗时:", time.Since(start))
}
```

## 🎯 检查点

完成本课后，你应该能够：
- ✅ 理解指针的概念和作用
- ✅ 正确使用 & 和 * 操作符
- ✅ 区分值传递和指针传递
- ✅ 在合适的场景使用指针
- ✅ 避免常见的指针错误

## 💡 重点提示

1. **指针是 Go 的核心特性**，必须熟练掌握
2. **方法接收者优先使用指针**（除非有特殊原因）
3. **切片、Map、Channel 本身就是引用类型**，不需要额外使用指针
4. **nil 指针检查很重要**，避免 panic

## 📚 扩展阅读

- [Go by Example - Pointers](https://gobyexample.com/pointers)
- [Go 语言圣经 - 第2.3节](https://gopl-zh.github.io/ch2/ch2-03.html)
- [Effective Go - Pointers vs Values](https://go.dev/doc/effective_go#pointers_vs_values)

## ⏭️ 下一课

[第6课：Slice 和 Map](./day6-collections.md)
