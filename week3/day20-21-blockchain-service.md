# 第20-21课：实战 - 区块链交互服务

> 学习时间：6-8小时 | 难度：⭐⭐⭐⭐⭐

## 📋 项目目标

- 构建完整的多链区块链交互服务
- 支持以太坊 + TON 双链
- 实现交易构建、签名、发送全流程
- 集成 Redis 缓存和异步任务队列
- 实现事件监听和交易追踪
- API 采用 RESTful 设计

## 1. 项目架构

```
blockchain-service/
├── cmd/
│   └── server/
│       └── main.go                  # 服务入口
├── config/
│   ├── config.go                    # 配置结构体 + 加载
│   ├── config.yaml                  # 默认配置
│   └── config.prod.yaml             # 生产配置
├── internal/
│   ├── blockchain/
│   │   ├── ethereum/
│   │   │   ├── client.go            # ETH 客户端连接
│   │   │   ├── balance.go           # 余额查询
│   │   │   ├── transaction.go       # 交易构建/签名/发送
│   │   │   ├── contract.go          # 合约调用
│   │   │   └── events.go            # 事件监听
│   │   ├── ton/
│   │   │   ├── client.go            # TON 客户端连接
│   │   │   ├── balance.go           # 余额查询
│   │   │   ├── transaction.go       # 转账
│   │   │   └── jetton.go            # Jetton 代币
│   │   └── factory.go               # 链工厂（根据 chain 类型创建客户端）
│   ├── service/
│   │   ├── wallet_service.go        # 钱包服务
│   │   ├── transaction_service.go   # 交易服务
│   │   ├── balance_service.go       # 余额服务
│   │   └── monitor_service.go       # 监控服务
│   ├── handler/
│   │   ├── wallet_handler.go        # 钱包 API
│   │   ├── transaction_handler.go   # 交易 API
│   │   ├── balance_handler.go       # 余额 API
│   │   └── health_handler.go        # 健康检查
│   ├── middleware/
│   │   ├── auth.go                  # JWT 认证
│   │   ├── ratelimit.go             # 限流
│   │   └── logger.go                # 日志
│   ├── model/
│   │   ├── wallet.go                # 钱包模型
│   │   ├── transaction.go           # 交易模型
│   │   └── request.go               # 请求/响应结构
│   └── queue/
│       ├── producer.go              # 任务发布
│       └── consumer.go              # 任务消费
├── pkg/
│   ├── cache/
│   │   └── redis.go                 # Redis 缓存封装
│   ├── logger/
│   │   └── logger.go                # 日志封装
│   └── utils/
│       ├── crypto.go                # 加密工具
│       ├── address.go               # 地址验证
│       └── response.go              # 统一响应
├── go.mod
├── go.sum
└── Dockerfile
```

## 2. 核心配置

### config/config.go

