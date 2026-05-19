# 第4课：结构体、方法、接口

> 学习时间：3-4小时 | 难度：⭐⭐⭐

## 📋 本课目标

- 掌握结构体的定义和使用
- 理解方法和接收者
- 掌握接口的定义和实现
- 学会组合和嵌入

## 1. 结构体（Struct）

### 基本定义

```go
package main

import "fmt"

// 定义结构体（类似 Java 的 class）
type User struct {
    ID       int
    Name     string
    Email    string
    Age      int
    IsActive bool
}

// 嵌套结构体
type Address struct {
    City    string
    Country string
}

type UserWithAddress struct {
    ID      int
    Name    string
    Address Address  // 嵌套
}

func main() {
    // 方式1：字段名初始化（推荐）
    user1 := User{
        ID:       1,
        Name:     "林燊",
        Email:    "linshen@example.com",
        Age:      23,
        IsActive: true,
    }
    
    // 方式2：按顺序初始化
    user2 := User{2, "张三", "zhangsan@example.com", 25, true}
    
    // 方式3：部分初始化（其他字段为零值）
    user3 := User{
        ID:   3,
        Name: "李四",
    }
    
    fmt.Println(user1)
    fmt.Println(user2)
    fmt.Println(user3)
    
    // 访问字段
    fmt.Println("用户名:", user1.Name)
    fmt.Println("邮箱:", user1.Email)
    
    // 修改字段
    user1.Age = 24
    fmt.Println("新年龄:", user1.Age)
}
```

### Web3 实战：钱包结构体

```go
package main

import "fmt"

// 钱包结构体
type Wallet struct {
    Address    string
    PrivateKey string
    Balance    float64
    Network    string
}

// 交易结构体
type Transaction struct {
    Hash      string
    From      string
    To        string
    Amount    float64
    GasUsed   int64
    Status    string
    Timestamp int64
}

// NFT 结构体
type NFT struct {
    TokenID     int
    Name        string
    Description string
    ImageURL    string
    Owner       string
    Contract    string
}

func main() {
    // 创建钱包
    wallet := Wallet{
        Address:    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
        PrivateKey: "0x...",  // 实际应该加密存储
        Balance:    1.5,
        Network:    "Ethereum",
    }
    
    // 创建交易
    tx := Transaction{
        Hash:      "0xabc123...",
        From:      wallet.Address,
        To:        "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
        Amount:    0.5,
        GasUsed:   21000,
        Status:    "confirmed",
        Timestamp: 1704067200,
    }
    
    fmt.Printf("钱包: %s, 余额: %.2f ETH\n", wallet.Address, wallet.Balance)
    fmt.Printf("交易: %s, 金额: %.2f ETH\n", tx.Hash, tx.Amount)
}
```

## 2. 方法（Methods）

### 值接收者 vs 指针接收者

```go
package main

import "fmt"

type Wallet struct {
    Address string
    Balance float64
}

// 值接收者：不会修改原始数据
func (w Wallet) GetBalance() float64 {
    return w.Balance
}

// 指针接收者：可以修改原始数据（推荐）⭐
func (w *Wallet) Deposit(amount float64) {
    w.Balance += amount
}

func (w *Wallet) Withdraw(amount float64) error {
    if amount > w.Balance {
        return fmt.Errorf("余额不足: 需要 %.2f, 当前 %.2f", amount, w.Balance)
    }
    w.Balance -= amount
    return nil
}

// 方法可以链式调用
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
        Balance: 10.0,
    }
    
    // 调用方法
    fmt.Println("初始余额:", wallet.GetBalance())
    
    wallet.Deposit(5.0)
    fmt.Println("存款后余额:", wallet.GetBalance())
    
    err := wallet.Withdraw(3.0)
    if err != nil {
        fmt.Println("取款失败:", err)
    } else {
        fmt.Println("取款后余额:", wallet.GetBalance())
    }
    
    // 转账
    wallet.Transfer("0x8ba1f109551bD432803012645Ac136ddd64DBA72", 2.0)
    fmt.Println("转账后余额:", wallet.GetBalance())
}
```

### Go vs Java 对比

```go
// Java 方式
/*
public class Wallet {
    private String address;
    private double balance;
    
    public double getBalance() {
        return balance;
    }
    
    public void deposit(double amount) {
        this.balance += amount;
    }
}
*/

// Go 方式
type Wallet struct {
    Address string
    Balance float64
}

func (w *Wallet) GetBalance() float64 {
    return w.Balance
}

func (w *Wallet) Deposit(amount float64) {
    w.Balance += amount
}
```

## 3. 接口（Interface）

### 基本接口

