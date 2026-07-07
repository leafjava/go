# 第25-28课：毕业项目 - 完整 Web3 后端

> 学习时间：12-16小时 | 难度：⭐⭐⭐⭐⭐

## 📋 项目目标

构建一个**生产级的多链 Web3 后端服务**，综合运用四周所学全部知识：

- Go 语言核心特性（Goroutine、Channel、接口）
- Gin + GORM Web 框架
- JWT 认证 + 权限控制
- 以太坊 + TON 双链集成
- Redis 缓存 + 消息队列
- 事件监听 + 交易管理
- 单元测试 + Docker 部署
- 日志和配置管理

## 1. 项目概述

### 项目名称：**ChainGate** — 多链区块链 API 网关

### 核心功能

1. **用户系统**：注册、登录、JWT 认证、角色权限
2. **钱包管理**：多链地址绑定、余额查询
3. **交易服务**：构建、签名、发送、追踪交易
4. **事件监控**：实时监听链上事件，大额交易告警
5. **代币管理**：ERC-20/Jetton 代币查询和转账
6. **数据统计**：Gas 趋势、交易量、热门合约

### 技术栈

| 层级 | 技术选型 |
|------|---------|
| 语言 | Go 1.22 |
| Web 框架 | Gin |
| ORM | GORM（SQLite 开发 / PostgreSQL 生产） |
| 认证 | golang-jwt |
| 缓存 | Redis（go-redis/v9） |
| 消息队列 | Redis Stream |
| 以太坊 | go-ethereum |
| TON | tonutils-go |
| 日志 | Zap + Lumberjack |
| 配置 | Viper |
| 部署 | Docker + Docker Compose |

## 2. 完整项目结构

```
chaingate/
├── cmd/
│   └── server/
│       └── main.go                     # 服务入口
├── config/
│   ├── config.go                       # 配置结构体 + 加载
│   ├── config.yaml                     # 默认配置
│   └── config.prod.yaml                # 生产配置
├── internal/
│   ├── blockchain/
│   │   ├── ethereum/
│   │   │   ├── client.go               # ETH 客户端
│   │   │   ├── balance.go              # 余额查询
│   │   │   ├── transaction.go          # 交易服务
│   │   │   ├── contract.go             # 合约调用
│   │   │   └── events.go               # 事件监听
│   │   ├── ton/
│   │   │   ├── client.go               # TON 客户端
│   │   │   ├── balance.go              # 余额查询
│   │   │   ├── transaction.go          # 转账服务
│   │   │   └── jetton.go               # Jetton 代币
│   │   └── factory.go                  # 客户端工厂
│   ├── model/
│   │   ├── user.go                     # 用户模型
│   │   ├── wallet.go                   # 钱包模型
│   │   ├── transaction.go              # 交易模型
│   │   └── alert.go                    # 告警模型
│   ├── handler/
│   │   ├── auth_handler.go             # 认证 API
│   │   ├── user_handler.go             # 用户 API
│   │   ├── wallet_handler.go           # 钱包 API
│   │   ├── transaction_handler.go      # 交易 API
│   │   ├── balance_handler.go          # 余额 API
│   │   ├── alert_handler.go            # 告警 API
│   │   └── health_handler.go           # 健康检查
│   ├── service/
│   │   ├── auth_service.go             # 认证服务
│   │   ├── user_service.go             # 用户服务
│   │   ├── wallet_service.go           # 钱包服务
│   │   ├── transaction_service.go      # 交易服务
│   │   ├── balance_service.go          # 余额服务
│   │   ├── alert_service.go            # 告警服务
│   │   └── monitor_service.go          # 监控服务
│   ├── middleware/
│   │   ├── auth.go                     # JWT 认证
│   │   ├── ratelimit.go               # 限流
│   │   ├── logger.go                   # 请求日志
│   │   └── recovery.go                 # panic 恢复
│   ├── queue/
│   │   ├── producer.go                 # 任务生产者
│   │   └── consumer.go                 # 任务消费者
│   └── database/
│       └── db.go                       # 数据库初始化
├── pkg/
│   ├── cache/
│   │   └── redis.go                    # Redis 封装
│   ├── logger/
│   │   └── logger.go                   # Zap 日志封装
│   └── utils/
│       ├── jwt.go                      # JWT 工具
│       ├── crypto.go                   # 加密工具
│       ├── address.go                  # 地址验证
│       ├── response.go                 # 统一响应
│       └── validator.go               # 参数验证
├── scripts/
│   ├── seed.go                         # 数据库种子脚本
│   └── migrate.go                      # 数据库迁移脚本
├── test/
│   ├── integration/
│   │   ├── auth_test.go               # 认证集成测试
│   │   ├── wallet_test.go             # 钱包集成测试
│   │   └── transaction_test.go        # 交易集成测试
│   └── mock/
│       ├── mock_blockchain.go          # Mock 区块链客户端
│       └── mock_redis.go              # Mock Redis
├── Dockerfile
├── docker-compose.yml
├── docker-compose.prod.yml
├── .github/
│   └── workflows/
│       ├── ci.yml
│       └── deploy.yml
├── go.mod
├── go.sum
└── README.md
```

