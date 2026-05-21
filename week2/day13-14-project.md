# 第13-14课：实战项目 - SkillsBay API

> 学习时间：6-8小时 | 难度：⭐⭐⭐⭐

## 📋 项目目标

- 构建完整的 RESTful API
- 实现用户认证和权限控制
- 集成数据库和缓存
- 实现钱包和交易管理

## 1. 项目结构

```
skillsbay-api/
├── main.go
├── config/
│   ├── config.go
│   └── config.yaml
├── models/
│   ├── user.go
│   ├── wallet.go
│   ├── transaction.go
│   └── nft.go
├── handlers/
│   ├── auth.go
│   ├── user.go
│   ├── wallet.go
│   ├── transaction.go
│   └── nft.go
├── middleware/
│   ├── auth.go
│   ├── cors.go
│   ├── logger.go
│   └── rate_limit.go
├── services/
│   ├── user_service.go
│   ├── wallet_service.go
│   └── blockchain_service.go
├── database/
│   └── db.go
├── logger/
│   └── logger.go
├── utils/
│   ├── jwt.go
│   ├── response.go
│   └── validator.go
├── go.mod
└── go.sum
```

## 2. 核心模型

### models/user.go

```go
package models

import (
    "golang.org/x/crypto/bcrypt"
    "gorm.io/gorm"
)

type User struct {
    gorm.Model
    Username string   `gorm:"uniqueIndex;not null" json:"username"`
    Email    string   `gorm:"uniqueIndex;not null" json:"email"`
    Password string   `gorm:"not null" json:"-"`
    Role     string   `gorm:"default:'user'" json:"role"`
    IsActive bool     `gorm:"default:true" json:"is_active"`
    Wallets  []Wallet `gorm:"foreignKey:UserID" json:"wallets,omitempty"`
}

func (u *User) SetPassword(password string) error {
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return err
    }
    u.Password = string(hashedPassword)
    return nil
}

func (u *User) CheckPassword(password string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
    return err == nil
}
```

### models/wallet.go

```go
package models

import "gorm.io/gorm"

type Wallet struct {
    gorm.Model
    UserID       uint          `gorm:"index" json:"user_id"`
    Address      string        `gorm:"uniqueIndex;not null" json:"address"`
    Balance      float64       `gorm:"default:0" json:"balance"`
    Network      string        `gorm:"not null" json:"network"`
    User         User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Transactions []Transaction `gorm:"foreignKey:WalletID" json:"transactions,omitempty"`
}
```

### models/transaction.go

```go
package models

import "gorm.io/gorm"

type Transaction struct {
    gorm.Model
    WalletID uint    `gorm:"index" json:"wallet_id"`
    Hash     string  `gorm:"uniqueIndex;not null" json:"hash"`
    From     string  `gorm:"not null" json:"from"`
    To       string  `gorm:"not null" json:"to"`
    Amount   float64 `gorm:"not null" json:"amount"`
    GasUsed  int64   `json:"gas_used"`
    GasPrice float64 `json:"gas_price"`
    Status   string  `gorm:"default:'pending'" json:"status"`
    BlockNum int64   `json:"block_num"`
    Wallet   Wallet  `gorm:"foreignKey:WalletID" json:"wallet,omitempty"`
}
```

## 3. 数据库初始化

### database/db.go

```go
package database

import (
    "fmt"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "your-project/config"
    "your-project/models"
)

var DB *gorm.DB

func InitDB() error {
    var err error
    
    dsn := config.AppConfig.GetDatabaseDSN()
    
    switch config.AppConfig.Database.Driver {
    case "sqlite":
        DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
    // case "mysql":
    //     DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
    // case "postgres":
    //     DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    default:
        return fmt.Errorf("不支持的数据库驱动: %s", config.AppConfig.Database.Driver)
    }
    
    if err != nil {
        return err
    }
    
    // 自动迁移
    err = DB.AutoMigrate(
        &models.User{},
        &models.Wallet{},
        &models.Transaction{},
        &models.NFT{},
    )
    
    return err
}
```