```go
package main

import "fmt"

// 接口定义（只声明方法，不实现）
type PaymentMethod interface {
    Pay(amount float64) error
    GetName() string
}

// 以太坊支付
type EthereumPayment struct {
    WalletAddress string
}

func (e *EthereumPayment) Pay(amount float64) error {
    fmt.Printf("使用以太坊支付 %.2f ETH\n", amount)
    return nil
}

func (e *EthereumPayment) GetName() string {
    return "Ethereum"
}

// TON 支付
type TONPayment struct {
    WalletAddress string
}

func (t *TONPayment) Pay(amount float64) error {
    fmt.Printf("使用 TON 支付 %.2f TON\n", amount)
    return nil
}

func (t *TONPayment) GetName() string {
    return "TON"
}

// 处理支付（接受任何实现了 PaymentMethod 的类型）
func processPayment(pm PaymentMethod, amount float64) {
    fmt.Printf("支付方式: %s\n", pm.GetName())
    if err := pm.Pay(amount); err != nil {
        fmt.Println("支付失败:", err)
    } else {
        fmt.Println("支付成功!")
    }
}

func main() {
    // 创建不同的支付方式
    ethPayment := &EthereumPayment{
        WalletAddress: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    }
    
    tonPayment := &TONPayment{
        WalletAddress: "EQD...",
    }
    
    // 使用相同的函数处理不同的支付方式
    processPayment(ethPayment, 1.5)
    fmt.Println("---")
    processPayment(tonPayment, 100)
}
```

### 空接口（interface{}）

```go
package main

import "fmt"

// 空接口可以接受任何类型（类似 Java 的 Object）
func printAnything(v interface{}) {
    fmt.Printf("类型: %T, 值: %v\n", v, v)
}

// 类型断言
func processValue(v interface{}) {
    switch val := v.(type) {
    case int:
        fmt.Println("整数:", val*2)
    case string:
        fmt.Println("字符串:", val+" World")
    case float64:
        fmt.Printf("浮点数: %.2f\n", val)
    default:
        fmt.Println("未知类型")
    }
}

func main() {
    printAnything(42)
    printAnything("Hello")
    printAnything(3.14)
    printAnything(true)
    
    fmt.Println("---")
    
    processValue(100)
    processValue("Hello")
    processValue(99.99)
}
```

### Web3 实战：区块链接口

```go
package main

import "fmt"

// 区块链接口
type Blockchain interface {
    GetBalance(address string) (float64, error)
    SendTransaction(from, to string, amount float64) (string, error)
    GetBlockNumber() (int64, error)
}

// 以太坊实现
type Ethereum struct {
    NodeURL string
}

func (e *Ethereum) GetBalance(address string) (float64, error) {
    // 模拟查询余额
    fmt.Println("查询以太坊余额:", address)
    return 1.5, nil
}

func (e *Ethereum) SendTransaction(from, to string, amount float64) (string, error) {
    // 模拟发送交易
    fmt.Printf("发送 %.2f ETH 从 %s 到 %s\n", amount, from, to)
    return "0xabc123...", nil
}

func (e *Ethereum) GetBlockNumber() (int64, error) {
    return 18500000, nil
}

// TON 实现
type TON struct {
    NodeURL string
}

func (t *TON) GetBalance(address string) (float64, error) {
    fmt.Println("查询 TON 余额:", address)
    return 100.0, nil
}

func (t *TON) SendTransaction(from, to string, amount float64) (string, error) {
    fmt.Printf("发送 %.2f TON 从 %s 到 %s\n", amount, from, to)
    return "abc123def456", nil
}

func (t *TON) GetBlockNumber() (int64, error) {
    return 35000000, nil
}

// 通用查询函数
func queryBlockchain(bc Blockchain, address string) {
    balance, err := bc.GetBalance(address)
    if err != nil {
        fmt.Println("查询失败:", err)
        return
    }
    
    blockNum, _ := bc.GetBlockNumber()
    fmt.Printf("余额: %.2f, 区块高度: %d\n", balance, blockNum)
}

func main() {
    // 创建不同的区块链实例
    eth := &Ethereum{NodeURL: "https://eth.llamarpc.com"}
    ton := &TON{NodeURL: "https://toncenter.com/api/v2"}
    
    // 使用相同的函数查询不同的链
    queryBlockchain(eth, "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
    fmt.Println("---")
    queryBlockchain(ton, "EQD...")
}
```

## 4. 结构体嵌入（组合）

```go
package main

import "fmt"

// 基础结构体
type BaseModel struct {
    ID        int
    CreatedAt int64
    UpdatedAt int64
}

// 嵌入 BaseModel（继承字段）
type User struct {
    BaseModel  // 匿名字段（嵌入）
    Name  string
    Email string
}

type Product struct {
    BaseModel
    Name  string
    Price float64
}

func main() {
    user := User{
        BaseModel: BaseModel{
            ID:        1,
            CreatedAt: 1704067200,
            UpdatedAt: 1704067200,
        },
        Name:  "林燊",
        Email: "linshen@example.com",
    }
    
    // 可以直接访问嵌入结构体的字段
    fmt.Println("用户ID:", user.ID)
    fmt.Println("用户名:", user.Name)
    fmt.Println("创建时间:", user.CreatedAt)
}
```