## 3. 数据模型

### internal/model/user.go

```go
package model

import (
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
)

type User struct {
    gorm.Model
    Username  string   `gorm:"uniqueIndex;size:50;not null" json:"username"`
    Email     string   `gorm:"uniqueIndex;size:100;not null" json:"email"`
    Password  string   `gorm:"not null" json:"-"`
    Role      string   `gorm:"size:20;default:'user'" json:"role"` // user, admin
    IsActive  bool     `gorm:"default:true" json:"is_active"`
    Wallets   []Wallet `gorm:"foreignKey:UserID" json:"wallets,omitempty"`
}

func (u *User) SetPassword(password string) error {
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.Password = string(hash)
    return nil
}

func (u *User) CheckPassword(password string) bool {
    return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)) == nil
}
```

### internal/model/wallet.go

```go
package model

import "gorm.io/gorm"

type Wallet struct {
    gorm.Model
    UserID  uint   `gorm:"index" json:"user_id"`
    Address string `gorm:"size:100;not null" json:"address"`
    Chain   string `gorm:"size:20;not null" json:"chain"` // ethereum, ton
    Label   string `gorm:"size:50" json:"label"`
    User    User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (Wallet) TableName() string {
    return "wallets"
}
```

### internal/model/transaction.go

```go
package model

import "gorm.io/gorm"

type Transaction struct {
    gorm.Model
    WalletID  uint    `gorm:"index" json:"wallet_id"`
    Hash      string  `gorm:"size:66;uniqueIndex" json:"hash"`
    From      string  `gorm:"size:100" json:"from"`
    To        string  `gorm:"size:100" json:"to"`
    Value     string  `gorm:"size:100" json:"value"`
    Chain     string  `gorm:"size:20" json:"chain"`
    TokenType string  `gorm:"size:20;default:'native'" json:"token_type"` // native, erc20, jetton
    Status    string  `gorm:"size:20;default:'pending'" json:"status"`   // pending, confirmed, failed
    GasUsed   uint64  `json:"gas_used"`
    GasPrice  string  `gorm:"size:50" json:"gas_price"`
    BlockNum  uint64  `json:"block_num"`
    Data      string  `gorm:"type:text" json:"data,omitempty"`
    Wallet    Wallet  `gorm:"foreignKey:WalletID" json:"wallet,omitempty"`
}

func (Transaction) TableName() string {
    return "transactions"
}
```

### internal/model/alert.go

```go
package model

import "gorm.io/gorm"

type AlertRule struct {
    gorm.Model
    Name        string  `gorm:"size:100;not null" json:"name"`
    Chain       string  `gorm:"size:20;not null" json:"chain"`
    TokenAddr   string  `gorm:"size:100" json:"token_address"`
    MinValue    float64 `json:"min_value"`    // 最小金额阈值
    WatchAddrs  string  `gorm:"type:text" json:"watch_addresses"` // JSON 数组
    IsActive    bool    `gorm:"default:true" json:"is_active"`
    UserID      uint    `gorm:"index" json:"user_id"`
}

type AlertLog struct {
    gorm.Model
    RuleID     uint   `gorm:"index" json:"rule_id"`
    TxHash     string `gorm:"size:66" json:"tx_hash"`
    FromAddr   string `gorm:"size:100" json:"from"`
    ToAddr     string `gorm:"size:100" json:"to"`
    Value      string `gorm:"size:100" json:"value"`
    BlockNum   uint64 `json:"block_num"`
    IsRead     bool   `gorm:"default:false" json:"is_read"`
}
```

