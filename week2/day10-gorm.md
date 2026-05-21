# 第10课：GORM 数据库操作

> 学习时间：3-4小时 | 难度：⭐⭐⭐

## 📋 本课目标

- 掌握 GORM 的基本使用
- 学会模型定义和数据库迁移
- 掌握 CRUD 操作
- 理解关联查询和事务处理

## 1. 安装和配置

```bash
# 安装 GORM
go get -u gorm.io/gorm
go get -u gorm.io/driver/sqlite  # SQLite 驱动
go get -u gorm.io/driver/mysql   # MySQL 驱动
go get -u gorm.io/driver/postgres # PostgreSQL 驱动
```

### 连接数据库

```go
package main

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "log"
)

func main() {
    // SQLite
    db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    if err != nil {
        log.Fatal("连接数据库失败:", err)
    }
    
    // MySQL
    // dsn := "user:password@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
    // db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    
    // PostgreSQL
    // dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable"
    // db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    
    log.Println("数据库连接成功")
}
```

## 2. 模型定义

### 基础模型

```go
package models

import (
    "gorm.io/gorm"
    "time"
)

// User 用户模型
type User struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    
    Username string `gorm:"uniqueIndex;not null" json:"username"`
    Email    string `gorm:"uniqueIndex" json:"email"`
    Password string `gorm:"not null" json:"-"`
    Age      int    `json:"age"`
    IsActive bool   `gorm:"default:true" json:"is_active"`
}

// 自定义表名
func (User) TableName() string {
    return "users"
}
```

### Web3 模型示例

```go
package models

import (
    "gorm.io/gorm"
    "time"
)

// Wallet 钱包模型
type Wallet struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    
    UserID  uint    `gorm:"index" json:"user_id"`
    Address string  `gorm:"uniqueIndex;not null" json:"address"`
    Balance float64 `gorm:"default:0" json:"balance"`
    Network string  `gorm:"not null" json:"network"` // Ethereum, TON, etc.
    
    // 关联
    User         User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Transactions []Transaction `gorm:"foreignKey:WalletID" json:"transactions,omitempty"`
}

// Transaction 交易模型
type Transaction struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    
    WalletID  uint    `gorm:"index" json:"wallet_id"`
    Hash      string  `gorm:"uniqueIndex;not null" json:"hash"`
    From      string  `gorm:"not null" json:"from"`
    To        string  `gorm:"not null" json:"to"`
    Amount    float64 `gorm:"not null" json:"amount"`
    GasUsed   int64   `json:"gas_used"`
    GasPrice  float64 `json:"gas_price"`
    Status    string  `gorm:"default:'pending'" json:"status"` // pending, confirmed, failed
    BlockNum  int64   `json:"block_num"`
    
    // 关联
    Wallet Wallet `gorm:"foreignKey:WalletID" json:"wallet,omitempty"`
}

// NFT 模型
type NFT struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    CreatedAt time.Time      `json:"created_at"`
    UpdatedAt time.Time      `json:"updated_at"`
    DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
    
    TokenID     int    `gorm:"not null" json:"token_id"`
    Name        string `gorm:"not null" json:"name"`
    Description string `json:"description"`
    ImageURL    string `json:"image_url"`
    Owner       string `gorm:"index;not null" json:"owner"`
    Contract    string `gorm:"index;not null" json:"contract"`
    Network     string `gorm:"not null" json:"network"`
}
```

## 3. 数据库迁移

```go
package main

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "log"
    "your-project/models"
)

func main() {
    db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    if err != nil {
        log.Fatal("连接数据库失败:", err)
    }
    
    // 自动迁移（创建表）
    err = db.AutoMigrate(
        &models.User{},
        &models.Wallet{},
        &models.Transaction{},
        &models.NFT{},
    )
    if err != nil {
        log.Fatal("迁移失败:", err)
    }
    
    log.Println("数据库迁移成功")
}
```

## 4. CRUD 操作

### 创建（Create）

