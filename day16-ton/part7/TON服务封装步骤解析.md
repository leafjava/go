# TON 服务封装 — 步骤解析

本文档梳理了 `TONService` 的设计，将 TON 区块链操作（连接、查询、转账）封装为可复用的服务结构体。

---

## 整体架构

```
┌──────────────────────────────────────────────────────────────────┐
│                         TONService                               │
│                                                                  │
│  ┌────────────────┐  ┌────────────────┐  ┌──────────────────┐   │
│  │ NewTONService  │  │ GetBalance     │  │ SendTON          │   │
│  │ (liteclient)   │  │ 查询 TON 余额  │  │ 发送 TON 转账    │   │
│  └────────────────┘  └────────────────┘  └──────────────────┘   │
│                                                                  │
│  ┌─────────────────────┐  ┌──────────────────────────────────┐   │
│  │ GetWalletAddress    │  │ GetMasterchainInfo               │   │
│  │ 助记词 → 地址       │  │ 主链最新区块                      │   │
│  └─────────────────────┘  └──────────────────────────────────┘   │
│                                                                  │
│  ┌─────────────────────┐  ┌──────────────────────────────────┐   │
│  │ GetAccountStatus    │  │ Close                            │   │
│  │ 账户状态            │  │ 释放连接                         │   │
│  └─────────────────────┘  └──────────────────────────────────┘   │
│                                                                  │
│          api (*ton.APIClient)          client (*ConnectionPool)   │
└──────────────────────────────────────────────────────────────────┘
                           │
                           ▼
                   TON 主网节点 (ADNL/LiteClient)
```

---

## 步骤 1：定义服务结构体

```go
type TONService struct {
    api    *ton.APIClient
    client *liteclient.ConnectionPool
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| **api** | `*ton.APIClient` | TON API 客户端，封装所有链上交互方法 |
| **client** | `*liteclient.ConnectionPool` | LiteClient 连接池，管理与 TON 节点的 ADNL 连接 |

> 两个字段都是小写（私有），外部只能通过方法访问，遵循封装原则。

---

## 步骤 2：构造函数

### 2.1 LiteClient 方式（推荐）

```go
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
```

| 步骤 | 说明 |
|------|------|
| ① `NewConnectionPool` | 创建 LiteClient 连接池 |
| ② `AddConnectionsFromConfigUrl` | 从 `ton.org/global.config.json` 下载节点列表并建立 ADNL 连接 |
| ③ `ton.NewAPIClient(client)` | 基于连接池创建 API 客户端 |
| ④ 返回 | `*TONService` 指针，调用方用 `defer svc.Close()` 释放 |

### 2.2 HTTP API 方式

```go
func NewTONServiceHTTP(apiURL, apiKey string) *TONService {
    return &TONService{
        api: ton.NewAPIClient(ton.NewHTTPClient(apiURL, apiKey)),
    }
}
```

| 参数 | 说明 |
|------|------|
| **apiURL** | HTTP API 地址（如 TonCenter `https://toncenter.com/api/v2`） |
| **apiKey** | API 密钥（可为空字符串） |
| **注意** | HTTP 方式 `client` 为 nil，`Close()` 是无操作 |

> ⚠️ LiteClient 方式直接连 TON 节点，去中心化；HTTP 方式依赖第三方中心化 API。

---

## 步骤 3：TON 余额查询

```go
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
```

| 步骤 | 代码 | 说明 |
|------|------|------|
| ① 解析地址 | `address.ParseAddr(addrStr)` | 字符串 → `*address.Address` |
| ② 查询账户 | `s.api.GetAccount(ctx, addr)` | 从 TON 节点获取账户状态 |
| ③ 提取余额 | `account.State.Balance.Nano()` | 返回 `*big.Int`（单位：nanoTON） |
| ④ 转换单位 | `/ 1e9` | nanoTON → TON（1 TON = 10⁹ nanoTON） |

> `account.State` 为 nil 时表示账户未激活，余额为 0。

---

## 步骤 4：获取账户状态

```go
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
```

| 状态值 | 含义 |
|--------|------|
| `uninitialized` | 账户不存在（从未收到过币） |
| `active` | 账户已激活（有合约代码和数据） |
| `frozen` | 账户被冻结 |

---

## 步骤 5：发送 TON 转账

```go
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
```

### 5.1 助记词 → 钱包

```go
words := strings.Split(mnemonic, " ")
w, err := wallet.FromSeed(s.api, words, wallet.V4R2)
```

| 项目 | 说明 |
|------|------|
| **mnemonic** | 24 个助记词，空格分隔 |
| **V4R2** | 钱包版本 W4R2（当前主流版本） |
| **FromSeed** | 从助记词推导私钥、公钥、地址，返回可签名的 `*wallet.Wallet` |

### 5.2 构造消息

```go
&wallet.Message{
    Mode: wallet.MsgWithRemainingValue,
    InternalMessage: &wallet.InternalMessage{...},
}
```

