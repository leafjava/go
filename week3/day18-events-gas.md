# 第18课：事件监听 + Gas 估算

> 学习时间：3-4小时 | 难度：⭐⭐⭐⭐

## 📋 本课目标

- 掌握以太坊事件监听机制
- 学会实时监听和批量查询事件
- 理解 Gas 估算原理和优化策略
- 实现多链事件订阅服务
- 掌握 WebSocket 和 HTTP 两种监听方式

## 1. 事件日志基础

### 事件 Log 数据结构

```go
package main

import (
    "fmt"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
)

// Log 结构体详解
// 每个事件触发时产生一条 Log
func inspectLog(log *types.Log) {
    fmt.Printf("合约地址: %s\n", log.Address.Hex())        // 触发事件的合约地址
    fmt.Printf("主题 Topic[0]: %s\n", log.Topics[0].Hex())  // 事件签名 keccak256
    fmt.Printf("主题 Topic[1]: %s\n", log.Topics[1].Hex())  // 索引参数 1
    fmt.Printf("主题 Topic[2]: %s\n", log.Topics[2].Hex())  // 索引参数 2
    fmt.Printf("数据 Data: %x\n", log.Data)                  // 非索引参数（ABI 编码）
    fmt.Printf("区块号: %d\n", log.BlockNumber)
    fmt.Printf("交易哈希: %s\n", log.TxHash.Hex())
    fmt.Printf("Log 索引: %d\n", log.Index)
}

// 事件签名计算
func EventSignature() {
    // ERC-20 Transfer 事件
    // event Transfer(address indexed from, address indexed to, uint256 value)
    signature := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
    fmt.Printf("Transfer 事件签名: %s\n", signature.Hex())
    // 输出: 0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef
}
```

## 2. WebSocket 实时监听

### 监听 ERC-20 Transfer 事件

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
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)

func main() {
    // 使用 WebSocket 连接（实时监听必须用 WS）
    client, err := ethclient.Dial("wss://eth.llamarpc.com")
    if err != nil {
        log.Fatal("WebSocket 连接失败:", err)
    }
    defer client.Close()

    // USDT 合约地址（主网）
    contractAddress := common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7")

    // 计算 Transfer 事件签名
    transferEventSignature := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))
    
    // 创建过滤器
    query := ethereum.FilterQuery{
        Addresses: []common.Address{contractAddress},
        Topics: [][]common.Hash{
            {transferEventSignature}, // Topic[0] = Transfer 事件签名
        },
    }

    // 订阅日志
    logs := make(chan types.Log)
    sub, err := client.SubscribeFilterLogs(context.Background(), query, logs)
    if err != nil {
        log.Fatal("订阅失败:", err)
    }

    fmt.Println("开始监听 USDT Transfer 事件...")
    fmt.Println("按 Ctrl+C 停止监听\n")

    // 监听循环
    for {
        select {
        case err := <-sub.Err():
            log.Fatal("订阅出错:", err)
        case vLog := <-logs:
            // 解析事件
            handleTransferLog(vLog)
        }
    }
}

func handleTransferLog(vLog types.Log) {
    fmt.Printf("╔═══════════════════════════════╗\n")
    fmt.Printf("║ 新区块: %d\n", vLog.BlockNumber)
    fmt.Printf("║ 交易: %s\n", vLog.TxHash.Hex())
    fmt.Printf("║ 合约: %s\n", vLog.Address.Hex())

    // 解码 Topic（索引参数）
    // Topic[1] = from（address 类型，需要填充为 32 字节）
    fromAddress := common.BytesToAddress(vLog.Topics[1].Bytes())
    fmt.Printf("║ 发送方: %s\n", fromAddress.Hex())

    // Topic[2] = to
    toAddress := common.BytesToAddress(vLog.Topics[2].Bytes())
    fmt.Printf("║ 接收方: %s\n", toAddress.Hex())

    // 解码 Data（非索引参数，value uint256）
    value := new(big.Int).SetBytes(vLog.Data)
    fmt.Printf("║ 金额: %s\n", value.String())

    fmt.Printf("╚═══════════════════════════════╝\n")
}
```

### 监听多个合约和事件

```go
// MultiEventWatcher 多事件监听器
type MultiEventWatcher struct {
    client  *ethclient.Client
    subs    map[string]ethereum.Subscription
    logs    map[string]chan types.Log
    handler EventHandler
}

