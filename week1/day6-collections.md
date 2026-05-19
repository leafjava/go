# 第6课：Slice 和 Map

> 学习时间：2-3小时 | 难度：⭐⭐

## 📋 本课目标

- 掌握 Slice（切片）的使用
- 掌握 Map（映射）的使用
- 理解 Slice 和 Array 的区别
- 学会常用的集合操作

## 1. Array（数组）

```go
package main

import "fmt"

func main() {
    // 数组：固定长度
    var arr1 [5]int
    arr1[0] = 10
    arr1[1] = 20
    
    // 初始化数组
    arr2 := [3]string{"Alice", "Bob", "Charlie"}
    
    // 自动推断长度
    arr3 := [...]int{1, 2, 3, 4, 5}
    
    fmt.Println(arr1, arr2, arr3)
    fmt.Println("长度:", len(arr3))
}
```

## 2. Slice（切片）⭐⭐⭐

### 基本操作

```go
package main

import "fmt"

func main() {
    // 创建切片
    var s1 []int  // nil 切片
    s2 := []int{}  // 空切片
    s3 := []int{1, 2, 3, 4, 5}
    s4 := make([]int, 5)  // 长度为5
    s5 := make([]int, 5, 10)  // 长度5，容量10
    
    fmt.Println(s1, s2, s3, s4, s5)
    
    // 追加元素
    s3 = append(s3, 6, 7, 8)
    fmt.Println("追加后:", s3)
    
    // 切片操作
    fmt.Println("s3[1:4]:", s3[1:4])  // [2 3 4]
    fmt.Println("s3[:3]:", s3[:3])    // [1 2 3]
    fmt.Println("s3[3:]:", s3[3:])    // [4 5 6 7 8]
    
    // 长度和容量
    fmt.Println("长度:", len(s3), "容量:", cap(s3))
}
```

### Web3 实战：交易列表

```go
package main

import "fmt"

type Transaction struct {
    Hash   string
    Amount float64
    Status string
}

func main() {
    // 创建交易列表
    txs := make([]Transaction, 0, 10)
    
    // 添加交易
    txs = append(txs, Transaction{
        Hash:   "0xabc123",
        Amount: 1.5,
        Status: "pending",
    })
    
    txs = append(txs, Transaction{
        Hash:   "0xdef456",
        Amount: 2.0,
        Status: "confirmed",
    })
    
    // 遍历
    for i, tx := range txs {
        fmt.Printf("%d: %s - %.2f ETH (%s)\n", i, tx.Hash, tx.Amount, tx.Status)
    }
    
    // 过滤已确认的交易
    confirmed := make([]Transaction, 0)
    for _, tx := range txs {
        if tx.Status == "confirmed" {
            confirmed = append(confirmed, tx)
        }
    }
    
    fmt.Println("已确认交易数:", len(confirmed))
}
```

### Slice 常用操作

```go
package main

import "fmt"

func main() {
    addresses := []string{
        "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
        "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
        "0x1234567890123456789012345678901234567890",
    }
    
    // 1. 检查是否包含
    target := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
    found := false
    for _, addr := range addresses {
        if addr == target {
            found = true
            break
        }
    }
    fmt.Println("包含地址:", found)
    
    // 2. 删除元素（通过切片重组）
    indexToRemove := 1
    addresses = append(addresses[:indexToRemove], addresses[indexToRemove+1:]...)
    fmt.Println("删除后:", addresses)
    
    // 3. 插入元素
    newAddr := "0xNEWADDRESS"
    insertIndex := 1
    addresses = append(addresses[:insertIndex], append([]string{newAddr}, addresses[insertIndex:]...)...)
    fmt.Println("插入后:", addresses)
    
    // 4. 复制切片
    copied := make([]string, len(addresses))
    copy(copied, addresses)
    fmt.Println("复制:", copied)
}
```

## 3. Map（映射）⭐⭐⭐

### 基本操作

```go
package main

import "fmt"

func main() {
    // 创建 Map
    var m1 map[string]int  // nil map
    m2 := map[string]int{}  // 空 map
    m3 := map[string]int{
        "Alice": 100,
        "Bob":   200,
    }
    m4 := make(map[string]int)
    
    // 添加/修改
    m4["Charlie"] = 300
    m4["Alice"] = 150
    
    // 读取
    value := m4["Alice"]
    fmt.Println("Alice:", value)
    
    // 检查键是否存在
    value, exists := m4["David"]
    if exists {
        fmt.Println("David:", value)
    } else {
        fmt.Println("David 不存在")
    }
    
    // 删除
    delete(m4, "Alice")
    
    // 遍历
    for key, value := range m4 {
        fmt.Printf("%s: %d\n", key, value)
    }
    
    // 长度
    fmt.Println("Map 长度:", len(m4))
}
```

