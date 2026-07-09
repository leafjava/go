# Part8 — TON 钱包 API 服务整体解析

## 1. 项目概述

Part8 是一个基于 **Gin + tonutils-go** 的 TON 区块链钱包 HTTP API 服务。它封装了 TON 链上核心操作（查询余额、账户状态、主链信息、转账），通过 RESTful API 暴露给前端或其他服务调用。

- **模块名**: `part8`
- **Go 版本**: `1.26.3`
- **监听端口**: `8083`
- **核心依赖**: `gin-gonic/gin`（Web 框架）、`xssnick/tonutils-go v1.17.2`（TON SDK）

---

## 2. 目录结构

```
part8/
├── go.mod                         # Go module 声明与依赖管理
├── go.sum                         # 依赖版本校验文件（自动生成）
├── main.go                        # 入口文件：Gin 路由注册 + HTTP Handler
└── services/
    └── ton_service.go             # 服务层：TON 区块链交互逻辑封装
```

项目采用**两层架构**：

```
┌──────────────────────────────┐
│   main.go (HTTP 层)          │  ← Gin 路由、请求校验、响应格式
│   GET/POST Handler           │
└──────────┬───────────────────┘
           │ 调用
┌──────────▼───────────────────┐
│   services/ton_service.go    │  ← 业务逻辑层
│   TONService 结构体          │  TON 链交互的核心实现
└──────────┬───────────────────┘
           │ 调用
┌──────────▼───────────────────┐
│   tonutils-go SDK            │  ← 第三方库
│   (地址解析/API/钱包/交易)     │
└──────────────────────────────┘
```

---

## 3. 启动流程

```
main()
  │
  ├─ 1. services.NewTONService("https://ton.org/global.config.json")
  │      │
  │      ├─ liteclient.NewConnectionPool()          // 创建 Lite 连接池
  │      ├─ AddConnectionsFromConfigUrl(url)         // 从配置 URL 加载节点并建立连接
  │      └─ ton.NewAPIClient(client)                 // 创建 TON API 客户端
  │
  ├─ 2. defer tonService.Close()                     // 确保退出时释放连接
  │
  ├─ 3. gin.Default()                                // 创建 Gin 引擎（含 Logger + Recovery 中间件）
  │
  ├─ 4. 注册路由组 /api/v1/ton
  │      ├─ GET  /balance/:address    → getTONBalance
  │      ├─ GET  /account/:address    → getTONAccount
  │      ├─ GET  /masterchain/info    → getMasterchainInfo
  │      └─ POST /transfer            → sendTONTransfer
  │
  └─ 5. r.Run(":8083")                               // 启动 HTTP 服务，监听 8083 端口
```

---

## 4. main.go 逐段解析

### 4.1 全局变量

```go
var tonService *services.TONService
```

服务实例在 `main()` 中初始化，作为**包级全局变量**供所有 handler 共享。这种方式适合单实例场景，避免了每个 handler 都传参。

### 4.2 路由注册

```go
ton := r.Group("/api/v1/ton")
{
    ton.GET("/balance/:address", getTONBalance)   // :address 是路径参数
    ton.GET("/account/:address", getTONAccount)
    ton.GET("/masterchain/info", getMasterchainInfo)
    ton.POST("/transfer", sendTONTransfer)
}
```

使用 Gin 的**路由组**将所有 TON API 归到 `/api/v1/ton` 前缀下，便于统一管理和后续加中间件（如鉴权）。

### 4.3 查询余额接口

```
GET /api/v1/ton/balance/:address
```

**请求示例**:
```
GET /api/v1/ton/balance/EQA...TON地址...
```

**处理流程**:
```
c.Param("address")                     → 步骤1: 提取路径参数
    ↓
tonService.GetBalance(address)         → 步骤2: 调用服务层查询链上余额
    ↓
返回 JSON {address, balance, unit}     → 步骤3: 统一 JSON 响应
```

**响应示例**:
```json
{
    "address": "EQA...",
    "balance": 1.5,
    "unit": "TON"
}
```

**错误处理**: 统一返回 `{"error": "错误描述"}` + HTTP 500。

### 4.4 查询账户信息接口

```
GET /api/v1/ton/account/:address
```