// EventHandler 事件处理器接口
type EventHandler interface {
    HandleTransfer(from, to common.Address, value *big.Int, blockNumber uint64)
    HandleApproval(owner, spender common.Address, value *big.Int, blockNumber uint64)
}

func NewMultiEventWatcher(wsURL string, handler EventHandler) (*MultiEventWatcher, error) {
    client, err := ethclient.Dial(wsURL)
    if err != nil {
        return nil, err
    }

    return &MultiEventWatcher{
        client:  client,
        subs:    make(map[string]ethereum.Subscription),
        logs:    make(map[string]chan types.Log),
        handler: handler,
    }, nil
}

// WatchTransfer 监听 Transfer 事件
func (w *MultiEventWatcher) WatchTransfer(contractAddresses []common.Address) error {
    transferSig := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

    query := ethereum.FilterQuery{
        Addresses: contractAddresses,
        Topics:    [][]common.Hash{{transferSig}},
    }

    logs := make(chan types.Log, 100)
    sub, err := w.client.SubscribeFilterLogs(context.Background(), query, logs)
    if err != nil {
        return err
    }

    w.subs["transfer"] = sub
    w.logs["transfer"] = logs

    // 启动监听协程
    go w.listenTransfer()

    return nil
}

// WatchApproval 监听 Approval 事件
func (w *MultiEventWatcher) WatchApproval(contractAddresses []common.Address) error {
    approvalSig := crypto.Keccak256Hash([]byte("Approval(address,address,uint256)"))

    query := ethereum.FilterQuery{
        Addresses: contractAddresses,
        Topics:    [][]common.Hash{{approvalSig}},
    }

    logs := make(chan types.Log, 100)
    sub, err := w.client.SubscribeFilterLogs(context.Background(), query, logs)
    if err != nil {
        return err
    }

    w.subs["approval"] = sub
    w.logs["approval"] = logs

    go w.listenApproval()

    return nil
}

func (w *MultiEventWatcher) listenTransfer() {
    for {
        select {
        case err := <-w.subs["transfer"].Err():
            log.Printf("Transfer 订阅出错: %v", err)
            return
        case vLog := <-w.logs["transfer"]:
            from := common.BytesToAddress(vLog.Topics[1].Bytes())
            to := common.BytesToAddress(vLog.Topics[2].Bytes())
            value := new(big.Int).SetBytes(vLog.Data)
            w.handler.HandleTransfer(from, to, value, vLog.BlockNumber)
        }
    }
}

func (w *MultiEventWatcher) listenApproval() {
    for {
        select {
        case err := <-w.subs["approval"].Err():
            log.Printf("Approval 订阅出错: %v", err)
            return
        case vLog := <-w.logs["approval"]:
            owner := common.BytesToAddress(vLog.Topics[1].Bytes())
            spender := common.BytesToAddress(vLog.Topics[2].Bytes())
            value := new(big.Int).SetBytes(vLog.Data)
            w.handler.HandleApproval(owner, spender, value, vLog.BlockNumber)
        }
    }
}
```

## 3. HTTP 批量查询事件

WebSocket 适合实时监听，HTTP 适合批量查询历史事件。

```go
// QueryHistoricalLogs 查询历史事件
func QueryHistoricalLogs(
    client *ethclient.Client,
    contractAddress common.Address,
    fromBlock, toBlock *big.Int,
) ([]types.Log, error) {
    transferSig := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

    query := ethereum.FilterQuery{
        FromBlock: fromBlock,
        ToBlock:   toBlock,
        Addresses: []common.Address{contractAddress},
        Topics: [][]common.Hash{
            {transferSig},
        },
    }

    logs, err := client.FilterLogs(context.Background(), query)
    if err != nil {
        return nil, fmt.Errorf("查询日志失败: %w", err)
    }

    return logs, nil
}

