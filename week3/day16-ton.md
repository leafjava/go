# 第16课：TON 区块链集成

> 学习时间：3-4小时 | 难度：⭐⭐⭐⭐

## 📋 本课目标

- 理解 TON 区块链基本概念
- 掌握 tonutils-go 库的使用
- 学会连接 TON 节点和查询余额
- 实现 TON 转账和合约调用
- 掌握 TON 特有的地址和消息格式

## 1. TON 简介

TON（The Open Network）是一条高性能、可分片的区块链，由 Telegram 团队设计。与以太坊的核心区别：

| 特性 | 以太坊 | TON |
|------|--------|-----|
| 共识机制 | PoS（原 PoW） | PoS（BFT） |
| 分片 | 无原生分片 | 动态多分片（workchain + shardchain） |
| 智能合约 | Solidity/EVM | FunC/TVM |
| 地址格式 | 0x + 40位十六进制 | 48位 + crc16（base64） |
| 代币标准 | ERC-20 | Jettons |
| NFT 标准 | ERC-721 | TON NFT Standard |
| 交易模型 | 账户模型 | 异步消息模型 |
| Gas 费 | 市场竞价 | 固定费用 |

### 核心概念

- **Workchain（工作链）**：最多 2^32 条，目前只有 Masterchain（chain_id=-1）和 Basechain（chain_id=0）
- **Shardchain（分片链）**：每条 workchain 可自动拆分为最多 2^60 条分片链
- **Account（账户）**：每个智能合约和钱包都是一个账户
- **Message（消息）**：TON 的交易是异步消息传递，而非同步调用

## 2. 安装依赖

```bash
# 核心库
go get github.com/xssnick/tonutils-go

# 依赖库（会自动安装）
go get github.com/sigurn/crc16
go get github.com/oasisprotocol/curve25519-voi
```

## 3. 连接 TON 节点

### 基本连接

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/xssnick/tonutils-go/liteclient"
    "github.com/xssnick/tonutils-go/ton"
)

func main() {
    // 创建 LiteClient 连接池
    client := liteclient.NewConnectionPool()

    // 连接 TON 主网 LiteServer
    // 使用公开的 LiteServer 配置
    configURL := "https://ton.org/global.config.json"
    err := client.AddConnectionsFromConfigUrl(context.Background(), configURL)
    if err != nil {
        log.Fatal("连接 TON 节点失败:", err)
    }

    // 创建 TON API 客户端
    api := ton.NewAPIClient(client)

    fmt.Println("TON 节点连接成功")
    _ = api
}
```

### 使用 HTTP API（更简单的方式）

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/xssnick/tonutils-go/ton"
)

func main() {
    // 使用 TON HTTP API（适合轻量级应用）
    api := ton.NewAPIClient(ton.NewHTTPClient("https://toncenter.com/api/v2", ""))

    // 获取主链信息
    info, err := api.GetMasterchainInfo(context.Background())
    if err != nil {
        log.Fatal("获取主链信息失败:", err)
    }

    fmt.Printf("最新区块 Seqno: %d\n", info.Last.Seqno)
    fmt.Printf("最新区块 Hash: %x\n", info.Last.RootHash)
}
```

## 4. 钱包操作

### 创建钱包

```go
package main

import (
    "context"
    "fmt"
    "log"
    "strings"

    "github.com/xssnick/tonutils-go/ton/wallet"
    "github.com/xssnick/tonutils-go/liteclient"
)

func main() {
    client := liteclient.NewConnectionPool()
    err := client.AddConnectionsFromConfigUrl(context.Background(), "https://ton.org/global.config.json")
    if err != nil {
        log.Fatal(err)
    }

    api := ton.NewAPIClient(client)

    // 生成助记词（新钱包）
    words := wallet.NewSeed()
    fmt.Println("助记词（请安全保存）:")
    fmt.Println(strings.Join(words, " "))

    // 从助记词恢复钱包
    w, err := wallet.FromSeed(api, words, wallet.V4R2)
    if err != nil {
        log.Fatal("创建钱包失败:", err)
    }

    // 获取钱包地址
    address := w.Address()
    fmt.Printf("钱包地址（Bounceable）: %s\n", address.String())
    fmt.Printf("钱包地址（Non-Bounceable）: %s\n", address.Bounce(false).String())
}
```