## 4. 认证处理器

### handlers/auth.go

```go
package handlers

import (
    "github.com/gin-gonic/gin"
    "net/http"
    "your-project/database"
    "your-project/models"
    "your-project/utils"
)

type AuthHandler struct{}

type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=20"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
        return
    }
    
    // 检查用户是否存在
    var existingUser models.User
    if err := database.DB.Where("email = ? OR username = ?", req.Email, req.Username).First(&existingUser).Error; err == nil {
        utils.ErrorResponse(c, http.StatusConflict, "用户已存在")
        return
    }
    
    // 创建用户
    user := models.User{
        Username: req.Username,
        Email:    req.Email,
        Role:     "user",
    }
    
    if err := user.SetPassword(req.Password); err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "密码加密失败")
        return
    }
    
    if err := database.DB.Create(&user).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "用户创建失败")
        return
    }
    
    utils.SuccessResponse(c, gin.H{
        "user": user,
    })
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
        return
    }
    
    // 查询用户
    var user models.User
    if err := database.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
        utils.ErrorResponse(c, http.StatusUnauthorized, "邮箱或密码错误")
        return
    }
    
    // 验证密码
    if !user.CheckPassword(req.Password) {
        utils.ErrorResponse(c, http.StatusUnauthorized, "邮箱或密码错误")
        return
    }
    
    // 生成 token
    token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
    if err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "token 生成失败")
        return
    }
    
    utils.SuccessResponse(c, gin.H{
        "token": token,
        "user":  user,
    })
}
```

## 5. 钱包处理器

### handlers/wallet.go

```go
package handlers

import (
    "github.com/gin-gonic/gin"
    "net/http"
    "your-project/database"
    "your-project/models"
    "your-project/utils"
)

type WalletHandler struct{}

type CreateWalletRequest struct {
    Address string `json:"address" binding:"required"`
    Network string `json:"network" binding:"required"`
}

func (h *WalletHandler) CreateWallet(c *gin.Context) {
    userID := c.GetUint("user_id")
    
    var req CreateWalletRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
        return
    }
    
    // 检查钱包是否已存在
    var existingWallet models.Wallet
    if err := database.DB.Where("address = ?", req.Address).First(&existingWallet).Error; err == nil {
        utils.ErrorResponse(c, http.StatusConflict, "钱包已存在")
        return
    }
    
    wallet := models.Wallet{
        UserID:  userID,
        Address: req.Address,
        Network: req.Network,
        Balance: 0,
    }
    
    if err := database.DB.Create(&wallet).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "钱包创建失败")
        return
    }
    
    utils.SuccessResponse(c, gin.H{
        "wallet": wallet,
    })
}

func (h *WalletHandler) GetWallets(c *gin.Context) {
    userID := c.GetUint("user_id")
    
    var wallets []models.Wallet
    if err := database.DB.Where("user_id = ?", userID).Preload("Transactions").Find(&wallets).Error; err != nil {
        utils.ErrorResponse(c, http.StatusInternalServerError, "查询失败")
        return
    }
    
    utils.SuccessResponse(c, gin.H{
        "wallets": wallets,
    })
}

func (h *WalletHandler) GetWallet(c *gin.Context) {
    userID := c.GetUint("user_id")
    walletID := c.Param("id")
    
    var wallet models.Wallet
    if err := database.DB.Where("id = ? AND user_id = ?", walletID, userID).Preload("Transactions").First(&wallet).Error; err != nil {
        utils.ErrorResponse(c, http.StatusNotFound, "钱包不存在")
        return
    }
    
    utils.SuccessResponse(c, gin.H{
        "wallet": wallet,
    })
}
```

## 6. 工具函数

### utils/response.go

```go
package utils

import "github.com/gin-gonic/gin"

type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

func SuccessResponse(c *gin.Context, data interface{}) {
    c.JSON(200, Response{
        Code:    200,
        Message: "success",
        Data:    data,
    })
}

func ErrorResponse(c *gin.Context, code int, message string) {
    c.JSON(code, Response{
        Code:    code,
        Message: message,
    })
}
```