// QueryLogsWithPagination 分页查询事件（避免单次查询量过大）
func QueryLogsWithPagination(
    client *ethclient.Client,
    contractAddress common.Address,
    startBlock, endBlock uint64,
    blockStep uint64,
) ([]types.Log, error) {
    var allLogs []types.Log

    for from := startBlock; from < endBlock; from += blockStep {
        to := from + blockStep - 1
        if to > endBlock {
            to = endBlock
        }

        logs, err := QueryHistoricalLogs(
            client,
            contractAddress,
            new(big.Int).SetUint64(from),
            new(big.Int).SetUint64(to),
        )
        if err != nil {
            log.Printf("查询区块 %d-%d 失败: %v", from, to, err)
            continue
        }

        allLogs = append(allLogs, logs...)
        log.Printf("已查询区块 %d-%d，找到 %d 条事件", from, to, len(logs))

        // 避免请求过于频繁
        time.Sleep(200 * time.Millisecond)
    }

    return allLogs, nil
}
```

## 4. Gas 估算

### Gas 估算基础

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
)

func main() {
    client, err := ethclient.Dial("https://eth.llamarpc.com")
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // 1. 估算 ETH 转账 Gas
    ethGasLimit := EstimateETHTransferGas()
    fmt.Printf("ETH 转账 Gas Limit: %d\n", ethGasLimit)
    // 标准 ETH 转账固定 21000

    // 2. 估算 ERC-20 转账 Gas
    erc20GasLimit, err := EstimateERC20TransferGas(client, common.Address{}, common.Address{})
    if err != nil {
        log.Printf("估算 ERC-20 Gas 失败: %v", err)
    }
    fmt.Printf("ERC-20 转账估算 Gas: %d\n", erc20GasLimit)

    // 3. 获取当前 Gas 价格
    gasPrice, _ := client.SuggestGasPrice(context.Background())
    fmt.Printf("当前 Gas 价格: %d Wei (%d Gwei)\n", gasPrice.Int64(), gasPrice.Int64()/1e9)

    // 4. 计算交易成本
    costWei := new(big.Int).Mul(gasPrice, big.NewInt(int64(erc20GasLimit)))
    costETH := new(big.Float).Quo(
        new(big.Float).SetInt(costWei),
        big.NewFloat(1e18),
    )
    fmt.Printf("预估 ERC-20 转账成本: %s Wei = %.6f ETH\n", costWei.String(), costETH)
}

// EstimateETHTransferGas ETH 转账固定 21000
func EstimateETHTransferGas() uint64 {
    return 21000
}

// EstimateERC20TransferGas 估算 ERC-20 转账 Gas
func EstimateERC20TransferGas(
    client *ethclient.Client,
    from common.Address,
    tokenContract common.Address,
) (uint64, error) {
    // 构造 transfer 调用消息
    toAddress := common.HexToAddress("0x0000000000000000000000000000000000000001")
    transferData := buildERC20TransferData(toAddress, big.NewInt(1000000))

    msg := ethereum.CallMsg{
        From: from,
        To:   &tokenContract,
        Data: transferData,
    }

    // 估算 Gas
    estimated, err := client.EstimateGas(context.Background(), msg)
    if err != nil {
        return 0, fmt.Errorf("估算 Gas 失败: %w", err)
    }

    // 增加 30% 安全缓冲
    safeGasLimit := estimated * 130 / 100
    return safeGasLimit, nil
}
```

### Gas 优化策略

