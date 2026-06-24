# Gin 实战 — 模拟数据库 map 详解

## 代码

```go
var wallets = map[string]*models.Wallet{
    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb": {
        Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
        Balance: 10.5,
        Network: "Ethereum",
    },
}
```

## 1. 整体作用

这是一个**包级变量**，用 Go 内置的 `map` 充当内存中的"假数据库"，以钱包地址为键来存储和查询钱包数据。

## 2. 逐层拆解

```
var wallets = map[string]*models.Wallet{ ... }
     ↓              ↓           ↓         ↓
  变量名         map 类型    值的类型   字面量初始化
```

### 第一层：变量声明

| 部分 | 写法 | 含义 |
|------|------|------|
| 关键字 | `var` | 声明一个包级变量（在整个 `handlers` 包内可用） |
| 变量名 | `wallets` | 小写开头，包内私有（只在 `handlers` 包内能访问） |

### 第二层：map 类型

```
map[string]*models.Wallet
 ↓       ↓         ↓
map   键的类型   值的类型
```

| 元素 | 说明 |
|------|------|
| `map` | Go 内置的哈希映射（字典/关联数组），通过键快速查找值 |
| `string` | **键的类型**：这里是钱包地址，如 `"0x742d..."` |
| `*models.Wallet` | **值的类型**：指向 `models.Wallet` 结构体的**指针** |

### 第三层：字面量初始化

```go
{
    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb": {  // 键：钱包地址
        Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
        Balance: 10.5,
        Network: "Ethereum",
    },
}
```

- `{ }` 内是 map 的初始数据（此处初始有一条钱包记录）
- 每条记录用 `key: value` 格式
- 值是一个 `models.Wallet` 结构体字面量，字段名直接赋值
- 末尾的 `,` 是 Go 语法要求（多行字面量最后一个元素必须加逗号）

## 3. 为什么使用指针 `*models.Wallet`？

```go
// 指针版本（当前代码）
map[string]*models.Wallet

// 对比：值版本
map[string]models.Wallet
```

| 对比维度 | `*models.Wallet`（指针） | `models.Wallet`（值） |
|----------|--------------------------|------------------------|
| 内存占用 | map 只存 8 字节指针（64 位系统） | map 直接存整个结构体（~48 字节） |
| 修改数据 | 从 map 取出后直接修改，会影响 map 中的原数据 | 从 map 取出的是**副本**，修改不影响 map |
| 取用后操作 | `wallet.Balance = 100` ✅ 生效 | `wallet.Balance = 100` ❌ 不生效（改的是副本） |

**示例对比：**

```go
// 指针版本 — 可以直接修改
wallet := wallets["0x742d..."]
wallet.Balance = 999.0   // ✅ map 里的值同步变化

// 值版本 — 修改无效
wallet := wallets["0x742d..."]
wallet.Balance = 999.0   // ❌ 只改了副本，map 里不变
```

> 💡 在 Gin 项目中使用指针，后续写转账/更新余额等接口时可以直接修改 map 中的钱包，无需再写回。

## 4. 在 Handler 中的使用方式

```go
// 查询钱包 — 利用 map 的 "comma ok" 惯用法
func GetWallet(c *gin.Context) {
    address := c.Param("address")      // 从 URL 取地址参数

    wallet, exists := wallets[address] // 查 map
    if !exists {                       // 不存在
        c.JSON(http.StatusNotFound, gin.H{"error": "钱包不存在"})
        return
    }

    c.JSON(http.StatusOK, wallet)      // 存在，返回钱包数据
}
```

| 步骤 | 代码 | 说明 |
|------|------|------|
| 取值 | `wallets[address]` | 用地址作键查 map |
| 判断 | `wallet, exists :=` | Go 的 "comma ok" 模式，`exists` 为 `bool` |
| 不存在 | `!exists` | 返回 404 错误 |
| 存在 | 直接 `c.JSON(..., wallet)` | 返回 JSON 响应 |

## 5. 对应 JSON 响应

当 GET 请求 `/wallet/0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb` 时，返回：

```json
{
    "address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "balance": 10.5,
    "network": "Ethereum"
}
```

字段名变成小写 —— 这是 `models.Wallet` 结构体中 `` `json:"address"` `` 标签的作用。

## 6. 跨文件协作全景

```
models/wallet.go              handlers/wallet.go
┌──────────────────┐          ┌──────────────────────────┐
│ type Wallet struct│          │ import "models"          │
│   Address string  │ ◄─────── │                          │
│   Balance float64 │  引用    │ var wallets = map[string]│
│   Network string  │          │     *models.Wallet{...} │
└──────────────────┘          │                          │
                              │ func GetWallet(c *gin) { │
                              │   wallets[address]       │
                              │   ...                    │
                              │ }                        │
                              └──────────────────────────┘
```

- `models` 包定义数据结构（"长什么样"）
- `handlers` 包持有数据实例并操作它（"存在哪、怎么用"）
- 通过 `*models.Wallet` 指针关联，两者解耦但协同工作

## 7. 总结

| 要点 | 一句话 |
|------|--------|
| `map[string]*models.Wallet` | 以地址为键、钱包指针为值的内存哈希表 |
| `var` 包级变量 | 整个 handlers 包都能读写这个"数据库" |
| 指针 `*` | 允许直接修改 map 中的数据，无需写回 |
| 字面量初始化 | 启动时就有一条示例数据 |
| "comma ok" 查询 | `wallet, exists := wallets[key]` 安全取值 |
