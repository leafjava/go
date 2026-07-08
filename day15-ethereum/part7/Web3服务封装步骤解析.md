# 7. Web3 服务封装 — 步骤解析

本文档梳理了将以太坊操作封装为可复用 `EthereumService` 的完整流程，包含 ETH 余额查询和转账交易。

---

## 整体架构

```
┌────────────────────────────────────────────────────────┐
│                   EthereumService                      │
│                                                        │
│  ┌─────────────┐  ┌──────────────┐  ┌──────────────┐  │
│  │ New...()    │  │ GetBalance() │  │ SendTx()     │  │
│  │ 创建客户端   │  │ 查询ETH余额  │  │ 发送交易      │  │
│  └─────────────┘  └──────────────┘  └──────────────┘  │
│         │                │                  │          │
│         └────────────────┼──────────────────┘          │
│                          ▼                             │
│               ethclient.Client (RPC)                   │
└────────────────────────────────────────────────────────┘
                          │
                          ▼
              以太坊节点 (JSON-RPC / WebSocket)
```

---

## 步骤 1：定义服务结构体

```go
type EthereumService struct {
    client *ethclient.Client
}
```

| 项目 | 说明 |
|------|------|
| **client** | `ethclient.Client` 的指针，封装了以太坊 JSON-RPC 交互 |
| **小写字段** | 未导出（私有），外部无法直接访问，强制通过方法操作 |

---

## 步骤 2：构造函数（工厂模式）

```go
func NewEthereumService(rpcURL string) (*EthereumService, error) {
    client, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, err
    }
    return &EthereumService{client: client}, nil
}
```

| 项目 | 说明 |
|------|------|
| **输入** | `rpcURL` — 以太坊节点地址（`https://...` 或 `wss://...`） |
| **输出** | `*EthereumService` 指针 + `error` |
| **Dial** | 建立与节点的 JSON-RPC 连接 |
| **模式** | 构造函数模式，不直接暴露创建细节 |

---

## 步骤 3：ETH 余额查询

```go
func (s *EthereumService) GetBalance(address string) (*big.Int, error) {
    addr := common.HexToAddress(address)
    balance, err := s.client.BalanceAt(context.Background(), addr, nil)
    if err != nil {
        return nil, err
    }
    return balance, nil
}
```

### 3.1 地址转换

```go
addr := common.HexToAddress(address)
```

| 项目 | 说明 |
|------|------|
| **输入** | `"0x742d35Cc..."` 十六进制字符串 |
| **输出** | `common.Address`（底层是 `[20]byte`） |

### 3.2 查询余额

```go
balance, err := s.client.BalanceAt(context.Background(), addr, nil)
```

| 参数 | 说明 |
|------|------|
| **ctx** | `context.Background()`，默认上下文（可替换为带超时的 ctx） |
| **addr** | 要查询的地址 |
| **blockNumber** | `nil` = 最新区块；可传入 `big.NewInt(区块号)` 查历史余额 |

| 项目 | 说明 |
|------|------|
| **返回值** | `*big.Int`，单位 wei（1 ETH = 10^18 wei） |
| **底层 RPC** | `eth_getBalance` |

---

## 步骤 4：发送 ETH 转账交易

### 4.1 从私钥恢复账户

```go
privateKey, err := crypto.HexToECDSA(privateKeyHex)
if err != nil {
    return "", err
}

publicKey := privateKey.Public()
publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
if !ok {
    return "", fmt.Errorf("无法转换公钥")
}

fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
```

| 步骤 | 代码 | 说明 |
|------|------|------|
| ① 解析私钥 | `crypto.HexToECDSA(privateKeyHex)` | 将十六进制私钥转为 `*ecdsa.PrivateKey` |
| ② 提取公钥 | `privateKey.Public()` | 从私钥派生公钥（椭圆曲线乘法） |
| ③ 类型断言 | `publicKey.(*ecdsa.PublicKey)` | `Public()` 返回接口，需断言为具体类型 |
| ④ 推导地址 | `crypto.PubkeyToAddress(...)` | 对公钥做 Keccak-256 取后 20 字节，得以太坊地址 |

> ⚠️ **安全提醒**：私钥永远不要硬编码在代码中，应从环境变量或配置文件读取。

### 4.2 获取 nonce 和 Gas

```go
nonce, err := s.client.PendingNonceAt(context.Background(), fromAddress)

gasLimit := uint64(21000)  // ETH 标准转账固定 21000
gasPrice, err := s.client.SuggestGasPrice(context.Background())
```

| 参数 | 说明 |
|------|------|
| **nonce** | 发送地址的交易计数，防止重放攻击 |
| **gasLimit** | ETH 转账固定 21000 gas（合约调用需更多） |
| **gasPrice** | 从节点获取当前建议的 Gas 价格（单位 wei） |

> `PendingNonceAt` 包含 pending 交易，确保链上交易不积压时不冲突。

### 4.3 构造交易

```go
toAddr := common.HexToAddress(toAddress)
tx := types.NewTransaction(nonce, toAddr, amount, gasLimit, gasPrice, nil)
```

| 参数 | 说明 |
|------|------|
| **nonce** | 发送者交易序号 |
| **toAddr** | 接收方地址 |
| **amount** | 转账金额（`*big.Int`，单位 wei） |
| **gasLimit** | Gas 上限 |
| **gasPrice** | Gas 价格 |
| **最后一个 nil** | data 字段，ETH 转账为 nil；合约调用时传入 calldata |