```go
package config

import (
    "fmt"
    "os"

    "github.com/spf13/viper"
)

// Config 全局配置
type Config struct {
    Server   ServerConfig   `mapstructure:"server"`
    Database DatabaseConfig `mapstructure:"database"`
    Redis    RedisConfig    `mapstructure:"redis"`
    Ethereum EthereumConfig `mapstructure:"ethereum"`
    TON      TONConfig      `mapstructure:"ton"`
    JWT      JWTConfig      `mapstructure:"jwt"`
    Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
    Port int    `mapstructure:"port"`
    Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
    Driver string `mapstructure:"driver"`
    DSN    string `mapstructure:"dsn"`
}

type RedisConfig struct {
    Addr     string `mapstructure:"addr"`
    Password string `mapstructure:"password"`
    DB       int    `mapstructure:"db"`
}

type EthereumConfig struct {
    RPCURL     string   `mapstructure:"rpc_url"`
    WSURL      string   `mapstructure:"ws_url"`
    ChainID    int64    `mapstructure:"chain_id"`
    WatchAddrs []string `mapstructure:"watch_addresses"`
}

type TONConfig struct {
    ConfigURL string `mapstructure:"config_url"`
    APIURL    string `mapstructure:"api_url"`
    APIKey    string `mapstructure:"api_key"`
}

type JWTConfig struct {
    Secret     string `mapstructure:"secret"`
    ExpireHour int    `mapstructure:"expire_hour"`
}

type LogConfig struct {
    Level      string `mapstructure:"level"`
    FilePath   string `mapstructure:"file_path"`
    MaxSize    int    `mapstructure:"max_size"`
    MaxBackups int    `mapstructure:"max_backups"`
    MaxAge     int    `mapstructure:"max_age"`
}

// LoadConfig 加载配置
func LoadConfig(configPath string) (*Config, error) {
    v := viper.New()

    // 默认值
    v.SetDefault("server.port", 8080)
    v.SetDefault("server.mode", "debug")
    v.SetDefault("database.driver", "sqlite")
    v.SetDefault("database.dsn", "blockchain.db")
    v.SetDefault("redis.addr", "localhost:6379")
    v.SetDefault("redis.db", 0)
    v.SetDefault("ethereum.chain_id", 1)
    v.SetDefault("jwt.expire_hour", 24)
    v.SetDefault("log.level", "info")
    v.SetDefault("log.max_size", 100)
    v.SetDefault("log.max_backups", 10)
    v.SetDefault("log.max_age", 30)

    // 读取配置文件
    if configPath != "" {
        v.SetConfigFile(configPath)
    } else {
        v.SetConfigName("config")
        v.SetConfigType("yaml")
        v.AddConfigPath(".")
        v.AddConfigPath("./config")
    }

    if err := v.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, fmt.Errorf("读取配置失败: %w", err)
        }
    }

    // 环境变量覆盖
    v.AutomaticEnv()
    v.SetEnvPrefix("BC") // BC_SERVER_PORT 覆盖 server.port

    var cfg Config
    if err := v.Unmarshal(&cfg); err != nil {
        return nil, fmt.Errorf("解析配置失败: %w", err)
    }

    return &cfg, nil
}
```

### config/config.yaml

```yaml
server:
  port: 8080
  mode: debug  # debug | release | test

database:
  driver: sqlite
  dsn: blockchain.db

redis:
  addr: localhost:6379
  password: ""
  db: 0

ethereum:
  rpc_url: https://eth.llamarpc.com
  ws_url: wss://eth.llamarpc.com
  chain_id: 1
  watch_addresses:
    - "0xdAC17F958D2ee523a2206206994597C13D831ec7"  # USDT
    - "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48"  # USDC

ton:
  config_url: https://ton.org/global.config.json
  api_url: https://toncenter.com/api/v2
  api_key: ""

jwt:
  secret: "change-me-in-production"
  expire_hour: 24

log:
  level: info
  file_path: logs/app.log
  max_size: 100
  max_backups: 10
  max_age: 30
```

## 3. 区块链客户端工厂

### internal/blockchain/factory.go