这个接口**同时查询余额和状态**，比单独查余额提供更多信息。

```
c.Param("address")
    ├─ tonService.GetBalance(address)       // 查询余额
    └─ tonService.GetAccountStatus(address) // 查询状态 (active/uninit/frozen)
```

**账户状态含义**:

| 状态 | 含义 |
|------|------|
| `ACTIVE` | 正常激活状态 |
| `UNINIT` | 未初始化（地址存在但无合约代码） |
| `FROZEN` | 已冻结 |
| `NON_EXIST` | 不存在 |

### 4.5 查询主链信息接口

```
GET /api/v1/ton/masterchain/info
```

返回 TON 主链（Masterchain）最新区块的关键信息：

| 字段 | 含义 |
|------|------|
| `seqno` | 区块序号（高度），递增 |
| `root_hash` | 区块的 Merkle 根哈希 |
| `file_hash` | 区块文件的哈希 |

**用途**: 监控链上最新状态、确认同步进度。

### 4.6 转账接口 — TransferRequest

```go
type TransferRequest struct {
    Mnemonic  string  `json:"mnemonic" binding:"required"`
    ToAddress string  `json:"to_address" binding:"required"`
    Amount    float64 `json:"amount" binding:"required"`
    Comment   string  `json:"comment"`
}
```

| 字段 | JSON 名 | 必填 | 说明 |
|------|---------|------|------|
| `Mnemonic` | `mnemonic` | ✅ | 发送方的 24 个助记词（空格分隔），用于恢复私钥签名 |
| `ToAddress` | `to_address` | ✅ | 接收方 TON 地址 |
| `Amount` | `amount` | ✅ | 转账金额（TON 单位，如 `1.5`） |
| `Comment` | `comment` | ❌ | 转账附言，写入链上消息体 |

**处理流程**:
```
POST /api/v1/ton/transfer
  ↓
c.ShouldBindJSON(&req)                          → 步骤1: Gin 自动校验 JSON 并反序列化
  │  (binding:"required" 的字段缺失会直接返回 400)
  ↓
fmt.Sprintf("%.9f", req.Amount)                 → 步骤2: float64 → 9位小数精度字符串
  │  （避免浮点精度丢失，如 0.1 + 0.2 问题）
  ↓
tonService.SendTON(mnemonic, to, amount, comment) → 步骤3: 执行链上转账
  ↓
返回 {message: "成功发送 X TON 到 Y"}             → 步骤4: 返回结果
```

> **安全注意事项**: 将助记词放在请求体中传输存在安全风险。生产环境应考虑：
> - 使用 HTTPS
> - 私钥/助记词不应由客户端传入，而应由服务端安全存储
> - 或使用签名机制替代明文助记词

---

## 5. services/ton_service.go 逐方法解析

### 5.1 TONService 结构体

```go
type TONService struct {
    api    *ton.APIClient             // 链上 API 客户端
    client *liteclient.ConnectionPool // Lite 网络连接池
}
```

两个字段的分工：
- **`api`**: 所有链上查询和操作的入口（查余额、查状态、发交易等）
- **`client`**: 管理底层 TCP/TLS 长连接到 TON Lite 节点，`Close()` 时需要释放

### 5.2 NewTONService — 服务初始化

```
NewTONService(configURL)
  ├─ liteclient.NewConnectionPool()
  │   └─ 创建连接池，管理多个 Lite 节点
  ├─ AddConnectionsFromConfigUrl(url)
  │   └─ 下载配置 JSON → 解析节点 IP/端口/公钥 → 建立连接
  └─ ton.NewAPIClient(client)
      └─ 封装连接池为 API 客户端
```

TON 网络使用 **Lite Client** 协议（轻客户端），通过连接 Lite 服务器获取区块数据，不需要同步全节点。

### 5.3 GetBalance — 查询余额

```
字符串地址
  ├─ address.ParseAddr(addrStr)
  │   └─ 解析用户友好的地址字符串（如 EQA...）为内部地址结构体
  ├─ GetMasterchainInfo()
  │   └─ 获取主链最新区块（v1.17.2 的 GetAccount 必须指定区块）
  ├─ GetAccount(ctx, block, addr)
  │   └─ 在指定区块高度查询账户信息（余额、状态、代码等）
  └─ Balance.Nano().Uint64() / 1e9
      └─ nanoTON → TON 转换（1 TON = 10^9 nanoTON）
```