```go
package main

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "log"
    "your-project/models"
)

func main() {
    db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    
    // 创建单条记录
    user := models.User{
        Username: "linshen",
        Email:    "linshen@example.com",
        Password: "hashed_password",
        Age:      23,
    }
    
    result := db.Create(&user)
    if result.Error != nil {
        log.Fatal("创建失败:", result.Error)
    }
    
    log.Printf("创建成功，ID: %d, 影响行数: %d\n", user.ID, result.RowsAffected)
    
    // 批量创建
    users := []models.User{
        {Username: "alice", Email: "alice@example.com", Password: "pass1", Age: 25},
        {Username: "bob", Email: "bob@example.com", Password: "pass2", Age: 30},
    }
    
    db.Create(&users)
}
```

### 查询（Read）

```go
package main

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "log"
    "your-project/models"
)

func main() {
    db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    
    // 查询单条记录
    var user models.User
    
    // 根据主键查询
    db.First(&user, 1)  // SELECT * FROM users WHERE id = 1;
    
    // 根据条件查询
    db.Where("username = ?", "linshen").First(&user)
    
    // 查询多条记录
    var users []models.User
    db.Find(&users)  // SELECT * FROM users;
    
    // 条件查询
    db.Where("age > ?", 20).Find(&users)
    
    // 复杂查询
    db.Where("age > ? AND is_active = ?", 20, true).
        Order("created_at desc").
        Limit(10).
        Find(&users)
    
    // 查询指定字段
    db.Select("username", "email").Find(&users)
    
    // 统计
    var count int64
    db.Model(&models.User{}).Where("age > ?", 20).Count(&count)
    log.Printf("符合条件的用户数: %d\n", count)
}
```

### 更新（Update）

```go
package main

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "your-project/models"
)

func main() {
    db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    
    // 更新单个字段
    db.Model(&models.User{}).Where("id = ?", 1).Update("age", 24)
    
    // 更新多个字段
    db.Model(&models.User{}).Where("id = ?", 1).Updates(map[string]interface{}{
        "age":   24,
        "email": "newemail@example.com",
    })
    
    // 使用结构体更新
    var user models.User
    db.First(&user, 1)
    user.Age = 25
    user.Email = "updated@example.com"
    db.Save(&user)
    
    // 批量更新
    db.Model(&models.User{}).Where("age < ?", 18).Update("is_active", false)
}
```

### 删除（Delete）

```go
package main

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "your-project/models"
)

func main() {
    db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    
    // 软删除（设置 DeletedAt）
    db.Delete(&models.User{}, 1)
    
    // 永久删除
    db.Unscoped().Delete(&models.User{}, 1)
    
    // 批量删除
    db.Where("age < ?", 18).Delete(&models.User{})
    
    // 查询包含软删除的记录
    var users []models.User
    db.Unscoped().Find(&users)
}
```

## 5. 关联查询

### 一对多关系

```go
package main

import (
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "log"
    "your-project/models"
)

func main() {
    db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    
    // 预加载关联数据
    var wallet models.Wallet
    db.Preload("Transactions").First(&wallet, 1)
    
    log.Printf("钱包地址: %s, 交易数: %d\n", wallet.Address, len(wallet.Transactions))
    
    // 条件预加载
    db.Preload("Transactions", "status = ?", "confirmed").First(&wallet, 1)
    
    // 嵌套预加载
    db.Preload("Transactions.Wallet").First(&wallet, 1)
}
```

### 多对多关系

```go
package models

type User struct {
    ID    uint
    Name  string
    Roles []Role `gorm:"many2many:user_roles;"`
}

type Role struct {
    ID    uint
    Name  string
    Users []User `gorm:"many2many:user_roles;"`
}

// 使用
func main() {
    db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    
    // 创建用户和角色
    user := User{Name: "linshen"}
    role1 := Role{Name: "admin"}
    role2 := Role{Name: "user"}
    
    db.Create(&user)
    db.Create(&role1)
    db.Create(&role2)
    
    // 关联
    db.Model(&user).Association("Roles").Append(&role1, &role2)
    
    // 查询
    var u User
    db.Preload("Roles").First(&u, user.ID)
}
```