```go
package blockchain

import (
    "fmt"

    "blockchain-service/internal/blockchain/ethereum"
    "blockchain-service/internal/blockchain/ton"
    "blockchain-service/config"
)

// ChainType 链类型
type ChainType string

const (
    ChainEthereum ChainType = "ethereum"
    ChainTON      ChainType = "ton"
)

// BalanceInfo 余额信息
type BalanceInfo struct {
    Address string  `json:"address"`
    Balance string  `json:"balance"`
    Symbol  string  `json:"symbol"`
    Chain   string  `json:"chain"`
}

// TransactionResult 交易结果
type TransactionResult struct {
    Hash    string `json:"hash"`
    Chain   string `json:"chain"`
    Status  string `json:"status"`
}

// BlockchainClient 区块链客户端接口
type BlockchainClient interface {
    GetBalance(address string) (*BalanceInfo, error)
    SendTransaction(req TransactionRequest) (*TransactionResult, error)
    GetTransactionStatus(txHash string) (string, error)
    Chain() ChainType
    Close() error
}

// TransactionRequest 交易请求
type TransactionRequest struct {
    FromAddress string `json:"from_address"`
    ToAddress   string `json:"to_address"`
    Amount      string `json:"amount"`
    PrivateKey  string `json:"private_key"`
    TokenAddress string `json:"token_address,omitempty"` // ERC-20/Jetton 代币地址
}

// ClientFactory 客户端工厂
type ClientFactory struct {
    ethClient *ethereum.Client
    tonClient *ton.Client
}

// NewClientFactory 创建客户端工厂
func NewClientFactory(cfg *config.Config) (*ClientFactory, error) {
    factory := &ClientFactory{}

    // 初始化以太坊客户端
    if cfg.Ethereum.RPCURL != "" {
        ethClient, err := ethereum.NewClient(cfg.Ethereum.RPCURL, cfg.Ethereum.ChainID)
        if err != nil {
            return nil, fmt.Errorf("以太坊客户端初始化失败: %w", err)
        }
        factory.ethClient = ethClient
    }

    // 初始化 TON 客户端
    if cfg.TON.ConfigURL != "" || cfg.TON.APIURL != "" {
        tonClient, err := ton.NewClient(cfg.TON.ConfigURL, cfg.TON.APIURL, cfg.TON.APIKey)
        if err != nil {
            return nil, fmt.Errorf("TON 客户端初始化失败: %w", err)
        }
        factory.tonClient = tonClient
    }

    return factory, nil
}

// GetClient 根据链类型获取客户端
func (f *ClientFactory) GetClient(chain ChainType) (BlockchainClient, error) {
    switch chain {
    case ChainEthereum:
        if f.ethClient == nil {
            return nil, fmt.Errorf("以太坊客户端未配置")
        }
        return f.ethClient, nil
    case ChainTON:
        if f.tonClient == nil {
            return nil, fmt.Errorf("TON 客户端未配置")
        }
        return f.tonClient, nil
    default:
        return nil, fmt.Errorf("不支持的链类型: %s", chain)
    }
}

// Close 关闭所有客户端
func (f *ClientFactory) Close() {
    if f.ethClient != nil {
        f.ethClient.Close()
    }
    if f.tonClient != nil {
        f.tonClient.Close()
    }
}
```

## 4. 以太坊客户端实现

### internal/blockchain/ethereum/client.go

```go
package ethereum

import (
    "context"
    "fmt"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/ethclient"
)

// Client 以太坊客户端
type Client struct {
    client  *ethclient.Client
    chainID *big.Int
}

// NewClient 创建以太坊客户端
func NewClient(rpcURL string, chainID int64) (*Client, error) {
    client, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, fmt.Errorf("连接以太坊节点失败: %w", err)
    }

    return &Client{
        client:  client,
        chainID: big.NewInt(chainID),
    }, nil
}

func (c *Client) Chain() string {
    return "ethereum"
}

func (c *Client) Close() error {
    c.client.Close()
    return nil
}

// GetClient 获取底层客户端
func (c *Client) GetClient() *ethclient.Client {
    return c.client
}

// GetChainID 获取链 ID
func (c *Client) GetChainID() *big.Int {
    return c.chainID
}
```

### internal/blockchain/ethereum/balance.go

```go
package ethereum

import (
    "context"
    "fmt"
    "math"
    "math/big"

    "github.com/ethereum/go-ethereum/common"
)

// GetBalance 查询 ETH 余额
func (c *Client) GetBalance(address string) (*BalanceInfo, error) {
    addr := common.HexToAddress(address)

    balance, err := c.client.BalanceAt(context.Background(), addr, nil)
    if err != nil {
        return nil, fmt.Errorf("查询余额失败: %w", err)
    }

    // 转换为 ETH
    fbalance := new(big.Float).SetInt(balance)
    ethValue := new(big.Float).Quo(fbalance, big.NewFloat(math.Pow10(18)))

    return &BalanceInfo{
        Address: address,
        Balance: fmt.Sprintf("%.18f", ethValue),
        Symbol:  "ETH",
        Chain:   "ethereum",
    }, nil
}

// GetTokenBalance 查询 ERC-20 代币余额
func (c *Client) GetTokenBalance(tokenAddress, userAddress string) (*big.Int, error) {
    // ... 实现见第15、17课
    return nil, nil
}

// GetBlockNumber 获取最新区块号
func (c *Client) GetBlockNumber() (uint64, error) {
    return c.client.BlockNumber(context.Background())
}
```

