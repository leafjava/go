# 第15课：Go 调用以太坊合约 — 学习笔记

> 学习时间：3-4小时 | 难度：⭐⭐⭐⭐

## 目录结构

```
day15-ethereum/
├── README.md          ← 当前文件
├── part2/             ← 连接以太坊节点 + 获取最新区块
│   ├── main.go
│   ├── go.mod
│   └── go.sum
└── part3/             ← 查询 ETH 余额
    ├── main.go
    ├── go.mod
    └── go.sum
```

## 核心概念速查

### go-ethereum 常用 API

| 方法 | 用途 | 返回值 |
|------|------|--------|
| `ethclient.Dial(url)` | 连接节点（支持 http/https/wss） | `*ethclient.Client` |
| `client.BlockNumber(ctx)` | 最新区块号 | `uint64` |
| `client.BalanceAt(ctx, addr, nil)` | 查询余额（`nil` = 最新区块） | `*big.Int`（单位 wei） |
| `client.PendingNonceAt(ctx, addr)` | 获取 nonce | `uint64` |
| `client.SuggestGasPrice(ctx)` | 建议 Gas 价格 | `*big.Int` |
| `client.NetworkID(ctx)` | 链 ID（1=主网） | `*big.Int` |
| `client.SendTransaction(ctx, tx)` | 发送已签名交易 | `error` |
| `client.CallContract(ctx, msg, nil)` | 调用合约（只读） | `[]byte` |
| `client.SubscribeFilterLogs(ctx, query, ch)` | 订阅事件日志 | `ethereum.Subscription` |

### 关键类型

```go
import (
    "github.com/ethereum/go-ethereum/common"      // 地址类型
    "github.com/ethereum/go-ethereum/ethclient"    // 客户端
    "github.com/ethereum/go-ethereum/crypto"       // 私钥/签名
    "github.com/ethereum/go-ethereum/core/types"   // 交易类型
)
```

### 以太坊单位换算

| 单位 | 值 |
|------|-----|
| 1 ETH | 10¹⁸ wei |
| 1 Gwei | 10⁹ wei |
| 1 wei | 最小单位 |

```go
// wei → ETH
ethValue := new(big.Float).Quo(
    new(big.Float).SetInt(balance),
    big.NewFloat(1e18),
)
```

## 测试方法

### part2 — 连接节点 + 区块号

```bash
cd D:\web3project\go\day15-ethereum\part2
go mod tidy
go run main.go
```

### part3 — 查询余额

```bash
cd D:\web3project\go\day15-ethereum\part3
go mod tidy
go run main.go
```

## 踩坑记录

### 1. `go.sum` 缺失传递依赖

```
missing go.sum entry for module providing package ...
```

**解决：** `go mod tidy`

### 2. RPC 节点不可用

`eth.llamarpc.com` 返回 Cloudflare 521 错误（源站挂了）。

**可用替代 RPC：**

| RPC URL | 备注 |
|---------|------|
| `https://ethereum-rpc.publicnode.com` | ✅ 推荐，稳定 |
| `https://rpc.ankr.com/eth` | 稳定 |
| `https://cloudflare-eth.com` | Cloudflare 提供 |
| `https://eth.drpc.org` | 免费公共节点 |
| `https://1rpc.io/eth` | 免费 |

### 3. `big.Int` 和 `big.Float` 混用

`BalanceAt` 返回 `*big.Int`（wei），如需显示 ETH 需手动转换，不能直接除。

### 4. 地址格式

必须用 `common.HexToAddress("0x...")` 转换，不能直接传字符串。

## 资源链接

- go-ethereum 官方文档：https://geth.ethereum.org/docs
- go-ethereum 源码：https://github.com/ethereum/go-ethereum
- Etherscan 浏览器：https://etherscan.io
- 公共 RPC 列表：https://chainlist.org/chain/1