## 📝 作业

### 作业1：NFT 市场系统

创建 `homework/day4/nft_marketplace.go`：

```go
package main

import "fmt"

// TODO: 定义 NFT 结构体
type NFT struct {
    // TokenID, Name, Owner, Price, IsListed
}

// TODO: 实现 NFT 方法
// 1. List(price float64) - 上架
// 2. Unlist() - 下架
// 3. Transfer(newOwner string) - 转移
// 4. GetInfo() string - 获取信息

// TODO: 定义 Marketplace 结构体
type Marketplace struct {
    // NFTs map[int]*NFT
    // TotalSales float64
}

// TODO: 实现 Marketplace 方法
// 1. AddNFT(nft *NFT)
// 2. BuyNFT(tokenID int, buyer string) error
// 3. GetListedNFTs() []*NFT
// 4. GetTotalSales() float64

func main() {
    // 测试你的代码
}
```

### 作业2：多链钱包管理器

创建 `homework/day4/multi_chain_wallet.go`：

```go
package main

import "fmt"

// TODO: 定义 ChainWallet 接口
type ChainWallet interface {
    // GetBalance() float64
    // Deposit(amount float64)
    // Withdraw(amount float64) error
    // GetChainName() string
}

// TODO: 实现 EthWallet 结构体和方法

// TODO: 实现 TONWallet 结构体和方法

// TODO: 实现 MultiChainWallet 结构体
type MultiChainWallet struct {
    // wallets map[string]ChainWallet
}

// TODO: 实现方法
// 1. AddWallet(chain string, wallet ChainWallet)
// 2. GetTotalBalance() float64
// 3. GetWallet(chain string) (ChainWallet, error)

func main() {
    // 测试多链钱包
}
```

### 作业3：DeFi 协议接口

创建 `homework/day4/defi_protocol.go`：

```go
package main

import "fmt"

// TODO: 定义 DeFiProtocol 接口
type DeFiProtocol interface {
    // Deposit(amount float64) error
    // Withdraw(amount float64) error
    // GetAPY() float64
    // GetTVL() float64
}

// TODO: 实现 Uniswap 结构体（模拟）
// TODO: 实现 Aave 结构体（模拟）
// TODO: 实现 Compound 结构体（模拟）

// TODO: 实现聚合器函数
func findBestAPY(protocols []DeFiProtocol) DeFiProtocol {
    // 找到 APY 最高的协议
    return nil
}

func main() {
    // 测试 DeFi 协议
}
```

### 作业4：交易记录系统

创建 `homework/day4/transaction_system.go`：

```go
package main

import "fmt"

// TODO: 定义 BaseTransaction 基础结构体
type BaseTransaction struct {
    // Hash, From, To, Amount, Timestamp
}

// TODO: 定义 EthTransaction（嵌入 BaseTransaction）
type EthTransaction struct {
    // BaseTransaction
    // GasUsed, GasPrice
}

// TODO: 实现方法
// 1. GetFee() float64 - 计算手续费
// 2. GetInfo() string - 获取交易信息

// TODO: 定义 TONTransaction（嵌入 BaseTransaction）
type TONTransaction struct {
    // BaseTransaction
    // ForwardFee
}

// TODO: 实现 TransactionManager
type TransactionManager struct {
    // transactions []interface{}
}

// TODO: 实现方法
// 1. AddTransaction(tx interface{})
// 2. GetTotalFees() float64
// 3. GetTransactionCount() int

func main() {
    // 测试交易系统
}
```

## 🎯 检查点

完成本课后，你应该能够：
- ✅ 定义和使用结构体
- ✅ 为结构体添加方法
- ✅ 理解值接收者和指针接收者的区别
- ✅ 定义和实现接口
- ✅ 使用结构体嵌入实现组合

## 💡 重点提示

1. **指针接收者 vs 值接收者**：
   - 需要修改数据 → 用指针接收者
   - 结构体较大 → 用指针接收者（避免复制）
   - 只读操作 → 可以用值接收者

2. **接口是隐式实现的**：
   - 不需要 `implements` 关键字
   - 只要实现了接口的所有方法就自动实现了接口

3. **组合优于继承**：
   - Go 没有继承，使用嵌入实现代码复用

## 📚 扩展阅读

- [Go by Example - Structs](https://gobyexample.com/structs)
- [Go by Example - Methods](https://gobyexample.com/methods)
- [Go by Example - Interfaces](https://gobyexample.com/interfaces)
- [Go 语言圣经 - 第6章](https://gopl-zh.github.io/ch6/ch6.html)

## ⏭️ 下一课

[第5课：指针（重点）](./day5-pointers.md)