## 5. 服务层

### internal/service/balance_service.go

```go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "blockchain-service/internal/blockchain"
    "blockchain-service/config"

    "github.com/redis/go-redis/v9"
)

// BalanceService 余额服务
type BalanceService struct {
    factory *blockchain.ClientFactory
    cache   *redis.Client
    cfg     *config.Config
}

// NewBalanceService 创建余额服务
func NewBalanceService(factory *blockchain.ClientFactory, cache *redis.Client, cfg *config.Config) *BalanceService {
    return &BalanceService{
        factory: factory,
        cache:   cache,
        cfg:     cfg,
    }
}

// GetBalance 获取余额（带缓存）
func (s *BalanceService) GetBalance(ctx context.Context, chain blockchain.ChainType, address string) (*blockchain.BalanceInfo, error) {
    // 1. 尝试从缓存获取
    cacheKey := fmt.Sprintf("balance:%s:%s", chain, address)
    cached, err := s.cache.Get(ctx, cacheKey).Result()
    if err == nil {
        var info blockchain.BalanceInfo
        if err := json.Unmarshal([]byte(cached), &info); err == nil {
            return &info, nil
        }
    }

    // 2. 获取区块链客户端
    client, err := s.factory.GetClient(chain)
    if err != nil {
        return nil, err
    }

    // 3. 查询链上余额
    info, err := client.GetBalance(address)
    if err != nil {
        return nil, fmt.Errorf("查询链上余额失败: %w", err)
    }

    // 4. 写入缓存（30秒过期）
    data, _ := json.Marshal(info)
    s.cache.Set(ctx, cacheKey, data, 30*time.Second)

    return info, nil
}

// GetMultiChainBalance 获取多链余额
func (s *BalanceService) GetMultiChainBalance(ctx context.Context, address string) (map[blockchain.ChainType]*blockchain.BalanceInfo, error) {
    results := make(map[blockchain.ChainType]*blockchain.BalanceInfo)

    chains := []blockchain.ChainType{blockchain.ChainEthereum, blockchain.ChainTON}
    for _, chain := range chains {
        info, err := s.GetBalance(ctx, chain, address)
        if err != nil {
            // 某条链不可用时跳过
            continue
        }
        results[chain] = info
    }

    return results, nil
}
```

### internal/service/monitor_service.go

```go
package service

import (
    "context"
    "log"
    "sync"

    "blockchain-service/internal/blockchain"

    "github.com/redis/go-redis/v9"
)

// MonitorService 监控服务
type MonitorService struct {
    factory   *blockchain.ClientFactory
    cache     *redis.Client
    stopCh    chan struct{}
    wg        sync.WaitGroup
}

// NewMonitorService 创建监控服务
func NewMonitorService(factory *blockchain.ClientFactory, cache *redis.Client) *MonitorService {
    return &MonitorService{
        factory: factory,
        cache:   cache,
        stopCh:  make(chan struct{}),
    }
}

// Start 启动监控
func (s *MonitorService) Start(ctx context.Context) {
    s.wg.Add(1)
    go s.monitorBlockNumber(ctx)

    s.wg.Add(1)
    go s.monitorGasPrice(ctx)

    log.Println("监控服务已启动")
}

// Stop 停止监控
func (s *MonitorService) Stop() {
    close(s.stopCh)
    s.wg.Wait()
    log.Println("监控服务已停止")
}

// monitorBlockNumber 监控最新区块号
func (s *MonitorService) monitorBlockNumber(ctx context.Context) {
    defer s.wg.Done()

    ticker := time.NewTicker(12 * time.Second) // 以太坊区块时间约12秒
    defer ticker.Stop()

    for {
        select {
        case <-s.stopCh:
            return
        case <-ticker.C:
            client, err := s.factory.GetClient(blockchain.ChainEthereum)
            if err != nil {
                continue
            }

            // 这里调用具体链的接口获取区块号
            // blockNumber := ...
            // s.cache.Set(ctx, "monitor:block_number", blockNumber, 30*time.Second)
        }
    }
}

// monitorGasPrice 监控 Gas 价格
func (s *MonitorService) monitorGasPrice(ctx context.Context) {
    defer s.wg.Done()

    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-s.stopCh:
            return
        case <-ticker.C:
            // 更新 Gas 价格缓存
        }
    }
}
```