### 查询余额

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/xssnick/tonutils-go/ton"
    "github.com/xssnick/tonutils-go/address"
    "github.com/xssnick/tonutils-go/liteclient"
)

func main() {
    client := liteclient.NewConnectionPool()
    err := client.AddConnectionsFromConfigUrl(context.Background(), "https://ton.org/global.config.json")
    if err != nil {
        log.Fatal(err)
    }

    api := ton.NewAPIClient(client)

    // 解析地址
    addr := address.MustParseAddr("EQCkR1cGZxYbG6QrGqKJ8HxSPfI3NZ9sNjCpHlLHyVTGc5Gq")

    // 获取账户状态
    account, err := api.GetAccount(context.Background(), addr)
    if err != nil {
        log.Fatal("查询账户失败:", err)
    }

    // TON 余额（单位：nanoTON，1 TON = 10^9 nanoTON）
    balanceTON := float64(account.State.Balance.Nano().Uint64()) / 1e9
    fmt.Printf("地址: %s\n", addr.String())
    fmt.Printf("余额: %.9f TON\n", balanceTON)
    fmt.Printf("状态: %s\n", account.State.Status)
}
```

## 5. 发送转账

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/xssnick/tonutils-go/ton"
    "github.com/xssnick/tonutils-go/ton/wallet"
    "github.com/xssnick/tonutils-go/address"
    "github.com/xssnick/tonutils-go/liteclient"
)

func main() {
    client := liteclient.NewConnectionPool()
    err := client.AddConnectionsFromConfigUrl(context.Background(), "https://ton.org/global.config.json")
    if err != nil {
        log.Fatal(err)
    }

    api := ton.NewAPIClient(client)

    // 从助记词恢复钱包
    words := strings.Split("your mnemonic words here", " ")
    w, err := wallet.FromSeed(api, words, wallet.V4R2)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("发送方地址: %s\n", w.Address().String())

    // 接收方地址
    toAddr := address.MustParseAddr("EQCkR1cGZxYbG6QrGqKJ8HxSPfI3NZ9sNjCpHlLHyVTGc5Gq")

    // 转账 0.05 TON
    // 注意：转账时附带 comment（备注）
    err = w.Send(context.Background(), &wallet.Message{
        Mode: wallet.MsgWithRemainingValue, // 128 + 32 = 160
        InternalMessage: &wallet.InternalMessage{
            IHRDisabled: false,
            Bounce:      false,
            Bounced:     false,
            SrcAddr:     w.WalletAddress(),
            DstAddr:     toAddr,
            Amount:      ton.MustToNano("0.05"),
            StateInit:   nil,
            Body:        wallet.Comment("转账测试"),
        },
    }, true) // true 表示等待确认

    if err != nil {
        log.Fatal("转账失败:", err)
    }

    fmt.Println("转账成功!")

    // 获取当前余额
    balance, err := w.GetBalance(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("当前余额: %.9f TON\n", float64(balance.Nano().Uint64())/1e9)
}
```

## 6. Jetton 代币操作

TON 上的代币称为 Jetton，相当于以太坊的 ERC-20。

### 查询 Jetton 余额

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math/big"

    "github.com/xssnick/tonutils-go/ton"
    "github.com/xssnick/tonutils-go/ton/jetton"
    "github.com/xssnick/tonutils-go/address"
    "github.com/xssnick/tonutils-go/liteclient"
)