## 7. 主程序

### main.go

```go
package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "your-project/config"
    "your-project/database"
    "your-project/handlers"
    "your-project/logger"
    "your-project/middleware"
)

func main() {
    // 1. 加载配置
    if err := config.LoadConfig("config.yaml"); err != nil {
        panic("加载配置失败: " + err.Error())
    }
    
    // 2. 初始化日志
    if err := logger.InitLogger(
        config.AppConfig.Log.Level,
        config.AppConfig.Log.FilePath,
        config.AppConfig.Log.MaxSize,
        config.AppConfig.Log.MaxBackups,
        config.AppConfig.Log.MaxAge,
    ); err != nil {
        panic("初始化日志失败: " + err.Error())
    }
    defer logger.Sync()
    
    // 3. 初始化数据库
    if err := database.InitDB(); err != nil {
        panic("初始化数据库失败: " + err.Error())
    }
    
    // 4. 创建路由
    gin.SetMode(config.AppConfig.Server.Mode)
    r := gin.New()
    
    // 5. 全局中间件
    r.Use(middleware.LoggerMiddleware())
    r.Use(middleware.RecoveryMiddleware())
    r.Use(middleware.CORSMiddleware())
    
    // 6. 初始化处理器
    authHandler := &handlers.AuthHandler{}
    walletHandler := &handlers.WalletHandler{}
    
    // 7. 公开路由
    r.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })
    
    r.POST("/api/v1/auth/register", authHandler.Register)
    r.POST("/api/v1/auth/login", authHandler.Login)
    
    // 8. 需要认证的路由
    auth := r.Group("/api/v1")
    auth.Use(middleware.JWTAuth())
    {
        // 钱包路由
        wallets := auth.Group("/wallets")
        {
            wallets.POST("", walletHandler.CreateWallet)
            wallets.GET("", walletHandler.GetWallets)
            wallets.GET("/:id", walletHandler.GetWallet)
        }
    }
    
    // 9. 启动服务
    addr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
    logger.Logger.Info("服务启动", zap.String("addr", addr))
    
    if err := r.Run(addr); err != nil {
        logger.Logger.Fatal("服务启动失败", zap.Error(err))
    }
}
```

## 8. API 测试

### 注册用户

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "linshen",
    "email": "linshen@example.com",
    "password": "password123"
  }'
```

### 登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "linshen@example.com",
    "password": "password123"
  }'
```

### 创建钱包

```bash
curl -X POST http://localhost:8080/api/v1/wallets \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    "network": "Ethereum"
  }'
```

### 查询钱包

```bash
curl -X GET http://localhost:8080/api/v1/wallets \
  -H "Authorization: Bearer YOUR_TOKEN"
```

## 📝 作业

### 作业1：完善用户功能

```go
// TODO: 实现以下功能
// 1. 获取用户信息
// 2. 更新用户信息
// 3. 修改密码
// 4. 用户列表（管理员）
```

### 作业2：实现交易功能

```go
// TODO: 实现交易管理
// 1. 记录交易
// 2. 查询交易列表
// 3. 查询交易详情
// 4. 更新交易状态
```

### 作业3：实现 NFT 功能

```go
// TODO: 实现 NFT 管理
// 1. 铸造 NFT
// 2. 转移 NFT
// 3. 查询 NFT 列表
// 4. 查询用户的 NFT
```

### 作业4：添加单元测试

```go
// TODO: 编写单元测试
// 1. 用户注册测试
// 2. 用户登录测试
// 3. 钱包创建测试
// 4. API 集成测试
```

## 🎯 检查点

- ✅ 完成项目结构搭建
- ✅ 实现用户认证系统
- ✅ 实现钱包管理功能
- ✅ 集成数据库和日志
- ✅ 完成 API 测试

## 🎉 恭喜

你已经完成了第二周的学习，掌握了 Go Web 开发的核心技能！

## ⏭️ 下一周

[第三周：Web3 后端开发](../week3/day15-ethereum.md)