## 6. API 处理器

### internal/handler/balance_handler.go

```go
package handler

import (
    "context"
    "net/http"

    "blockchain-service/internal/blockchain"
    "blockchain-service/internal/service"
    "blockchain-service/pkg/utils"

    "github.com/gin-gonic/gin"
)

// BalanceHandler 余额处理器
type BalanceHandler struct {
    balanceService *service.BalanceService
}

// NewBalanceHandler 创建余额处理器
func NewBalanceHandler(balanceService *service.BalanceService) *BalanceHandler {
    return &BalanceHandler{balanceService: balanceService}
}

// GetBalance 查询余额
// GET /api/v1/balance?chain=ethereum&address=0x...
func (h *BalanceHandler) GetBalance(c *gin.Context) {
    chain := blockchain.ChainType(c.Query("chain"))
    address := c.Query("address")

    if address == "" {
        utils.ErrorResponse(c, http.StatusBadRequest, "address 参数不能为空")
        return
    }

    if chain == "" {
        chain = blockchain.ChainEthereum
    }

    info, err := h.balanceService.GetBalance(context.Background(), chain, address)
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
        return
    }

    utils.SuccessResponse(c, info)
}

// GetMultiChainBalance 查询多链余额
// GET /api/v1/balance/multi?address=0x...
func (h *BalanceHandler) GetMultiChainBalance(c *gin.Context) {
    address := c.Query("address")
    if address == "" {
        utils.ErrorResponse(c, http.StatusBadRequest, "address 参数不能为空")
        return
    }

    results, err := h.balanceService.GetMultiChainBalance(context.Background(), address)
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
        return
    }

    utils.SuccessResponse(c, results)
}
```

## 7. 服务入口

### cmd/server/main.go

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "blockchain-service/config"
    "blockchain-service/internal/blockchain"
    "blockchain-service/internal/handler"
    "blockchain-service/internal/middleware"
    "blockchain-service/internal/service"
    "blockchain-service/pkg/cache"
    "blockchain-service/pkg/logger"

    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
)