func main() {
    client := liteclient.NewConnectionPool()
    err := client.AddConnectionsFromConfigUrl(context.Background(), "https://ton.org/global.config.json")
    if err != nil {
        log.Fatal(err)
    }

    api := ton.NewAPIClient(client)

    // 用户钱包地址
    userAddr := address.MustParseAddr("EQCkR1cGZxYbG6QrGqKJ8HxSPfI3NZ9sNjCpHlLHyVTGc5Gq")

    // USDT Jetton Master 合约地址（主网）
    // 实际地址请查询 https://tonviewer.com
    jettonMaster := address.MustParseAddr("EQCxE6mUtQJKFnGfaROTKOt1lZtbDkxXqWn4qBKJdCNL1QO-")

    // 创建 Jetton 客户端
    jettonClient := jetton.NewJettonMasterClient(api, jettonMaster)

    // 查询用户持有的 Jetton 余额
    balance, err := jettonClient.GetJettonWalletBalance(context.Background(), userAddr)
    if err != nil {
        log.Fatal("查询余额失败:", err)
    }

    fmt.Printf("Jetton 钱包地址: %s\n", userAddr.String())
    fmt.Printf("Jetton 余额: %s\n", balance.String())
}
```

### Jetton 转账

```go
// 向目标地址转账 Jetton 代币
func TransferJetton(
    ctx context.Context,
    w *wallet.Wallet,
    jettonMaster *address.Address,
    toAddr *address.Address,
    amount *big.Int,
) error {
    // Jetton Transfer 消息体
    // 格式: transfer#0f8a7ea5 query_id:uint64 amount:(VarUInteger 16) 
    //        destination:MsgAddress response_destination:MsgAddress 
    //        custom_payload:(Maybe ^Cell) forward_ton_amount:(VarUInteger 16) 
    //        forward_payload:(Either Cell ^Cell)

    transferBody, err := jetton.BuildTransferBody(
        toAddr,           // 接收方地址
        amount,           // 转账数量
        big.NewInt(1),    // forward_ton_amount（转发 TON 数量，通常为 1 nanoTON）
        nil,              // forward_payload（可选）
        nil,              // response_destination（可选）
        nil,              // custom_payload（可选）
    )
    if err != nil {
        return fmt.Errorf("构建转账消息体失败: %w", err)
    }

    // 获取用户的 Jetton 钱包地址
    jettonWalletAddr, err := jetton.GetJettonWalletAddress(
        ctx, api, jettonMaster, w.WalletAddress(),
    )
    if err != nil {
        return fmt.Errorf("获取 Jetton 钱包地址失败: %w", err)
    }

    // 发送转账消息
    err = w.Send(ctx, &wallet.Message{
        Mode: wallet.MsgWithRemainingValue,
        InternalMessage: &wallet.InternalMessage{
            IHRDisabled: false,
            Bounce:      true,
            SrcAddr:     w.WalletAddress(),
            DstAddr:     jettonWalletAddr,
            Amount:      ton.MustToNano("0.05"), // Gas 费（TON）
            Body:        transferBody,
        },
    }, true)

    return err
}
```

## 7. TON 服务层封装

### services/ton_service.go

```go
package services

import (
    "context"
    "fmt"
    "log"
    "math/big"
    "strings"

    "github.com/xssnick/tonutils-go/address"
    "github.com/xssnick/tonutils-go/liteclient"
    "github.com/xssnick/tonutils-go/ton"
    "github.com/xssnick/tonutils-go/ton/wallet"
)

// TONService TON 区块链服务
type TONService struct {
    api    *ton.APIClient
    client *liteclient.ConnectionPool
}

// NewTONService 创建 TON 服务
func NewTONService(configURL string) (*TONService, error) {
    client := liteclient.NewConnectionPool()
    err := client.AddConnectionsFromConfigUrl(context.Background(), configURL)
    if err != nil {
        return nil, fmt.Errorf("连接 TON 失败: %w", err)
    }

    return &TONService{
        api:    ton.NewAPIClient(client),
        client: client,
    }, nil
}

// NewTONServiceHTTP 使用 HTTP API 创建服务
func NewTONServiceHTTP(apiURL, apiKey string) *TONService {
    return &TONService{
        api: ton.NewAPIClient(ton.NewHTTPClient(apiURL, apiKey)),
    }
}