| 字段 | 说明 |
|------|------|
| **Mode** | 消息模式，控制余额如何分配 |
| **IHRDisabled** | 禁用即时超路由（Instant Hypercube Routing） |
| **Bounce** | `false` → 普通地址不收 bounce；`true` → 合约地址需要 |
| **SrcAddr** | 发送方地址 |
| **DstAddr** | 接收方地址 |
| **Amount** | 转账金额（nanoTON） |
| **Body** | 消息体（可附带备注 comment） |

### 5.3 send 流程

```
助记词 → FromSeed → Wallet → Send → 签名 → 广播 → TON 网络
```

| 步骤 | 内部操作 |
|------|----------|
| 签名 | 使用钱包私钥对消息签名 |
| nonce | 自动获取 seqno（防重放） |
| 广播 | 通过 LiteClient 发送到 TON 节点 |
| 等待 | `true` → 等待交易确认后再返回 |

---

## 步骤 6：从助记词获取钱包地址

```go
func (s *TONService) GetWalletAddress(mnemonic string) (string, error) {
    words := strings.Split(mnemonic, " ")
    w, err := wallet.FromSeed(s.api, words, wallet.V4R2)
    if err != nil {
        return "", fmt.Errorf("恢复钱包失败: %w", err)
    }

    return w.Address().String(), nil
}
```

| 用途 | 导入助记词后获取对应地址 |
|------|------|
| **输入** | 24 个助记词 |
| **输出** | Bounceable 地址（`EQ...` 格式） |

---

## 步骤 7：获取主链信息

```go
func (s *TONService) GetMasterchainInfo() (*ton.BlockIDExt, error) {
    info, err := s.api.GetMasterchainInfo(context.Background())
    if err != nil {
        return nil, fmt.Errorf("获取主链信息失败: %w", err)
    }
    return info.Last, nil
}
```

| 返回值 | 说明 |
|--------|------|
| `info.Last.SeqNo` | 最新主链区块序号 |
| `info.Last.RootHash` | 最新主链区块根哈希 |
| `info.Last.FileHash` | 最新主链区块文件哈希 |

---

## 步骤 8：释放资源

```go
func (s *TONService) Close() {
    if s.client != nil {
        s.client.Stop()
    }
}
```

| 项目 | 说明 |
|------|------|
| **nil 判断** | HTTP 方式 `client` 为 nil，跳过 |
| **Stop()** | 关闭所有 LiteClient ADNL 连接 |

---

## 完整代码

```go
package services

import (
    "context"
    "fmt"
    "strings"

    "github.com/xssnick/tonutils-go/address"
    "github.com/xssnick/tonutils-go/liteclient"
    "github.com/xssnick/tonutils-go/ton"
    "github.com/xssnick/tonutils-go/ton/wallet"
)

type TONService struct {
    api    *ton.APIClient
    client *liteclient.ConnectionPool
}

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

func (s *TONService) SendTON(mnemonic, toAddrStr, amountTON, comment string) (string, error) {
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

func (s *TONService) GetWalletAddress(mnemonic string) (string, error) {
    words := strings.Split(mnemonic, " ")
    w, err := wallet.FromSeed(s.api, words, wallet.V4R2)
    if err != nil {
        return "", fmt.Errorf("恢复钱包失败: %w", err)
    }
    return w.Address().String(), nil
}

func (s *TONService) GetMasterchainInfo() (*ton.BlockIDExt, error) {
    info, err := s.api.GetMasterchainInfo(context.Background())
    if err != nil {
        return nil, fmt.Errorf("获取主链信息失败: %w", err)
    }
    return info.Last, nil
}

func (s *TONService) Close() {
    if s.client != nil {
        s.client.Stop()
    }
}
```

---

## 使用示例

```go
package main

import (
    "fmt"
    "log"
    "part7/services"
)

func main() {
    svc, err := services.NewTONService("https://ton.org/global.config.json")
    if err != nil {
        log.Fatal(err)
    }
    defer svc.Close()

    // 查询余额
    balance, err := svc.GetBalance("EQA0gvftODUs9itSl5whd01IxERToRZkA_-tsriaZPxuRkS4")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("余额: %.9f TON\n", balance)

    // 从助记词获取地址
    addr, err := svc.GetWalletAddress("your 24 mnemonic words here")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("钱包地址: %s\n", addr)
}
```

---

## 关键知识点

| 概念 | 解释 |
|------|------|
| **LiteClient** | TON 轻客户端协议（ADNL），直接连接全节点，无需信任第三方 |
| **ConnectionPool** | 连接池，管理与多个 TON 节点的连接，提高可用性 |
| **V4R2** | 当前主流钱包合约版本，支持任意顺序交易 |
| **nanoTON** | TON 最小单位，1 TON = 10^9 nanoTON |
| **Bounce** | 消息退回机制：发给未激活合约的消息会被退回并扣除 Gas |
| **助记词** | BIP-39 标准，24 个单词可恢复整个钱包（所有地址和资金） |
| **seqno** | 钱包交易序号，每发一笔 +1，防止重放 |
| **Send(..., true)** | 最后参数为 true 等待交易确认；false 只广播不等待 |