## 4. 核心服务实现

### internal/service/auth_service.go

```go
package service

import (
    "context"
    "errors"
    "fmt"
    "time"

    "chaingate/internal/model"
    "chaingate/pkg/utils"

    "gorm.io/gorm"
)

type AuthService struct {
    db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
    return &AuthService{db: db}
}

type RegisterInput struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6,max=100"`
}

type LoginInput struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type AuthResult struct {
    Token string      `json:"token"`
    User  *model.User `json:"user"`
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*AuthResult, error) {
    // 检查用户是否已存在
    var count int64
    s.db.Model(&model.User{}).Where("email = ? OR username = ?", input.Email, input.Username).Count(&count)
    if count > 0 {
        return nil, errors.New("用户已存在")
    }

    // 创建用户
    user := model.User{
        Username: input.Username,
        Email:    input.Email,
        Role:     "user",
        IsActive: true,
    }
    if err := user.SetPassword(input.Password); err != nil {
        return nil, fmt.Errorf("密码加密失败: %w", err)
    }

    if err := s.db.Create(&user).Error; err != nil {
        return nil, fmt.Errorf("创建用户失败: %w", err)
    }

    // 生成 Token
    token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
    if err != nil {
        return nil, fmt.Errorf("生成 Token 失败: %w", err)
    }

    return &AuthResult{Token: token, User: &user}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthResult, error) {
    var user model.User
    if err := s.db.Where("email = ?", input.Email).First(&user).Error; err != nil {
        return nil, errors.New("邮箱或密码错误")
    }

    if !user.CheckPassword(input.Password) {
        return nil, errors.New("邮箱或密码错误")
    }

    if !user.IsActive {
        return nil, errors.New("账户已被禁用")
    }

    token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
    if err != nil {
        return nil, fmt.Errorf("生成 Token 失败: %w", err)
    }

    return &AuthResult{Token: token, User: &user}, nil
}
```

### internal/service/balance_service.go

```go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "chaingate/internal/blockchain"
    "chaingate/internal/model"

    "github.com/redis/go-redis/v9"
    "gorm.io/gorm"
)

type BalanceService struct {
    factory  *blockchain.ClientFactory
    db       *gorm.DB
    cache    *redis.Client
}

func NewBalanceService(factory *blockchain.ClientFactory, db *gorm.DB, cache *redis.Client) *BalanceService {
    return &BalanceService{factory: factory, db: db, cache: cache}
}

type BalanceResponse struct {
    Address   string  `json:"address"`
    Balance   string  `json:"balance"`
    Symbol    string  `json:"symbol"`
    Chain     string  `json:"chain"`
    CachedAt  int64   `json:"cached_at,omitempty"`
}

// GetBalance 获取余额（带缓存）
func (s *BalanceService) GetBalance(ctx context.Context, chain blockchain.ChainType, address string) (*BalanceResponse, error) {
    // 1. 查缓存
    cacheKey := fmt.Sprintf("balance:%s:%s", chain, address)
    if cached, err := s.cache.Get(ctx, cacheKey).Result(); err == nil {
        var resp BalanceResponse
        if json.Unmarshal([]byte(cached), &resp) == nil {
            return &resp, nil
        }
    }

    // 2. 查链上
    client, err := s.factory.GetClient(chain)
    if err != nil {
        return nil, err
    }

    info, err := client.GetBalance(address)
    if err != nil {
        return nil, fmt.Errorf("链上查询失败: %w", err)
    }

    resp := &BalanceResponse{
        Address:  info.Address,
        Balance:  info.Balance,
        Symbol:   info.Symbol,
        Chain:    info.Chain,
        CachedAt: time.Now().Unix(),
    }

    // 3. 写缓存（15秒过期，适合区块链）
    data, _ := json.Marshal(resp)
    s.cache.Set(ctx, cacheKey, data, 15*time.Second)

    return resp, nil
}

// GetUserBalances 获取用户所有钱包的余额
func (s *BalanceService) GetUserBalances(ctx context.Context, userID uint) ([]BalanceResponse, error) {
    var wallets []model.Wallet
    if err := s.db.Where("user_id = ?", userID).Find(&wallets).Error; err != nil {
        return nil, err
    }

    results := make([]BalanceResponse, 0, len(wallets))
    for _, w := range wallets {
        resp, err := s.GetBalance(ctx, blockchain.ChainType(w.Chain), w.Address)
        if err != nil {
            // 某个钱包查询失败不影响其他钱包
            continue
        }
        results = append(results, *resp)
    }

    return results, nil
}
```

### internal/service/monitor_service.go

```go
package service

