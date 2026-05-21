# 第15课：Go 调用以太坊合约

> 学习时间：3-4小时 | 难度：⭐⭐⭐⭐

## 📋 本课目标

- 掌握 go-ethereum 库的使用
- 学会连接以太坊节点
- 实现余额查询和转账
- 调用智能合约

## 1. 安装 go-ethereum

```bash
go get github.com/ethereum/go-ethereum
```

## 2. 连接以太坊节点

### 基本连接

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math/big"
    
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/common"
)

func main() {
    // 连接以太坊节点
    client, err := ethclient.Dial("https://eth.llamarpc.com")
    if err != nil {
        log.Fatal("连接失败:", err)
    }
    defer client.Close()
    
    fmt.Println("连接成功")
    
    // 获取最新区块号
    blockNumber, err := client.BlockNumber(context.Background())
    if err != nil {
        log.Fatal("获取区块号失败:", err)
    }
    
    fmt.Printf("最新区块: %d\n", blockNumber)
}
```

## 3. 查询余额

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math"
    "math/big"
    
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

func main() {
    client, err := ethclient.Dial("https://eth.llamarpc.com")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // 地址
    address := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
    
    // 查询余额（Wei）
    balance, err := client.BalanceAt(context.Background(), address, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("余额（Wei）: %s\n", balance.String())
    
    // 转换为 ETH
    fbalance := new(big.Float)
    fbalance.SetString(balance.String())
    ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))
    
    fmt.Printf("余额（ETH）: %f\n", ethValue)
}
```

## 4. 发送交易

### 转账 ETH

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

func main() {
    client, err := ethclient.Dial("https://eth.llamarpc.com")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // 私钥（注意：实际应用中应该加密存储）
    privateKey, err := crypto.HexToECDSA("your-private-key-without-0x")
    if err != nil {
        log.Fatal(err)
    }
    
    publicKey := privateKey.Public()
    publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
    if !ok {
        log.Fatal("无法转换公钥")
    }
    
    fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
    
    // 获取 nonce
    nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        log.Fatal(err)
    }
    
    // 转账金额（0.01 ETH）
    value := big.NewInt(10000000000000000) // 0.01 ETH in wei
    
    // Gas 限制
    gasLimit := uint64(21000)
    
    // Gas 价格
    gasPrice, err := client.SuggestGasPrice(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    
    // 接收地址
    toAddress := common.HexToAddress("0x8ba1f109551bD432803012645Ac136ddd64DBA72")
    
    // 创建交易
    tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
    
    // 获取链 ID
    chainID, err := client.NetworkID(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    
    // 签名交易
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
    if err != nil {
        log.Fatal(err)
    }
    
    // 发送交易
    err = client.SendTransaction(context.Background(), signedTx)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("交易已发送: %s\n", signedTx.Hash().Hex())
}
```

## 5. 调用智能合约

### ERC20 代币余额查询

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math/big"
    
    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ethereum/go-ethereum/accounts/abi"
    "strings"
)

// ERC20 ABI（简化版）
const erc20ABI = `[
    {
        "constant": true,
        "inputs": [{"name": "_owner", "type": "address"}],
        "name": "balanceOf",
        "outputs": [{"name": "balance", "type": "uint256"}],
        "type": "function"
    },
    {
        "constant": true,
        "inputs": [],
        "name": "decimals",
        "outputs": [{"name": "", "type": "uint8"}],
        "type": "function"
    }
]`

func main() {
    client, err := ethclient.Dial("https://eth.llamarpc.com")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // USDT 合约地址
    tokenAddress := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
    
    // 用户地址
    userAddress := common.HexToAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
    
    // 解析 ABI
    parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
    if err != nil {
        log.Fatal(err)
    }
    
    // 编码 balanceOf 调用
    data, err := parsedABI.Pack("balanceOf", userAddress)
    if err != nil {
        log.Fatal(err)
    }
    
    // 调用合约
    msg := ethereum.CallMsg{
        To:   &tokenAddress,
        Data: data,
    }
    
    result, err := client.CallContract(context.Background(), msg, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // 解码结果
    balance := new(big.Int)
    balance.SetBytes(result)
    
    fmt.Printf("USDT 余额: %s\n", balance.String())
}
```

## 6. 监听事件

```go
package main

import (
    "context"
    "fmt"
    "log"
    "math/big"
    
    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/ethclient"
)

func main() {
    client, err := ethclient.Dial("wss://eth.llamarpc.com")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()
    
    // 合约地址
    contractAddress := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")
    
    // 创建过滤器
    query := ethereum.FilterQuery{
        Addresses: []common.Address{contractAddress},
    }
    
    // 订阅日志
    logs := make(chan types.Log)
    sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("开始监听事件...")
    
    for {
        select {
        case err := <-sub.Err():
            log.Fatal(err)
        case vLog := <-logs:
            fmt.Printf("区块: %d, 交易: %s\n", vLog.BlockNumber, vLog.TxHash.Hex())
        }
    }
}
```

## 7. Web3 服务封装

### services/ethereum_service.go

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

func (s *EthereumService) GetBlockNumber() (uint64, error) {
    return s.client.BlockNumber(context.Background())
}

func (s *EthereumService) SendTransaction(privateKeyHex, toAddress string, amount *big.Int) (string, error) {
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
    
    nonce, err := s.client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return "", err
    }
    
    gasLimit := uint64(21000)
    gasPrice, err := s.client.SuggestGasPrice(context.Background())
    if err != nil {
        return "", err
    }
    
    toAddr := common.HexToAddress(toAddress)
    tx := types.NewTransaction(nonce, toAddr, amount, gasLimit, gasPrice, nil)
    
    chainID, err := s.client.NetworkID(context.Background())
    if err != nil {
        return "", err
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
    if err != nil {
        return "", err
    }
    
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

## 📝 作业

### 作业1：余额查询 API

创建 `homework/day15/balance-api`：

```go
// TODO: 实现余额查询 API
// GET /api/v1/ethereum/balance/:address
// 返回 ETH 余额和 USD 价值
```

### 作业2：交易查询

```go
// TODO: 实现交易查询
// 1. 根据交易哈希查询交易详情
// 2. 查询地址的交易历史
// 3. 计算交易手续费
```

### 作业3：ERC20 代币管理

```go
// TODO: 实现 ERC20 代币功能
// 1. 查询代币余额
// 2. 转账代币
// 3. 查询代币信息（名称、符号、精度）
```

## 🎯 检查点

- ✅ 能够连接以太坊节点
- ✅ 掌握余额查询
- ✅ 能够发送交易
- ✅ 能够调用智能合约
- ✅ 理解事件监听

## ⏭️ 下一课

[第16课：TON 区块链集成](./day16-ton.md)