```go
// GasOptimizer Gas 优化器
type GasOptimizer struct {
    client *ethclient.Client
}

// GetOptimalGasPrice 获取最优 Gas 价格（比当前低 10%）
func (o *GasOptimizer) GetOptimalGasPrice() (*big.Int, error) {
    currentGasPrice, err := o.client.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }

    // 降低 10%，节省成本（如果网络不拥堵）
    optimal := new(big.Int).Mul(currentGasPrice, big.NewInt(90))
    optimal = optimal.Div(optimal, big.NewInt(100))

    return optimal, nil
}

// GetGasPriceLevels 获取多档 Gas 价格建议
type GasPriceLevel struct {
    Slow     *big.Int // 慢速（等待约 5 分钟）
    Standard *big.Int // 标准（等待约 30 秒）
    Fast     *big.Int // 快速（等待约 15 秒）
    Instant  *big.Int // 即时（下一个区块）
}

func (o *GasOptimizer) GetGasPriceLevels() (*GasPriceLevel, error) {
    base, err := o.client.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }

    return &GasPriceLevel{
        Slow:     new(big.Int).Mul(base, big.NewInt(80)).Div(base, big.NewInt(100)),
        Standard: base,
        Fast:     new(big.Int).Mul(base, big.NewInt(120)).Div(base, big.NewInt(100)),
        Instant:  new(big.Int).Mul(base, big.NewInt(150)).Div(base, big.NewInt(100)),
    }, nil
}

// IsGasPriceTooHigh 判断 Gas 是否过高（超过阈值）
func (o *GasOptimizer) IsGasPriceTooHigh(maxGwei int64) bool {
    gasPrice, err := o.client.SuggestGasPrice(context.Background())
    if err != nil {
        return true // 无法获取时默认认为过高
    }

    currentGwei := new(big.Int).Div(gasPrice, big.NewInt(1e9))
    return currentGwei.Int64() > maxGwei
}

// WaitForLowGas 等待 Gas 降到合适水平
func (o *GasOptimizer) WaitForLowGas(ctx context.Context, maxGwei int64, checkInterval time.Duration) error {
    ticker := time.NewTicker(checkInterval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return fmt.Errorf("等待 Gas 降低超时")
        case <-ticker.C:
            if !o.IsGasPriceTooHigh(maxGwei) {
                return nil
            }
            log.Printf("当前 Gas 较高（> %d Gwei），等待中...", maxGwei)
        }
    }
}
```

### Gas 费用估算速查表

| 操作 | Gas Limit | Gas Price (正常) | 费用估算（ETH） |
|------|-----------|-----------------|----------------|
| ETH 转账 | 21,000 | 20 Gwei | ~0.00042 |
| ERC-20 转账 | ~65,000 | 20 Gwei | ~0.0013 |
| ERC-20 approve | ~46,000 | 20 Gwei | ~0.00092 |
| Uniswap V2 Swap | ~150,000 | 20 Gwei | ~0.003 |
| Uniswap V3 Swap | ~200,000 | 20 Gwei | ~0.004 |
| NFT Mint | ~150,000 | 20 Gwei | ~0.003 |
| 合约部署 | ~500,000+ | 20 Gwei | ~0.01+ |

## 5. 事件监听服务封装

### services/event_service.go

```go
package services

import (
    "context"
    "fmt"
    "log"
    "math/big"
    "sync"

    "github.com/ethereum/go-ethereum"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)

// EventService 事件监听服务
type EventService struct {
    client *ethclient.Client
    subs   map[string]ethereum.Subscription
    mu     sync.RWMutex
}

// TransferEvent Transfer 事件
type TransferEvent struct {
    From        common.Address
    To          common.Address
    Value       *big.Int
    BlockNumber uint64
    TxHash      common.Hash
    ContractAddress common.Address
}

// EventCallback 事件回调
type EventCallback func(event TransferEvent)

// NewEventService 创建事件服务
func NewEventService(wsURL string) (*EventService, error) {
    client, err := ethclient.Dial(wsURL)
    if err != nil {
        return nil, fmt.Errorf("WebSocket 连接失败: %w", err)
    }

    return &EventService{
        client: client,
        subs:   make(map[string]ethereum.Subscription),
    }, nil
}

// SubscribeTransferEvents 订阅 Transfer 事件
func (s *EventService) SubscribeTransferEvents(
    contractAddresses []common.Address,
    callback EventCallback,
) (string, error) {
    transferSig := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

    query := ethereum.FilterQuery{
        Addresses: contractAddresses,
        Topics:    [][]common.Hash{{transferSig}},
    }

    logs := make(chan types.Log, 100)
    sub, err := s.client.SubscribeFilterLogs(context.Background(), query, logs)
    if err != nil {
        return "", fmt.Errorf("订阅失败: %w", err)
    }

    subID := fmt.Sprintf("transfer_%d", len(s.subs))
    s.mu.Lock()
    s.subs[subID] = sub
    s.mu.Unlock()

    // 启动监听
    go func() {
        for {
            select {
            case err := <-sub.Err():
                log.Printf("订阅 %s 出错: %v", subID, err)
                return
            case vLog := <-logs:
                event := TransferEvent{
                    From:            common.BytesToAddress(vLog.Topics[1].Bytes()),
                    To:              common.BytesToAddress(vLog.Topics[2].Bytes()),
                    Value:           new(big.Int).SetBytes(vLog.Data),
                    BlockNumber:     vLog.BlockNumber,
                    TxHash:          vLog.TxHash,
                    ContractAddress: vLog.Address,
                }
                callback(event)
            }
        }
    }()

    return subID, nil
}

// Unsubscribe 取消订阅
func (s *EventService) Unsubscribe(subID string) {
    s.mu.Lock()
    defer s.mu.Unlock()

    if sub, ok := s.subs[subID]; ok {
        sub.Unsubscribe()
        delete(s.subs, subID)
    }
}

// Close 关闭服务
func (s *EventService) Close() {
    s.mu.Lock()
    defer s.mu.Unlock()

    for id, sub := range s.subs {
        sub.Unsubscribe()
        delete(s.subs, id)
    }
    s.client.Close()
}
```

