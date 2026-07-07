# 第17课：交易构建、签名、发送

> 学习时间：3-4小时 | 难度：⭐⭐⭐⭐

## 📋 本课目标

- 深入理解以太坊交易结构
- 掌握 EIP-1559 动态费用机制
- 学会交易签名和离线签名
- 实现批量交易和 ERC-20 转账
- 理解交易生命周期和状态追踪

## 1. 以太坊交易结构

### Legacy 交易（EIP-155 之前）

```go
package main

import (
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/core/types"
)

func main() {
    // Legacy 交易结构
    tx := types.NewTransaction(
        0,                          // nonce（发送方交易计数）
        common.HexToAddress("..."),  // to（接收方地址）
        big.NewInt(1000000000000000000), // value（转账金额，1 ETH = 10^18 Wei）
        21000,                       // gasLimit
        big.NewInt(20000000000),     // gasPrice（20 Gwei）
        nil,                         // data（合约调用数据）
    )

    fmt.Printf("交易哈希: %s\n", tx.Hash().Hex())
    fmt.Printf("交易类型: Legacy\n")
    fmt.Printf("Gas 费用上限: %d\n", tx.Gas())
    fmt.Printf("Gas 价格: %s Wei\n", tx.GasPrice().String())
}
```

### EIP-1559 动态费用交易

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
)

func main() {
    client, err := ethclient.Dial("https://eth.llamarpc.com")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 获取当前的 Gas 费用建议
    tipCap, _ := client.SuggestGasTipCap(context.Background())  // maxPriorityFeePerGas
    feeCap, _ := client.SuggestGasPrice(context.Background())     // maxFeePerGas（作为参考）

    chainID, _ := client.NetworkID(context.Background())

    // 创建 EIP-1559 动态费用交易
    tx := types.NewTx(&types.DynamicFeeTx{
        ChainID:   chainID,
        Nonce:     0,
        To:        &common.Address{},
        Value:     big.NewInt(1000000000000000000),
        Gas:       21000,
        GasFeeCap: feeCap,       // maxFeePerGas（最高愿意支付的单价）
        GasTipCap: tipCap,       // maxPriorityFeePerGas（给矿工的小费）
        Data:      nil,
    })

    fmt.Printf("交易类型: DynamicFee (EIP-1559)\n")
    fmt.Printf("MaxFeePerGas: %s Wei\n", feeCap.String())
    fmt.Printf("MaxPriorityFeePerGas: %s Wei\n", tipCap.String())
    fmt.Printf("链 ID: %d\n", chainID)
}
```

### 交易类型对比

| 特性 | Legacy (EIP-155) | DynamicFee (EIP-1559) |
|------|------------------|-----------------------|
| 费用模型 | gasPrice（单一价格） | maxFeePerGas + maxPriorityFeePerGas |
| 烧毁机制 | 无 | 烧毁 baseFee |
| 费用预估 | 困难，容易多付 | 更可预测 |
| 交易加速 | 需要重新发送 | 提高 tip 即可 |
| 兼容性 | 所有链 | Post-London 硬分叉 |

## 2. 交易签名

### 在线签名（完整流程）

```go
package main

import (
    "context"
    "crypto/ecdsa"
    "fmt"
    "log"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)