import (
    "context"
    "encoding/json"
    "log"
    "math/big"
    "sync"

    "chaingate/internal/blockchain"
    "chaingate/internal/model"

    "github.com/redis/go-redis/v9"
    "gorm.io/gorm"
)

// MonitorService 区块链监控服务
type MonitorService struct {
    factory    *blockchain.ClientFactory
    db         *gorm.DB
    cache      *redis.Client
    alertCh    chan AlertNotification
    stopCh     chan struct{}
    wg         sync.WaitGroup
}

type AlertNotification struct {
    RuleID    uint
    TxHash    string
    FromAddr  string
    ToAddr    string
    Value     string
    Chain     string
    BlockNum  uint64
}

func NewMonitorService(factory *blockchain.ClientFactory, db *gorm.DB, cache *redis.Client) *MonitorService {
    return &MonitorService{
        factory: factory,
        db:      db,
        cache:   cache,
        alertCh: make(chan AlertNotification, 100),
        stopCh:  make(chan struct{}),
    }
}

// Start 启动监控
func (m *MonitorService) Start(ctx context.Context) {
    m.wg.Add(1)
    go m.monitorGasPrice(ctx)

    m.wg.Add(1)
    go m.monitorBlockNumber(ctx)

    m.wg.Add(1)
    go m.processAlertRules(ctx)

    log.Println("[Monitor] 监控服务已启动")
}

// Stop 停止监控
func (m *MonitorService) Stop() {
    close(m.stopCh)
    m.wg.Wait()
    log.Println("[Monitor] 监控服务已停止")
}

// monitorGasPrice 监控 Gas 价格并缓存
func (m *MonitorService) monitorGasPrice(ctx context.Context) {
    defer m.wg.Done()
    // 具体实现参考第18课
}

// monitorBlockNumber 监控最新区块号
func (m *MonitorService) monitorBlockNumber(ctx context.Context) {
    defer m.wg.Done()
    // 具体实现参考第18课
}