**调用链深度**:
```
TONService.GetBalance(addrStr)
  → ton.APIClient.GetMasterchainInfo(ctx)
  → ton.APIClient.GetAccount(ctx, block, addr)
    → client.QueryLiteserver(ctx, GetAccountState{...})
      → TCP/TLS 发送到 Lite 节点
        → Lite 节点从本地区块数据查询返回
```

### 5.4 GetAccountStatus — 查询账户状态

与 `GetBalance` 流程基本相同，区别在于返回 `account.State.Status`（`tlb.AccountStatus` 类型）。

**关键细节**: `AccountStatus` 是 `type AccountStatus string` 的自定义类型，必须用 `string()` 显式转换才能返回 `string`：

```go
return string(account.State.Status), nil
```

### 5.5 SendTON — 转账（核心方法）

这是整个服务最复杂的操作，流程如下：

```
助记词字符串
  ├─ strings.Split(" ")
  │   └─ 拆分 24 个单词 → []string
  ├─ wallet.FromSeed(api, words, wallet.V4R2)
  │   └─ 通过 PBKDF2 从助记词派生种子 → 生成 ed25519 私钥/公钥 → 计算钱包地址
  ├─ address.ParseAddr(toAddrStr)
  │   └─ 解析接收方地址
  ├─ tlb.FromTON(amountTON)
  │   └─ "1.5" → Coins{val: 1500000000, decimals: 9}
  ├─ w.BuildTransfer(toAddr, amount, bounce, comment)
  │   ├─ wallet.CreateCommentCell(comment)      // 将文本编码为 TON Cell
  │   └─ 构建 Message{Mode, InternalMessage}
  │       ├─ Mode: PayGasSeparately(1) + IgnoreErrors(2) = 3
  │       │   含义: 用合约余额单独付 Gas，出错不中断整体执行
  │       └─ InternalMessage: {IHRDisabled, Bounce, DstAddr, Amount, Body}
  └─ w.Send(ctx, msg, true)
      ├─ BuildMessage(ctx, messages)     // 构建外部消息
      │   └─ 组装 StateInit + 签名 + 消息体 → ExternalMessage
      ├─ SendExternalMessageWaitTransaction(...) // 发送并等待确认
      └─ 返回成功消息
```

**V4R2 钱包版本**: TON 目前最常用的钱包合约，支持插件扩展和订阅功能。

**消息模式说明**: `PayGasSeparately(1) + IgnoreErrors(2) = 3`：
- `PayGasSeparately(1)`: Gas 费由合约余额支付
- `IgnoreErrors(2)`: 出错时不中断，允许批量交易中部分失败

### 5.6 GetWalletAddress — 从助记词恢复地址

辅助方法，从助记词派生出钱包地址，适用于"输入助记词查看地址"的场景。

```
助记词 → wallet.FromSeed → w.Address().String()
```

### 5.7 GetMasterchainInfo — 查询主链信息

```
GetMasterchainInfo()
  → Client.QueryLiteserver(ctx, GetMasterchainInf{})
    → 返回 *BlockIDExt{Workchain, Shard, SeqNo, RootHash, FileHash}
```

- **直接返回** `*BlockIDExt`（不是包装结构体），不要访问 `.Last`
- 返回的区块信息可用于判断链上最新状态

### 5.8 Close — 释放资源

```go
if s.client != nil {
    s.client.Stop()
}
```

一定要在程序退出前调用，否则可能有连接泄漏。

---

## 6. API 接口文档汇总

| 方法 | 路径 | 说明 | 请求参数 | 响应示例 |
|------|------|------|----------|----------|
| GET | `/api/v1/ton/balance/:address` | 查询余额 | `:address` 路径参数 | `{address, balance: 1.5, unit: "TON"}` |
| GET | `/api/v1/ton/account/:address` | 查询账户信息 | `:address` 路径参数 | `{address, balance, status}` |
| GET | `/api/v1/ton/masterchain/info` | 主链信息 | 无 | `{seqno, root_hash, file_hash}` |
| POST | `/api/v1/ton/transfer` | 转账 | JSON Body: `{mnemonic, to_address, amount, comment?}` | `{message: "成功发送 X TON 到 Y"}` |