// BuildAndSignTransaction 构建并签名交易
func BuildAndSignTransaction(
    client *ethclient.Client,
    privateKeyHex string,
    toAddress string,
    amountWei *big.Int,
    data []byte,
) (*types.Transaction, error) {
    // 1. 解析私钥
    privateKey, err := crypto.HexToECDSA(privateKeyHex)
    if err != nil {
        return nil, fmt.Errorf("私钥解析失败: %w", err)
    }

    // 2. 获取发送方地址
    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        return nil, fmt.Errorf("公钥类型转换失败")
    }
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

    // 3. 获取 nonce
    nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, fmt.Errorf("获取 nonce 失败: %w", err)
    }

    // 4. 获取链 ID
    chainID, err := client.NetworkID(context.Background())
    if err != nil {
        return nil, fmt.Errorf("获取链 ID 失败: %w", err)
    }

    // 5. 估算 Gas（如果是合约调用）
    gasLimit := uint64(21000) // 默认 ETH 转账
    if len(data) > 0 {
        msg := ethereum.CallMsg{
            From:  fromAddress,
            To:    &common.HexToAddress(toAddress),
            Data:  data,
            Value: amountWei,
        }
        estimated, err := client.EstimateGas(context.Background(), msg)
        if err != nil {
            return nil, fmt.Errorf("估算 Gas 失败: %w", err)
        }
        gasLimit = estimated + estimated/5 // 增加 20% 缓冲
    }

    // 6. 获取 Gas 费用
    gasTipCap, err := client.SuggestGasTipCap(context.Background())
    if err != nil {
        return nil, fmt.Errorf("获取 GasTipCap 失败: %w", err)
    }

    gasFeeCap, err := client.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, fmt.Errorf("获取 GasFeeCap 失败: %w", err)
    }

    // 7. 构建交易
    toAddr := common.HexToAddress(toAddress)
    tx := types.NewTx(&types.DynamicFeeTx{
        ChainID:   chainID,
        Nonce:     nonce,
        GasTipCap: gasTipCap,
        GasFeeCap: gasFeeCap,
        Gas:       gasLimit,
        To:        &toAddr,
        Value:     amountWei,
        Data:      data,
    })

    // 8. 签名
    signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
    if err != nil {
        return nil, fmt.Errorf("签名失败: %w", err)
    }

    return signedTx, nil
}
```

### 离线签名（不需要网络连接）

```go
// OfflineSignTransaction 离线签名交易
func OfflineSignTransaction(
    privateKeyHex string,
    toAddress string,
    amountWei *big.Int,
    nonce uint64,
    chainID *big.Int,
    gasLimit uint64,
    gasTipCap *big.Int,
    gasFeeCap *big.Int,
    data []byte,
) (*types.Transaction, error) {
    // 解析私钥（无需连接节点）
    privateKey, err := crypto.HexToECDSA(privateKeyHex)
    if err != nil {
        return nil, fmt.Errorf("私钥解析失败: %w", err)
    }

    toAddr := common.HexToAddress(toAddress)

    // 构建交易
    tx := types.NewTx(&types.DynamicFeeTx{
        ChainID:   chainID,
        Nonce:     nonce,
        GasTipCap: gasTipCap,
        GasFeeCap: gasFeeCap,
        Gas:       gasLimit,
        To:        &toAddr,
        Value:     amountWei,
        Data:      data,
    })

    // 离线签名
    signer := types.LatestSignerForChainID(chainID)
    signedTx, err := types.SignTx(tx, signer, privateKey)
    if err != nil {
        return nil, fmt.Errorf("签名失败: %w", err)
    }

    return signedTx, nil
}
```

## 3. ERC-20 代币转账

### ERC-20 transfer ABI 编码

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math/big"
    "strings"

    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/accounts/abi"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)

// ERC-20 transfer 函数的 ABI 定义
const erc20ABI = `[
    {
        "constant": false,
        "inputs": [
            {"name": "_to", "type": "address"},
            {"name": "_value", "type": "uint256"}
        ],
        "name": "transfer",
        "outputs": [{"name": "success", "type": "bool"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [{"name": "_owner", "type": "address"}],
        "name": "balanceOf",
        "outputs": [{"name": "balance", "type": "uint256"}],
        "type": "function"
    }
]`

// SendERC20Token 发送 ERC-20 代币
func SendERC20Token(
    client *ethclient.Client,
    privateKeyHex string,
    tokenContractAddress string,
    toAddress string,
    amount *big.Int,
) (string, error) {
    // 1. 解析私钥
    privateKey, err := crypto.HexToECDSA(privateKeyHex)
    if err != nil {
        return "", fmt.Errorf("私钥解析失败: %w", err)
    }

    // 2. 获取发送方地址
    publicKey := privateKey.Public()
    publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

    // 3. 解析 ABI
    parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
    if err != nil {
        return "", fmt.Errorf("ABI 解析失败: %w", err)
    }

    // 4. 编码 transfer 函数调用
    toAddr := common.HexToAddress(toAddress)
    data, err := parsedABI.Pack("transfer", toAddr, amount)
    if err != nil {
        return "", fmt.Errorf("ABI 编码失败: %w", err)
    }

    // 5. 获取 nonce
    nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return "", fmt.Errorf("获取 nonce 失败: %w", err)
    }

    // 6. 估算 Gas
    tokenAddr := common.HexToAddress(tokenContractAddress)
    msg := ethereum.CallMsg{
        From: fromAddress,
        To:   &tokenAddr,
        Data: data,
    }
    gasLimit, err := client.EstimateGas(context.Background(), msg)
    if err != nil {
        return "", fmt.Errorf("估算 Gas 失败: %w", err)
    }
    gasLimit = gasLimit * 12 / 10 // 增加 20% 缓冲

    // 7. 获取 Gas 费用
    gasTipCap, _ := client.SuggestGasTipCap(context.Background())
    gasFeeCap, _ := client.SuggestGasPrice(context.Background())

    // 8. 获取链 ID
    chainID, _ := client.NetworkID(context.Background())

    // 9. 构建交易（to 是代币合约地址，value 为 0，data 是 transfer 调用）
    tx := types.NewTx(&types.DynamicFeeTx{
        ChainID:   chainID,
        Nonce:     nonce,
        GasTipCap: gasTipCap,
        GasFeeCap: gasFeeCap,
        Gas:       gasLimit,
        To:        &tokenAddr,
        Value:     big.NewInt(0), // ETH 转账金额为 0
        Data:      data,
    })

    // 10. 签名
    signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
    if err != nil {
        return "", fmt.Errorf("签名失败: %w", err)
    }

    // 11. 发送
    err = client.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return "", fmt.Errorf("发送失败: %w", err)
    }

    return signedTx.Hash().Hex(), nil
}
```

## 4. 批量交易

### 并发发送多笔交易

```go
// BatchTransfer 批量转账
type TransferRequest struct {
    To     string
    Amount *big.Int
}

type TransferResult struct {
    Index int
    Hash  string
    Error error
}

func BatchTransfer(
    client *ethclient.Client,
    privateKeyHex string,
    requests []TransferRequest,
) ([]TransferResult, error) {
    privateKey, err := crypto.HexToECDSA(privateKeyHex)
    if err != nil {
        return nil, fmt.Errorf("私钥解析失败: %w", err)
    }

    publicKey := privateKey.Public()
    publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

    // 获取初始 nonce
    nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, fmt.Errorf("获取 nonce 失败: %w", err)
    }

    chainID, _ := client.NetworkID(context.Background())
    gasTipCap, _ := client.SuggestGasTipCap(context.Background())
    gasFeeCap, _ := client.SuggestGasPrice(context.Background())

    results := make([]TransferResult, len(requests))
    var mu sync.Mutex
    var wg sync.WaitGroup

    // 并发构建和签名交易（注意 nonce 的顺序）
    signedTxs := make([]*types.Transaction, len(requests))

    for i, req := range requests {
        toAddr := common.HexToAddress(req.To)
        tx := types.NewTx(&types.DynamicFeeTx{
            ChainID:   chainID,
            Nonce:     nonce + uint64(i), // nonce 递增
            GasTipCap: gasTipCap,
            GasFeeCap: gasFeeCap,
            Gas:       21000,
            To:        &toAddr,
            Value:     req.Amount,
            Data:      nil,
        })

        signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
        if err != nil {
            results[i] = TransferResult{Index: i, Error: err}
            continue
        }
        signedTxs[i] = signedTx
    }

    // 并发送交易
    for i, signedTx := range signedTxs {
        if signedTx == nil {
            continue
        }

        wg.Add(1)
        go func(idx int, tx *types.Transaction) {
            defer wg.Done()

            err := client.SendTransaction(context.Background(), tx)
            mu.Lock()
            if err != nil {
                results[idx] = TransferResult{Index: idx, Error: err}
            } else {
                results[idx] = TransferResult{Index: idx, Hash: tx.Hash().Hex()}
            }
            mu.Unlock()
        }(i, signedTx)
    }

    wg.Wait()
    return results, nil
}
```

## 5. 交易状态追踪

### 等待交易确认

```go
// WaitForTransaction 等待交易确认
func WaitForTransaction(
    ctx context.Context,
    client *ethclient.Client,
    txHash common.Hash,
    timeout time.Duration,
) (*types.Receipt, error) {
    ctx, cancel := context.WithTimeout(ctx, timeout)
    defer cancel()

    for {
        select {
        case <-ctx.Done():
            return nil, fmt.Errorf("等待交易确认超时: %s", txHash.Hex())
        default:
            receipt, err := client.TransactionReceipt(ctx, txHash)
            if err == nil {
                return receipt, nil
            }
            // 如果还没上链，继续等待
            if err == ethereum.NotFound {
                time.Sleep(2 * time.Second)
                continue
            }
            return nil, err
        }
    }
}

// CheckTransactionStatus 检查交易状态
func CheckTransactionStatus(receipt *types.Receipt) string {
    // status = 1 表示成功，0 表示失败
    if receipt.Status == 1 {
        return "success"
    }
    return "failed"
}

// GetTransactionCost 计算交易实际花费
func GetTransactionCost(receipt *types.Receipt, tx *types.Transaction) *big.Int {
    gasUsed := new(big.Int).SetUint64(receipt.GasUsed)
    // 对于 EIP-1559 交易，实际 gasPrice = baseFee + effectiveTip
    effectiveGasPrice := receipt.EffectiveGasPrice
    return new(big.Int).Mul(gasUsed, effectiveGasPrice)
}
```

### 完整交易生命周期流程

```
1. 构建交易 (Build)
   ├── 设置 nonce（从节点获取 PendingNonceAt）
   ├── 设置 Gas 参数（估算或手动指定）
   ├── 设置接收地址和金额
   └── 编码合约调用数据（如需要）

2. 签名交易 (Sign)
   ├── 使用 LatestSignerForChainID
   ├── EIP-155 重放保护（chainID 内嵌）
   └── 生成 v, r, s 签名值

3. 发送交易 (Send)
   ├── SendTransaction → 返回 txHash
   ├── 交易进入节点的 pending pool
   └── 节点广播给其他节点

4. 等待打包 (Pending)
   ├── 矿工/验证者选择交易打包
   ├── Gas 竞价影响打包优先级
   └── 交易进入区块

5. 确认 (Confirmed)
   ├── 获取 TransactionReceipt
   ├── 检查 status（1=成功，0=失败）
   ├── 读取 GasUsed 和 EffectiveGasPrice
   └── 读取 Logs（合约事件）

6. 最终确认 (Finalized)
   ├── 需要等待 N 个区块确认
   └── N >= 12（以太坊）或 >= 64（BSC）
```

## 6. 交易服务层封装

### services/transaction_service.go

```go
package services

import (
    "context"
    "crypto/ecdsa"
    "fmt"
    "math/big"
    "sync"
    "time"

    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)

// TransactionService 交易服务
type TransactionService struct {
    client *ethclient.Client
}

// NewTransactionService 创建交易服务
func NewTransactionService(rpcURL string) (*TransactionService, error) {
    client, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, fmt.Errorf("连接节点失败: %w", err)
    }
    return &TransactionService{client: client}, nil
}

// SendETH 发送 ETH
func (s *TransactionService) SendETH(
    privateKeyHex string,
    to string,
    amountWei *big.Int,
) (string, error) {
    privateKey, err := crypto.HexToECDSA(privateKeyHex)
    if err != nil {
        return "", err
    }

    publicKey := privateKey.Public()
    publicKeyECDSA := publicKey.(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

    nonce, err := s.client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return "", err
    }

    chainID, _ := s.client.NetworkID(context.Background())
    gasTipCap, _ := s.client.SuggestGasTipCap(context.Background())
    gasFeeCap, _ := s.client.SuggestGasPrice(context.Background())

    toAddr := common.HexToAddress(to)
    tx := types.NewTx(&types.DynamicFeeTx{
        ChainID:   chainID,
        Nonce:     nonce,
        GasTipCap: gasTipCap,
        GasFeeCap: gasFeeCap,
        Gas:       21000,
        To:        &toAddr,
        Value:     amountWei,
    })

    signedTx, err := types.SignTx(tx, types.LatestSignerForChainID(chainID), privateKey)
    if err != nil {
        return "", err
    }

    err = s.client.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return "", err
    }

    return signedTx.Hash().Hex(), nil
}

// GetTransaction 查询交易详情
func (s *TransactionService) GetTransaction(txHash string) (*TransactionDetail, error) {
    hash := common.HexToHash(txHash)

    tx, isPending, err := s.client.TransactionByHash(context.Background(), hash)
    if err != nil {
        return nil, fmt.Errorf("查询交易失败: %w", err)
    }

    detail := &TransactionDetail{
        Hash:      tx.Hash().Hex(),
        Value:     tx.Value().String(),
        Gas:       tx.Gas(),
        GasPrice:  tx.GasPrice().String(),
        Nonce:     tx.Nonce(),
        IsPending: isPending,
    }

    // 尝试获取收据
    receipt, err := s.client.TransactionReceipt(context.Background(), hash)
    if err == nil {
        detail.Status = receipt.Status
        detail.BlockNumber = receipt.BlockNumber.Uint64()
        detail.GasUsed = receipt.GasUsed
        detail.EffectiveGasPrice = receipt.EffectiveGasPrice.String()
        detail.Logs = receipt.Logs
    }

    return detail, nil
}

// TransactionDetail 交易详情
type TransactionDetail struct {
    Hash               string
    Value              string
    Gas                uint64
    GasPrice           string
    EffectiveGasPrice  string
    GasUsed            uint64
    Nonce              uint64
    Status             uint64
    BlockNumber        uint64
    IsPending          bool
    Logs               []*types.Log
}

// Close 关闭连接
func (s *TransactionService) Close() {
    s.client.Close()
}
```

## 7. 常见交易错误与解决

| 错误 | 原因 | 解决方案 |
|------|------|----------|
| `nonce too low` | nonce 已被使用 | 使用 `PendingNonceAt` 获取最新 nonce |
| `replacement transaction underpriced` | 替换交易 Gas 太低 | 提高 10% Gas 价格重新发送 |
| `insufficient funds` | 余额不足 | 检查 ETH 余额 >= value + gasLimit * gasPrice |
| `gas required exceeds allowance` | Gas 估算不足 | 增加 gasLimit，使用 EstimateGas |
| `execution reverted` | 合约执行失败 | 检查合约逻辑、输入参数 |
| `transaction underpriced` | Gas 价格低于节点接受的最小值 | 提高 Gas 价格 |

## 📝 作业

### 作业1：多签钱包交易

```go
// TODO: 实现多签钱包
// 1. 定义多签钱包结构体（owners, required 最少签名数）
// 2. 实现提交交易（submitTransaction）
// 3. 实现确认交易（confirmTransaction）
// 4. 实现执行交易（executeTransaction）
// 5. 验证足够确认数后才执行
type MultiSigWallet struct {
    Owners   []common.Address
    Required int
    // TODO: 添加交易列表和确认映射
}
```

### 作业2：交易加速/取消

```go
// TODO: 实现交易加速和取消功能
// 1. 加速：使用相同 nonce、更高 Gas 价格重新发送
// 2. 取消：使用相同 nonce、发送 0 ETH 给自己、更高 Gas 价格
func SpeedUpTransaction(client *ethclient.Client, privateKeyHex string, txHash common.Hash) (string, error) {
    // TODO
}
func CancelTransaction(client *ethclient.Client, privateKeyHex string, nonce uint64) (string, error) {
    // TODO
}
```

### 作业3：交易历史查询 API

```go
// TODO: 实现交易历史查询
// GET /api/v1/transactions/:address
// 1. 查询地址的最近交易（使用区块浏览器 API 或扫描区块）
// 2. 分页支持
// 3. 过滤（时间范围、交易类型）
```

## 🎯 检查点

- ✅ 理解 Ethereum 交易结构（Legacy vs EIP-1559）
- ✅ 掌握交易签名流程
- ✅ 能够发送 ETH 和 ERC-20 交易
- ✅ 了解批量交易实现
- ✅ 掌握交易状态追踪
- ✅ 封装交易服务层

## ⏭️ 下一课

[第18课：事件监听 + Gas 估算](./day18-events-gas.md)