> `types.NewTransaction` 创建的是 **Legacy 交易**（pre-EIP-1559）。如需 EIP-1559 动态费用，应使用 `types.NewTx(&types.DynamicFeeTx{...})`。

### 4.4 签名交易

```go
chainID, err := s.client.NetworkID(context.Background())

signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
```

| 项目 | 说明 |
|------|------|
| **chainID** | 链 ID（以太坊主网 = 1），从节点查询获取 |
| **EIP155Signer** | EIP-155 签名器，防止跨链重放攻击 |
| **SignTx** | 用私钥对交易签名，返回已签名交易 |

> EIP-155 于 2016 年实施，将 chainID 加入签名数据，使一条链上的签名在另一条链上无效。

### 4.5 广播交易

```go
err = s.client.SendTransaction(context.Background(), signedTx)
if err != nil {
    return "", err
}
return signedTx.Hash().Hex(), nil
```

| 项目 | 说明 |
|------|------|
| **SendTransaction** | 底层 `eth_sendRawTransaction`，将签名交易广播到全网 |
| **signedTx.Hash()** | 交易哈希（`common.Hash`） |
| **.Hex()** | 转为 `0x...` 格式字符串返回给调用方 |

> 广播成功 ≠ 交易确认。需要等待矿工打包进区块才算上链。

---

## 步骤 5：资源释放

```go
func (s *EthereumService) Close() {
    s.client.Close()
}
```

| 项目 | 说明 |
|------|------|
| **Close** | 释放 RPC 连接。使用方通常用 `defer svc.Close()` |
| **注意** | 对于 HTTP 客户端，`Close` 关闭空闲连接；WebSocket 则关闭长连接 |

---

## 完整流程：SendTransaction 全链路

```
┌─────────┐   ┌─────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐
│ 私钥    │ → │ 公钥     │ → │ 地址      │ → │ nonce    │ → │ 构造交易  │
│ Hex→ECDSA│   │ Public()│   │ Keccak256│   │ Pending  │   │ NewTx     │
└─────────┘   └─────────┘   └──────────┘   └──────────┘   └──────────┘
                                                                    │
                                                                    ▼
┌──────────┐   ┌──────────┐   ┌──────────┐   ┌──────────┐    ┌──────────┐
│ 返回哈希  │ ← │ 广播交易  │ ← │ EIP155   │ ← │ 获取     │ ← │ Gas 参数  │
│ Hex()    │   │ SendTx   │   │ 签名    │   │ chainID  │    │ price/limit│
└──────────┘   └──────────┘   └──────────┘   └──────────┘    └──────────┘
```

---

## 完整代码

```go
package services

import (
    "context"
    "crypto/ecdsa"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)

type EthereumService struct {
    client *ethclient.Client
}

func NewEthereumService(rpcURL string) (*EthereumService, error) {
    client, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, err
    }
    return &EthereumService{client: client}, nil
}

func (s *EthereumService) GetBalance(address string) (*big.Int, error) {
    addr := common.HexToAddress(address)
    balance, err := s.client.BalanceAt(context.Background(), addr, nil)
    if err != nil {
        return nil, err
    }
    return balance, nil
}

func (s *EthereumService) SendTransaction(privateKeyHex, toAddress string, amount *big.Int) (string, error) {
    // 1. 私钥 → 公钥 → 地址
    privateKey, err := crypto.HexToECDSA(privateKeyHex)
    if err != nil {
        return "", err
    }
    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        return "", fmt.Errorf("无法转换公钥")
    }
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

    // 2. nonce + gas
    nonce, err := s.client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return "", err
    }
    gasLimit := uint64(21000)
    gasPrice, err := s.client.SuggestGasPrice(context.Background())
    if err != nil {
        return "", err
    }

    // 3. 构造交易
    toAddr := common.HexToAddress(toAddress)
    tx := types.NewTransaction(nonce, toAddr, amount, gasLimit, gasPrice, nil)

    // 4. 签名
    chainID, err := s.client.NetworkID(context.Background())
    if err != nil {
        return "", err
    }
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
    if err != nil {
        return "", err
    }

    // 5. 广播
    err = s.client.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return "", err
    }
    return signedTx.Hash().Hex(), nil
}

func (s *EthereumService) Close() {
    s.client.Close()
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
    svc, err := services.NewEthereumService("https://ethereum.publicnode.com")
    if err != nil {
        log.Fatal(err)
    }
    defer svc.Close()

    // 查询余额
    balance, err := svc.GetBalance("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("ETH 余额: %s wei\n", balance.String())
}
```

---

## 关键知识点

| 概念 | 解释 |
|------|------|
| **wei** | ETH 最小单位，1 ETH = 10^18 wei |
| **nonce** | 地址的交易计数器，每发一笔 +1，防重放 |
| **Gas** | 执行交易的计算费用 = gasUsed × gasPrice |
| **gasLimit** | 愿意支付的最大 gas 量（ETH 转账固定 21000） |
| **gasPrice** | 每单位 gas 出价（wei），越高越快被打包 |
| **EIP-155** | 将 chainID 纳入签名，防止一条链的交易在另一条链重放 |
| **私钥 → 地址** | 私钥 → 椭圆曲线乘法 → 公钥 → Keccak-256 → 取后 20 字节 |