## 6. 事务处理

```go
package main

import (
    "errors"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "your-project/models"
)

func main() {
    db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    
    // 方式1：手动事务
    tx := db.Begin()
    
    if err := tx.Create(&models.User{Username: "test"}).Error; err != nil {
        tx.Rollback()
        return
    }
    
    if err := tx.Create(&models.Wallet{Address: "0x..."}).Error; err != nil {
        tx.Rollback()
        return
    }
    
    tx.Commit()
    
    // 方式2：自动事务
    err := db.Transaction(func(tx *gorm.DB) error {
        // 创建用户
        user := models.User{Username: "linshen"}
        if err := tx.Create(&user).Error; err != nil {
            return err
        }
        
        // 创建钱包
        wallet := models.Wallet{
            UserID:  user.ID,
            Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
            Network: "Ethereum",
        }
        if err := tx.Create(&wallet).Error; err != nil {
            return err
        }
        
        // 返回 nil 提交事务
        return nil
    })
    
    if err != nil {
        // 事务回滚
        return
    }
}
```

## 7. Web3 实战：钱包管理 API

### handlers/wallet.go

```go
package handlers

import (
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
    "net/http"
    "your-project/models"
)

type WalletHandler struct {
    DB *gorm.DB
}

// 创建钱包
func (h *WalletHandler) CreateWallet(c *gin.Context) {
    var req struct {
        UserID  uint   `json:"user_id" binding:"required"`
        Address string `json:"address" binding:"required"`
        Network string `json:"network" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    wallet := models.Wallet{
        UserID:  req.UserID,
        Address: req.Address,
        Network: req.Network,
        Balance: 0,
    }
    
    if err := h.DB.Create(&wallet).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "创建钱包失败"})
        return
    }
    
    c.JSON(http.StatusCreated, gin.H{
        "message": "钱包创建成功",
        "wallet":  wallet,
    })
}

// 获取钱包列表
func (h *WalletHandler) GetWallets(c *gin.Context) {
    userID := c.Query("user_id")
    
    var wallets []models.Wallet
    query := h.DB.Preload("Transactions")
    
    if userID != "" {
        query = query.Where("user_id = ?", userID)
    }
    
    if err := query.Find(&wallets).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "wallets": wallets,
    })
}

// 更新余额
func (h *WalletHandler) UpdateBalance(c *gin.Context) {
    id := c.Param("id")
    
    var req struct {
        Balance float64 `json:"balance" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if err := h.DB.Model(&models.Wallet{}).Where("id = ?", id).Update("balance", req.Balance).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{"message": "余额更新成功"})
}
```

## 📝 作业

### 作业1：完整的用户管理系统

创建 `homework/day10/user-system`：

```go
// TODO: 实现用户管理 API
// 1. 创建用户
// 2. 查询用户（支持分页、搜索）
// 3. 更新用户信息
// 4. 删除用户（软删除）
// 5. 用户统计
```

### 作业2：交易记录系统

```go
// TODO: 实现交易记录 API
// 1. 记录交易
// 2. 查询交易（按地址、状态、时间范围）
// 3. 更新交易状态
// 4. 交易统计（总金额、手续费等）
```

### 作业3：NFT 管理系统

```go
// TODO: 实现 NFT 管理 API
// 1. 铸造 NFT
// 2. 转移 NFT
// 3. 查询用户的 NFT
// 4. NFT 市场（上架、下架、购买）
```

## 🎯 检查点

- ✅ 能够定义 GORM 模型
- ✅ 掌握 CRUD 操作
- ✅ 理解关联查询
- ✅ 能够使用事务
- ✅ 能够集成 GORM 到 Gin 项目

## ⏭️ 下一课

[第11课：JWT 认证 + 权限控制](./day11-jwt.md)