**统一错误响应**: `{"error": "错误描述"}` + 对应 HTTP 状态码（400/500）。

---

## 7. 依赖关系图

```
part8 (module: part8)
│
├── github.com/gin-gonic/gin v1.12.0
│   └── HTTP 框架：路由、中间件、JSON 绑定/校验
│
├── github.com/xssnick/tonutils-go v1.17.2
│   ├── ton          — APIClient（链上交互入口）
│   ├── ton/wallet   — 钱包恢复、转账、消息构建
│   ├── liteclient   — Lite 客户端连接池管理
│   ├── tlb          — TON 数据结构（Coins, Account, InternalMessage）
│   └── address      — 地址解析与格式化
│
└── 间接依赖
    ├── filippo.io/edwards25519   — ed25519 椭圆曲线（钱包签名）
    ├── golang.org/x/crypto       — pbkdf2（助记词派生密钥）
    └── github.com/pierrec/lz4    — LZ4 压缩（区块数据传输）
```

---

## 8. 完整数据流图（以转账为例）

```
客户端 POST /api/v1/ton/transfer
  │  {"mnemonic":"word1 word2 ...", "to_address":"EQA...", "amount":0.5, "comment":"备注"}
  ▼
┌─ main.go ──────────────────────────────────────────┐
│ sendTONTransfer(c)                                  │
│   ├─ c.ShouldBindJSON(&req)                         │
│   │   └─ Gin 校验 JSON → TransferRequest 结构体      │
│   ├─ fmt.Sprintf("%.9f", req.Amount) → "0.500000000"│
│   └─ tonService.SendTON(mnemonic, to, amount, note) │
└──────────────────┬──────────────────────────────────┘
                   ▼
┌─ services/ton_service.go ───────────────────────────┐
│ SendTON()                                           │
│   ├─ strings.Split → 助记词单词数组                    │
│   ├─ wallet.FromSeed → 恢复钱包                       │
│   ├─ address.ParseAddr → 解析目标地址                   │
│   ├─ tlb.FromTON → 金额转换                           │
│   ├─ w.BuildTransfer → 构建 InternalMessage           │
│   └─ w.Send(..., true) → 签名 + 广播 + 等待确认        │
└──────────────────┬──────────────────────────────────┘
                   ▼
┌─ tonutils-go SDK ───────────────────────────────────┐
│   ├─ CreateCommentCell → 附言编码为 TON Cell           │
│   ├─ BuildMessage → 组装 ExternalMessage + 签名       │
│   ├─ SendExternalMessageWaitTransaction              │
│   │   ├─ QueryLiteserver → TCP/TLS 发送              │
│   │   └─ WaitForBlock → 轮询等待区块确认               │
└──────────────────┬──────────────────────────────────┘
                   ▼
              TON 区块链网络
         (Lite Server → Validator)
```

---

## 9. 关键技术点总结

| 技术点 | 说明 |
|--------|------|
| **Lite Client 协议** | 轻客户端协议，通过连接 Lite 服务器获取区块数据，无需同步全节点 |
| **V4R2 钱包** | TON 最常用的钱包合约版本，支持插件和订阅 |
| **ed25519 签名** | TON 使用 ed25519 椭圆曲线进行交易签名 |
| **nanoTON** | TON 最小单位，1 TON = 10⁹ nanoTON，链上余额均以 nanoTON 存储 |
| **PBKDF2 密钥派生** | 助记词通过 PBKDF2 + HMAC-SHA512 派生种子 → 私钥/公钥/地址 |
| **消息模式** | Mode=3: PayGasSeparately(1) + IgnoreErrors(2)，合约余额付 Gas，出错不中断 |
| **Bounce 标志** | `false`: 目标地址不存在时资金不退回（普通转账）；`true`: 退回（合约调用推荐） |
| **IHRDisabled** | `true`: 禁用即时超立方路由，消息走标准路由 |
| **Gin binding** | `binding:"required"` 自动校验必填字段，减少手写校验代码 |