### Web3 实战：钱包余额管理

```go
package main

import "fmt"

func main() {
    // 地址 -> 余额
    balances := make(map[string]float64)
    
    // 添加余额
    balances["0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"] = 10.5
    balances["0x8ba1f109551bD432803012645Ac136ddd64DBA72"] = 5.2
    
    // 查询余额
    addr := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
    if balance, exists := balances[addr]; exists {
        fmt.Printf("地址 %s 余额: %.2f ETH\n", addr, balance)
    }
    
    // 转账
    from := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
    to := "0x8ba1f109551bD432803012645Ac136ddd64DBA72"
    amount := 2.0
    
    if balances[from] >= amount {
        balances[from] -= amount
        balances[to] += amount
        fmt.Println("转账成功")
    }
    
    // 统计总余额
    total := 0.0
    for _, balance := range balances {
        total += balance
    }
    fmt.Printf("总余额: %.2f ETH\n", total)
}
```

### Map 嵌套

```go
package main

import "fmt"

func main() {
    // 用户 -> (代币 -> 余额)
    userBalances := make(map[string]map[string]float64)
    
    // 初始化用户
    user1 := "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
    userBalances[user1] = make(map[string]float64)
    userBalances[user1]["ETH"] = 10.0
    userBalances[user1]["USDT"] = 1000.0
    userBalances[user1]["USDC"] = 500.0
    
    // 查询
    if tokens, exists := userBalances[user1]; exists {
        for token, balance := range tokens {
            fmt.Printf("%s: %.2f\n", token, balance)
        }
    }
}
```

## 4. Slice vs Map 选择

| 场景 | 使用 Slice | 使用 Map |
|------|-----------|----------|
| 有序数据 | ✅ | ❌ |
| 快速查找 | ❌ | ✅ |
| 键值对 | ❌ | ✅ |
| 遍历顺序 | 固定 | 随机 |
| 内存占用 | 较小 | 较大 |

## 📝 作业

### 作业1：交易过滤器

创建 `homework/day6/tx_filter.go`：

```go
package main

type Transaction struct {
    Hash   string
    From   string
    To     string
    Amount float64
    Status string
}

// TODO: 实现过滤函数
func FilterByStatus(txs []Transaction, status string) []Transaction {
    // 返回指定状态的交易
    return nil
}

func FilterByAmount(txs []Transaction, minAmount float64) []Transaction {
    // 返回金额大于等于 minAmount 的交易
    return nil
}

func FilterByAddress(txs []Transaction, address string) []Transaction {
    // 返回发送方或接收方包含该地址的交易
    return nil
}

func main() {
    // 测试过滤器
}
```

### 作业2：地址簿管理

创建 `homework/day6/address_book.go`：

```go
package main

type AddressBook struct {
    // TODO: 使用 Map 存储 名称 -> 地址
}

// TODO: 实现方法
// 1. Add(name, address string) error
// 2. Get(name string) (string, bool)
// 3. Remove(name string) bool
// 4. List() map[string]string
// 5. Search(keyword string) []string

func main() {
    // 测试地址簿
}
```

### 作业3：代币余额管理器

创建 `homework/day6/token_manager.go`：

```go
package main

type TokenManager struct {
    // TODO: 用户地址 -> (代币符号 -> 余额)
}

// TODO: 实现方法
// 1. AddBalance(user, token string, amount float64)
// 2. GetBalance(user, token string) float64
// 3. Transfer(from, to, token string, amount float64) error
// 4. GetUserTokens(user string) map[string]float64
// 5. GetTotalSupply(token string) float64

func main() {
    // 测试代币管理器
}
```

### 作业4：NFT 集合管理

创建 `homework/day6/nft_collection.go`：

```go
package main

type NFT struct {
    TokenID int
    Name    string
    Owner   string
}

type NFTCollection struct {
    // TODO: 使用 Slice 和 Map 管理 NFT
}

// TODO: 实现方法
// 1. Mint(tokenID int, name, owner string) error
// 2. Transfer(tokenID int, newOwner string) error
// 3. GetNFT(tokenID int) (*NFT, bool)
// 4. GetOwnerNFTs(owner string) []NFT
// 5. GetTotalSupply() int

func main() {
    // 测试 NFT 集合
}
```

## 🎯 检查点

- ✅ 熟练使用 Slice 的增删改查
- ✅ 熟练使用 Map 的增删改查
- ✅ 理解 Slice 和 Map 的适用场景
- ✅ 能够组合使用 Slice 和 Map

## ⏭️ 下一课

[第7课：并发编程：Goroutine + Channel](./day7-concurrency.md)