// processAlertRules 处理告警规则
func (m *MonitorService) processAlertRules(ctx context.Context) {
    defer m.wg.Done()

    for {
        select {
        case <-m.stopCh:
            return
        case alert := <-m.alertCh:
            // 保存告警日志
            alertLog := model.AlertLog{
                RuleID:   alert.RuleID,
                TxHash:   alert.TxHash,
                FromAddr: alert.FromAddr,
                ToAddr:   alert.ToAddr,
                Value:    alert.Value,
                BlockNum: alert.BlockNum,
            }
            m.db.Create(&alertLog)

            log.Printf("[告警] 大额交易: %s → %s, 金额: %s, 交易: %s",
                alert.FromAddr, alert.ToAddr, alert.Value, alert.TxHash)
        }
    }
}
```

## 5. API 路由设计

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

    "chaingate/config"
    "chaingate/internal/blockchain"
    "chaingate/internal/handler"
    "chaingate/internal/middleware"
    "chaingate/internal/model"
    "chaingate/internal/service"
    "chaingate/pkg/cache"
    "chaingate/pkg/logger"

    "github.com/gin-gonic/gin"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
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

    // 3. 初始化数据库
    db, err := gorm.Open(sqlite.Open(cfg.Database.DSN), &gorm.Config{})
    if err != nil {
        log.Fatal("数据库初始化失败:", err)
    }
    db.AutoMigrate(&model.User{}, &model.Wallet{}, &model.Transaction{}, &model.AlertRule{}, &model.AlertLog{})

    // 4. 初始化 Redis
    rdb, err := cache.NewRedisClient(cfg.Redis)
    if err != nil {
        logger.Logger.Warn("Redis 连接失败:", err)
    }

    // 5. 初始化区块链客户端
    factory, err := blockchain.NewClientFactory(cfg)
    if err != nil {
        log.Fatal("区块链客户端初始化失败:", err)
    }
    defer factory.Close()

    // 6. 初始化服务层
    authService := service.NewAuthService(db)
    balanceService := service.NewBalanceService(factory, db, rdb)
    monitorService := service.NewMonitorService(factory, db, rdb)
    go monitorService.Start(context.Background())
    defer monitorService.Stop()

    // 7. 创建路由
    gin.SetMode(cfg.Server.Mode)
    r := gin.New()
    r.Use(middleware.LoggerMiddleware())
    r.Use(middleware.RecoveryMiddleware())
    r.Use(middleware.CORSMiddleware())

    // 8. 注册路由
    // 健康检查
    healthHandler := handler.NewHealthHandler(factory, rdb)
    r.GET("/health", healthHandler.Check)

    // 公开路由
    authHandler := handler.NewAuthHandler(authService)
    r.POST("/api/v1/auth/register", authHandler.Register)
    r.POST("/api/v1/auth/login", authHandler.Login)

    // 需要认证的路由
    api := r.Group("/api/v1")
    api.Use(middleware.JWTAuth())
    {
        // 用户路由
        userHandler := handler.NewUserHandler(db)
        api.GET("/user/profile", userHandler.GetProfile)
        api.PUT("/user/profile", userHandler.UpdateProfile)

        // 钱包路由
        walletHandler := handler.NewWalletHandler(db)
        api.POST("/wallets", walletHandler.Create)
        api.GET("/wallets", walletHandler.List)
        api.DELETE("/wallets/:id", walletHandler.Delete)

        // 余额路由
        balanceHandler := handler.NewBalanceHandler(balanceService)
        api.GET("/balance", balanceHandler.GetBalance)
        api.GET("/balance/user", balanceHandler.GetUserBalances)

        // 交易路由
        transactionHandler := handler.NewTransactionHandler(service.NewTransactionService(factory, db, rdb))
        api.POST("/transactions/send", transactionHandler.Send)
        api.GET("/transactions/:hash", transactionHandler.GetByHash)
        api.GET("/transactions", transactionHandler.List)
    }

    // 管理员路由
    admin := r.Group("/api/v1/admin")
    admin.Use(middleware.JWTAuth(), middleware.RequireRole("admin"))
    {
        admin.GET("/users", userHandler.ListUsers)
        admin.GET("/stats", handler.NewStatsHandler(db, rdb).GetStats)
    }

    // 9. 启动服务
    addr := fmt.Sprintf(":%d", cfg.Server.Port)
    srv := &http.Server{Addr: addr, Handler: r}

    go func() {
        logger.Logger.Info("ChainGate 服务启动", zap.String("addr", addr))
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()

    // 10. 优雅关闭
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    logger.Logger.Info("正在关闭服务...")
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
    logger.Logger.Info("ChainGate 服务已关闭")
}
```

## 6. 完整 API 文档

### 认证 API

```bash
# 注册
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","email":"alice@example.com","password":"Test1234"}'

# 登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"alice@example.com","password":"Test1234"}'
```

### 钱包 API

```bash
# 创建钱包
curl -X POST http://localhost:8080/api/v1/wallets \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"address":"0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb","chain":"ethereum","label":"我的ETH钱包"}'

# 查询钱包列表
curl http://localhost:8080/api/v1/wallets \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 余额 API

```bash
# 查询余额
curl "http://localhost:8080/api/v1/balance?chain=ethereum&address=0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb" \
  -H "Authorization: Bearer YOUR_TOKEN"

# 查询用户全部钱包余额
curl http://localhost:8080/api/v1/balance/user \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 交易 API