// GetBalance 查询 TON 余额
func (s *TONService) GetBalance(addrStr string) (float64, error) {
    addr, err := address.ParseAddr(addrStr)
    if err != nil {
        return 0, fmt.Errorf("地址格式错误: %w", err)
    }

    account, err := s.api.GetAccount(context.Background(), addr)
    if err != nil {
        return 0, fmt.Errorf("查询账户失败: %w", err)
    }

    return float64(account.State.Balance.Nano().Uint64()) / 1e9, nil
}

// GetAccountStatus 获取账户状态
func (s *TONService) GetAccountStatus(addrStr string) (string, error) {
    addr, err := address.ParseAddr(addrStr)
    if err != nil {
        return "", fmt.Errorf("地址格式错误: %w", err)
    }

    account, err := s.api.GetAccount(context.Background(), addr)
    if err != nil {
        return "", fmt.Errorf("查询账户失败: %w", err)
    }

    return account.State.Status, nil
}

// SendTON 发送 TON 转账
func (s *TONService) SendTON(mnemonic string, toAddrStr string, amountTON string, comment string) (string, error) {
    words := strings.Split(mnemonic, " ")
    w, err := wallet.FromSeed(s.api, words, wallet.V4R2)
    if err != nil {
        return "", fmt.Errorf("恢复钱包失败: %w", err)
    }

    toAddr, err := address.ParseAddr(toAddrStr)
    if err != nil {
        return "", fmt.Errorf("接收地址格式错误: %w", err)
    }

    amount := ton.MustToNano(amountTON)

    err = w.Send(context.Background(), &wallet.Message{
        Mode: wallet.MsgWithRemainingValue,
        InternalMessage: &wallet.InternalMessage{
            IHRDisabled: false,
            Bounce:      false,
            SrcAddr:     w.WalletAddress(),
            DstAddr:     toAddr,
            Amount:      amount,
            Body:        wallet.Comment(comment),
        },
    }, true)

    if err != nil {
        return "", fmt.Errorf("转账失败: %w", err)
    }

    return fmt.Sprintf("成功发送 %s TON 到 %s", amountTON, toAddrStr), nil
}

// GetWalletAddress 从助记词获取钱包地址
func (s *TONService) GetWalletAddress(mnemonic string) (string, error) {
    words := strings.Split(mnemonic, " ")
    w, err := wallet.FromSeed(s.api, words, wallet.V4R2)
    if err != nil {
        return "", fmt.Errorf("恢复钱包失败: %w", err)
    }

    return w.Address().String(), nil
}

// GetMasterchainInfo 获取主链信息
func (s *TONService) GetMasterchainInfo() (*ton.BlockIDExt, error) {
    info, err := s.api.GetMasterchainInfo(context.Background())
    if err != nil {
        return nil, fmt.Errorf("获取主链信息失败: %w", err)
    }
    return info.Last, nil
}

// Close 关闭连接
func (s *TONService) Close() {
    if s.client != nil {
        s.client.Stop()
    }
}
```

## 8. TON + Gin API 示例

```go
package main

import (
    "net/http"

    "your-project/services"

    "github.com/gin-gonic/gin"
)

var tonService *services.TONService

func main() {
    // 初始化 TON 服务
    var err error
    tonService, err = services.NewTONService("https://ton.org/global.config.json")
    if err != nil {
        panic("TON 服务初始化失败: " + err.Error())
    }
    defer tonService.Close()

    r := gin.Default()

    // TON API 路由组
    ton := r.Group("/api/v1/ton")
    {
        ton.GET("/balance/:address", getTONBalance)
        ton.GET("/account/:address", getTONAccount)
        ton.GET("/masterchain/info", getMasterchainInfo)
        ton.POST("/transfer", sendTONTransfer)
    }

    r.Run(":8080")
}