func main() {
    // 1. 加载配置
    cfg, err := config.LoadConfig("config/config.yaml")
    if err != nil {
        log.Fatal("加载配置失败:", err)
    }

    // 2. 初始化日志
    if err := logger.InitLogger(cfg.Log); err != nil {
        log.Fatal("初始化日志失败:", err)
    }
    defer logger.Sync()

    // 3. 初始化 Redis
    rdb := redis.NewClient(&redis.Options{
        Addr:     cfg.Redis.Addr,
        Password: cfg.Redis.Password,
        DB:       cfg.Redis.DB,
    })
    if _, err := rdb.Ping(context.Background()).Result(); err != nil {
        logger.Logger.Warn("Redis 连接失败，缓存功能不可用:", err)
    }

    // 4. 初始化区块链客户端工厂
    factory, err := blockchain.NewClientFactory(cfg)
    if err != nil {
        log.Fatal("初始化区块链客户端失败:", err)
    }
    defer factory.Close()

    // 5. 初始化服务层
    balanceService := service.NewBalanceService(factory, rdb, cfg)

    // 6. 初始化监控服务
    monitorService := service.NewMonitorService(factory, rdb)
    go monitorService.Start(context.Background())
    defer monitorService.Stop()

    // 7. 创建路由
    gin.SetMode(cfg.Server.Mode)
    r := gin.New()
    r.Use(middleware.LoggerMiddleware())
    r.Use(middleware.RecoveryMiddleware())

    // 8. 注册路由
    healthHandler := handler.NewHealthHandler(factory)
    balanceHandler := handler.NewBalanceHandler(balanceService)

    r.GET("/health", healthHandler.Check)

    api := r.Group("/api/v1")
    {
        api.GET("/balance", balanceHandler.GetBalance)
        api.GET("/balance/multi", balanceHandler.GetMultiChainBalance)
    }

    // 9. 启动 HTTP 服务器
    addr := fmt.Sprintf(":%d", cfg.Server.Port)
    srv := &http.Server{
        Addr:    addr,
        Handler: r,
    }

    // 10. 优雅关闭
    go func() {
        logger.Logger.Info("服务启动", zap.String("addr", addr))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal("服务启动失败:", err)
        }
    }()

    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Logger.Info("正在关闭服务...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("服务关闭异常:", err)
    }

    logger.Logger.Info("服务已关闭")
}
```

## 8. API 测试

### 查询余额

```bash
# 以太坊余额
curl "http://localhost:8080/api/v1/balance?chain=ethereum&address=0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"

# TON 余额
curl "http://localhost:8080/api/v1/balance?chain=ton&address=EQCkR1cGZxYbG6QrGqKJ8HxSPfI3NZ9sNjCpHlLHyVTGc5Gq"

# 多链余额
curl "http://localhost:8080/api/v1/balance/multi?address=0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb"
```

### 健康检查

```bash
curl http://localhost:8080/health

# 返回示例：
# {
#   "status": "ok",
#   "ethereum": "connected",
#   "ton": "connected",
#   "redis": "connected",
#   "timestamp": "2026-07-07T10:30:00Z"
# }
```

## 📝 作业

### 作业1：交易历史 API

```go
// TODO: 实现交易历史查询
// 1. 设计 transactions 数据库表
// 2. 实现 GET /api/v1/transactions?address=xxx&chain=ethereum&page=1&limit=20
// 3. 从链上事件和数据库结合查询
// 4. 支持按日期范围、状态、代币类型过滤
```

### 作业2：WebSocket 推送服务

```go
// TODO: 实现 WebSocket 实时推送
// 1. 客户端连接 ws://localhost:8080/ws?address=0x...
// 2. 监听指定地址的交易事件
// 3. 实时推送给 WebSocket 客户端
// 4. 使用 channel 在 goroutine 间通信
```

### 作业3：交易模拟

```go
// TODO: 实现交易模拟（仿真）
// 1. 在发送交易前先模拟执行
// 2. 返回预估 Gas 费用和预期结果
// 3. 检测可能的失败原因（余额不足、合约报错等）
// POST /api/v1/transactions/simulate
```

### 作业4：费率管理

```go
// TODO: 实现费率管理功能
// 1. 查询不同链的交易费率
// 2. 统计用户交易费用
// 3. 实现费率预警（Gas 过高时通知）
// 4. 使用缓存减少 RPC 调用频率
```

## 🎯 检查点

- ✅ 完成项目架构搭建
- ✅ 实现多链客户端工厂
- ✅ 实现余额查询服务（带缓存）
- ✅ 实现监控服务
- ✅ 完成 RESTful API
- ✅ 集成 Redis 缓存
- ✅ 支持优雅关闭
- ✅ 完成 API 测试

## 🎉 恭喜

你已经完成了第三周的学习！掌握了以太坊、TON 区块链集成、Redis 缓存消息队列，并构建了完整的区块链交互服务！

## ⏭️ 下一周

[第四周：工程化 + 面试准备](../week4/day22-testing.md)