## 6. 完整示例：交易监控服务

```go
package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/signal"

    "your-project/services"

    "github.com/ethereum/go-ethereum/common"
)

func main() {
    // 创建事件服务
    eventService, err := services.NewEventService("wss://eth.llamarpc.com")
    if err != nil {
        log.Fatal(err)
    }
    defer eventService.Close()

    // 监控的合约列表
    monitoredContracts := []common.Address{
        common.HexToAddress("0xdAC17F958D2ee523a2206206994597C13D831ec7"), // USDT
        common.HexToAddress("0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"), // USDC
    }

    // 订阅事件
    subID, err := eventService.SubscribeTransferEvents(
        monitoredContracts,
        func(event services.TransferEvent) {
            fmt.Printf("\n📢 交易通知\n")
            fmt.Printf("  合约: %s\n", event.ContractAddress.Hex())
            fmt.Printf("  发送方: %s\n", event.From.Hex())
            fmt.Printf("  接收方: %s\n", event.To.Hex())
            fmt.Printf("  金额: %s\n", event.Value.String())
            fmt.Printf("  交易: %s\n", event.TxHash.Hex())
            fmt.Printf("  区块: %d\n", event.BlockNumber)

            // 在这里可以：
            // 1. 存入数据库
            // 2. 发送 WebSocket 通知给前端
            // 3. 推送企业微信/钉钉消息
            // 4. 触发后续业务流程
        },
    )
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("订阅成功，ID: %s\n", subID)
    fmt.Println("正在监听合约事件...")

    // 等待中断信号
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, os.Interrupt)
    <-sig

    fmt.Println("\n监听已停止")
}
```

## 📝 作业

### 作业1：大额交易监控

```go
// TODO: 实现大额交易监控
// 1. 监听 USDT/USDC Transfer 事件
// 2. 当单笔转账金额 > 100,000 USDT 时发出告警
// 3. 告警方式：打印日志 + 记录到文件
// 4. 过滤交易所已知地址的黑白名单
```

### 作业2：Gas 费用预测 API

```go
// TODO: 实现 Gas 费用预测 API
// GET /api/v1/gas/estimate
// 1. 返回当前 Gas 价格（慢速/标准/快速）
// 2. 估算指定操作的 Gas 总成本
// 3. 支持多种链（Ethereum、BSC、Polygon）
```

### 作业3：事件索引器

```go
// TODO: 实现简单的事件索引器
// 1. 从指定区块高度开始索引 Transfer 事件
// 2. 存入 SQLite 数据库
// 3. 提供查询 API：按地址、时间范围查询
// 4. 支持断点续传（下次从上次停止的区块继续）
```

## 🎯 检查点

- ✅ 理解事件日志结构
- ✅ 掌握 WebSocket 实时监听
- ✅ 能够批量查询历史事件
- ✅ 掌握 Gas 估算和优化
- ✅ 封装事件监听服务
- ✅ 实现交易监控系统

## ⏭️ 下一课

[第19课：Redis 缓存 + 消息队列](./day19-redis-mq.md)