// GET /api/v1/ton/balance/:address
func getTONBalance(c *gin.Context) {
    address := c.Param("address")

    balance, err := tonService.GetBalance(address)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "address": address,
        "balance": balance,
        "unit":    "TON",
    })
}

// GET /api/v1/ton/account/:address
func getTONAccount(c *gin.Context) {
    address := c.Param("address")

    balance, err := tonService.GetBalance(address)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    status, err := tonService.GetAccountStatus(address)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "address": address,
        "balance": balance,
        "status":  status,
    })
}

// GET /api/v1/ton/masterchain/info
func getMasterchainInfo(c *gin.Context) {
    info, err := tonService.GetMasterchainInfo()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "seqno":     info.Seqno,
        "root_hash": fmt.Sprintf("%x", info.RootHash),
        "file_hash": fmt.Sprintf("%x", info.FileHash),
    })
}

// POST /api/v1/ton/transfer
type TransferRequest struct {
    Mnemonic  string  `json:"mnemonic" binding:"required"`
    ToAddress string  `json:"to_address" binding:"required"`
    Amount    float64 `json:"amount" binding:"required"`
    Comment   string  `json:"comment"`
}

func sendTONTransfer(c *gin.Context) {
    var req TransferRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    result, err := tonService.SendTON(
        req.Mnemonic,
        req.ToAddress,
        fmt.Sprintf("%.9f", req.Amount),
        req.Comment,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": result,
    })
}
```

## 9. 以太坊 vs TON 对比总结

| 场景 | 以太坊实现 | TON 实现 |
|------|-----------|----------|
| 查询余额 | `client.BalanceAt(ctx, addr, nil)` | `api.GetAccount(ctx, addr)` → `State.Balance` |
| 发送交易 | `client.SendTransaction(ctx, signedTx)` | `wallet.Send(ctx, msg, wait)` |
| 地址格式 | `0x + 40hex` | Base64 编码（含 crc16 校验） |
| 代币标准 | ERC-20（合约调用 `balanceOf`） | Jetton（查询 Jetton Wallet） |
| 合约调用 | `client.CallContract()`（同步） | 发送 Internal Message（异步） |
| Gas 模型 | 动态竞价（EIP-1559） | 固定费率 |

## 📝 作业

### 作业1：TON 钱包信息查询

创建 `homework/day16/ton-wallet`：

```go
// TODO: 实现 TON 钱包信息查询
// 1. 输入助记词，显示钱包地址
// 2. 查询并显示 TON 余额
// 3. 格式化输出（带单位）
```

### 作业2：TON 转账 CLI 工具

```go
// TODO: 实现命令行转账工具
// 1. 支持 --from（助记词）、--to（地址）、--amount（数量）、--comment（备注）
// 2. 转账前显示确认信息
// 3. 转账后显示交易结果
```

### 作业3：Jetton 代币查询 API

```go
// TODO: 实现 Jetton 代币查询 API
// 1. GET /api/v1/ton/jetton/balance/:address?master=xxx
// 2. 查询指定 Jetton Master 的代币余额
// 3. 支持常见的 Jetton（USDT、NOT 等）
```

## 🎯 检查点

- ✅ 理解 TON 区块链核心概念
- ✅ 能够连接 TON 节点
- ✅ 掌握钱包创建和助记词管理
- ✅ 能够发送 TON 转账
- ✅ 了解 Jetton 代币操作
- ✅ 封装 TON 服务层

## 💡 TON 开发注意事项

1. **地址格式**：TON 地址有 Bounceable 和 Non-Bounceable 两种，转账时通常使用 Bounceable
2. **异步消息**：智能合约调用是异步的，不像以太坊那样同步返回结果
3. **Jetton 架构**：每个用户对每个 Jetton 都有一个独立的 Jetton Wallet 合约
4. **Gas 费**：TON 的 Gas 费是固定的，不像以太坊那样波动
5. **分片**：交易可能跨分片，确认时间会增加

## ⏭️ 下一课

[第17课：交易构建、签名、发送](./day17-transactions.md)