```bash
# 发送交易
curl -X POST http://localhost:8080/api/v1/transactions/send \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "from_address": "0x...",
    "to_address": "0x...",
    "amount": "0.01",
    "chain": "ethereum",
    "private_key": "your-private-key"
  }'

# 查询交易
curl http://localhost:8080/api/v1/transactions/0x... \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 7. 实现路线图（4天计划）

### 第1天（第25课）：项目骨架 + 认证系统
- [ ] 创建项目目录结构
- [ ] 实现配置管理（Viper + YAML）
- [ ] 实现日志系统（Zap）
- [ ] 实现数据库初始化和模型
- [ ] 实现用户注册/登录 API
- [ ] 实现 JWT 认证中间件

### 第2天（第26课）：钱包管理 + 余额查询
- [ ] 实现钱包 CRUD API
- [ ] 实现区块链客户端工厂
- [ ] 实现以太坊客户端（连接、余额查询）
- [ ] 实现 TON 客户端（连接、余额查询）
- [ ] 实现余额缓存服务
- [ ] 实现多链余额查询 API

### 第3天（第27课）：交易服务 + 事件监控
- [ ] 实现交易构建、签名、发送
- [ ] 实现交易历史 API
- [ ] 实现 Redis 交易队列
- [ ] 实现事件监听服务
- [ ] 实现大额交易告警
- [ ] 实现管理员统计 API

### 第4天（第28课）：测试 + 部署 + 验收
- [ ] 编写核心服务单元测试
- [ ] 编写 API 集成测试
- [ ] 完善 Dockerfile 和 docker-compose
- [ ] 配置 GitHub Actions CI/CD
- [ ] 编写项目 README
- [ ] 整体测试和 Bug 修复

## 8. 验收标准

### 功能验收

| 功能 | 验收标准 |
|------|---------|
| 用户注册/登录 | 正常注册、登录，返回 JWT Token |
| JWT 鉴权 | 未登录访问返回 401 |
| 角色权限 | 普通用户不能访问管理员路由 |
| 钱包管理 | CRUD 操作正常，数据持久化 |
| ETH 余额查询 | 返回正确的 ETH 余额 |
| TON 余额查询 | 返回正确的 TON 余额 |
| 余额缓存 | 15秒内重复查询命中缓存 |
| 交易发送 | 交易成功发送到链上，返回 txHash |
| 交易查询 | 返回交易详情和状态 |
| 事件监控 | 正常监听链上事件并记录 |
| 健康检查 | /health 返回所有组件状态 |

### 技术验收

| 项目 | 标准 |
|------|------|
| 单元测试覆盖率 | ≥ 70% |
| 数据竞争 | `go test -race` 无报错 |
| Docker 镜像大小 | ≤ 50MB |
| 启动时间 | ≤ 5 秒 |
| 代码规范 | golangci-lint 无 error |

## 9. 项目扩展建议

完成基础版本后，可以进一步扩展：

1. **WebSocket 推送**：实现 `ws://localhost:8080/ws` 实时推送交易通知
2. **gRPC 接口**：为高性能场景添加 gRPC API
3. **Swagger 文档**：集成 swaggo 生成 API 文档
4. **Prometheus 监控**：暴露 `/metrics` 端点，集成 Prometheus + Grafana
5. **多语言错误信息**：使用 i18n 包支持中英文错误提示
6. **IPFS 集成**：支持上传和查询 IPFS 内容
7. **NFT 管理**：支持 ERC-721/NFT 查询
8. **多签钱包**：实现多签交易管理

## 📝 作业

### 完成毕业项目

按照实现路线图分 4 天完成 ChainGate 项目，要求：

1. **代码提交**：每天提交代码到 Git，至少 4 个 commit
2. **测试覆盖**：核心服务达到 70% 测试覆盖率
3. **Docker 部署**：项目可以通过 `docker compose up -d` 一键启动
4. **API 测试**：使用 curl 或 Postman 验证所有 API
5. **项目文档**：编写完整的 README.md（项目介绍、技术栈、快速开始、API 文档）

## 🎯 检查点

- ✅ 项目结构完整
- ✅ 用户认证系统正常
- ✅ 多链钱包管理正常
- ✅ 余额查询正常（带缓存）
- ✅ 交易服务正常
- ✅ 事件监控正常
- ✅ 单元测试覆盖率 ≥ 70%
- ✅ Docker 一键部署
- ✅ CI/CD 配置完成

## 🎓 毕业！

恭喜完成 4 周 Go 语言 Web3 全栈开发学习！

### 你已掌握的技能

| 周期 | 技能 |
|------|------|
| Week1 | Go 语法、Goroutine、Channel、指针、接口 |
| Week2 | Gin 框架、GORM、JWT、配置管理、日志 |
| Week3 | 以太坊集成、TON 集成、交易签名、事件监听、Redis |
| Week4 | 测试、性能优化、Docker、CI/CD、面试准备 |

### 后续学习建议

1. **深入区块链**：学习 Solidity、智能合约开发
2. **分布式系统**：学习微服务、Kubernetes、服务网格
3. **数据库进阶**：学习 PostgreSQL 高级特性、分库分表
4. **开源贡献**：参与 go-ethereum、tonutils-go 等开源项目
5. **真实项目**：参与或启动自己的 Web3 项目

---

**💪 继续前进，构建更好的 Web3 世界！**
