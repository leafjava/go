# 第8课：Gin 框架入门

> 学习时间：3-4小时 | 难度：⭐⭐⭐

## 📋 本课目标

- 理解 Gin 框架的核心概念
- 搭建第一个 Gin Web 服务
- 掌握路由和请求处理
- 学会返回 JSON 响应

## 1. 安装 Gin

```bash
# 初始化项目
cd D:\webProject\go
mkdir skillsbay-api
cd skillsbay-api
go mod init skillsbay-api

# 安装 Gin
go get -u github.com/gin-gonic/gin
```

## 2. Hello World

### main.go

```go
package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func main() {
    // 创建 Gin 引擎
    r := gin.Default()
    
    // 定义路由
    r.GET("/", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "message": "Hello, Gin!",
        })
    })
    
    // 启动服务（默认 8080 端口）
    r.Run(":8080")
}
```

```bash
# 运行
go run main.go

# 测试
# 浏览器访问: http://localhost:8080
# 或使用 curl: curl http://localhost:8080
```

## 3. 路由定义

### 基本路由

```go
package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func main() {
    r := gin.Default()
    
    // GET 请求
    r.GET("/users", getUsers)
    
    // POST 请求
    r.POST("/users", createUser)
    
    // PUT 请求
    r.PUT("/users/:id", updateUser)
    
    // DELETE 请求
    r.DELETE("/users/:id", deleteUser)
    
    r.Run(":8080")
}

func getUsers(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "users": []string{"Alice", "Bob", "Charlie"},
    })
}

func createUser(c *gin.Context) {
    c.JSON(http.StatusCreated, gin.H{
        "message": "用户创建成功",
    })
}

func updateUser(c *gin.Context) {
    id := c.Param("id")
    c.JSON(http.StatusOK, gin.H{
        "message": "用户更新成功",
        "id":      id,
    })
}

func deleteUser(c *gin.Context) {
    id := c.Param("id")
    c.JSON(http.StatusOK, gin.H{
        "message": "用户删除成功",
        "id":      id,
    })
}
```

### 路由分组

```go
package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func main() {
    r := gin.Default()
    
    // API v1 分组
    v1 := r.Group("/api/v1")
    {
        v1.GET("/users", getUsers)
        v1.POST("/users", createUser)
        
        // 钱包相关
        wallets := v1.Group("/wallets")
        {
            wallets.GET("/:address", getWallet)
            wallets.GET("/:address/balance", getBalance)
            wallets.POST("/:address/transfer", transfer)
        }
    }
    
    // API v2 分组
    v2 := r.Group("/api/v2")
    {
        v2.GET("/users", getUsersV2)
    }
    
    r.Run(":8080")
}

func getWallet(c *gin.Context) {
    address := c.Param("address")
    c.JSON(http.StatusOK, gin.H{
        "address": address,
        "balance": 10.5,
    })
}

func getBalance(c *gin.Context) {
    address := c.Param("address")
    c.JSON(http.StatusOK, gin.H{
        "address": address,
        "balance": 10.5,
    })
}

func transfer(c *gin.Context) {
    address := c.Param("address")
    c.JSON(http.StatusOK, gin.H{
        "from":    address,
        "message": "转账成功",
    })
}

func getUsers(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"version": "v1"})
}

func createUser(c *gin.Context) {
    c.JSON(http.StatusCreated, gin.H{"version": "v1"})
}

func getUsersV2(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"version": "v2"})
}
```

## 4. 请求参数

### 路径参数

```go
// GET /users/:id
r.GET("/users/:id", func(c *gin.Context) {
    id := c.Param("id")
    c.JSON(http.StatusOK, gin.H{"id": id})
})
```

### 查询参数

```go
// GET /search?keyword=eth&page=1
r.GET("/search", func(c *gin.Context) {
    keyword := c.Query("keyword")
    page := c.DefaultQuery("page", "1")
    
    c.JSON(http.StatusOK, gin.H{
        "keyword": keyword,
        "page":    page,
    })
})
```

### 请求体（JSON）

```go
type CreateUserRequest struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
    Age   int    `json:"age" binding:"required,gte=18"`
}

r.POST("/users", func(c *gin.Context) {
    var req CreateUserRequest
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusCreated, gin.H{
        "message": "用户创建成功",
        "user":    req,
    })
})
```

## 5. Web3 实战：钱包 API

### 项目结构

```
skillsbay-api/
├── main.go
├── models/
│   └── wallet.go
├── handlers/
│   └── wallet.go
└── go.mod
```

### models/wallet.go

```go
package models

type Wallet struct {
    Address string  `json:"address"`
    Balance float64 `json:"balance"`
    Network string  `json:"network"`
}

type TransferRequest struct {
    To     string  `json:"to" binding:"required"`
    Amount float64 `json:"amount" binding:"required,gt=0"`
}
```

### handlers/wallet.go

```go
package handlers

import (
    "github.com/gin-gonic/gin"
    "net/http"
    "skillsbay-api/models"
)

// 模拟数据库
var wallets = map[string]*models.Wallet{
    "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb": {
        Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
        Balance: 10.5,
        Network: "Ethereum",
    },
}

// 获取钱包信息
func GetWallet(c *gin.Context) {
    address := c.Param("address")
    
    wallet, exists := wallets[address]
    if !exists {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "钱包不存在",
        })
        return
    }
    
    c.JSON(http.StatusOK, wallet)
}

// 获取余额
func GetBalance(c *gin.Context) {
    address := c.Param("address")
    
    wallet, exists := wallets[address]
    if !exists {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "钱包不存在",
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "address": address,
        "balance": wallet.Balance,
    })
}

// 转账
func Transfer(c *gin.Context) {
    from := c.Param("address")
    
    var req models.TransferRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": err.Error(),
        })
        return
    }
    
    // 验证发送方钱包
    fromWallet, exists := wallets[from]
    if !exists {
        c.JSON(http.StatusNotFound, gin.H{
            "error": "发送方钱包不存在",
        })
        return
    }
    
    // 验证余额
    if fromWallet.Balance < req.Amount {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "余额不足",
        })
        return
    }
    
    // 执行转账
    fromWallet.Balance -= req.Amount
    
    c.JSON(http.StatusOK, gin.H{
        "message": "转账成功",
        "from":    from,
        "to":      req.To,
        "amount":  req.Amount,
    })
}
```

### main.go

```go
package main

import (
    "github.com/gin-gonic/gin"
    "skillsbay-api/handlers"
)

func main() {
    r := gin.Default()
    
    // API 路由
    api := r.Group("/api/v1")
    {
        wallets := api.Group("/wallets")
        {
            wallets.GET("/:address", handlers.GetWallet)
            wallets.GET("/:address/balance", handlers.GetBalance)
            wallets.POST("/:address/transfer", handlers.Transfer)
        }
    }
    
    r.Run(":8080")
}
```

### 测试 API

```bash
# 1. 获取钱包信息
curl http://localhost:8080/api/v1/wallets/0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb

# 2. 获取余额
curl http://localhost:8080/api/v1/wallets/0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb/balance

# 3. 转账
curl -X POST http://localhost:8080/api/v1/wallets/0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb/transfer \
  -H "Content-Type: application/json" \
  -d '{"to":"0x8ba1f109551bD432803012645Ac136ddd64DBA72","amount":1.5}'
```

## 📝 作业

### 作业1：NFT API

创建 `homework/day8/nft-api`：

```go
// TODO: 实现 NFT API
// 1. GET /api/v1/nfts - 获取所有 NFT
// 2. GET /api/v1/nfts/:id - 获取指定 NFT
// 3. POST /api/v1/nfts - 创建 NFT
// 4. GET /api/v1/nfts/owner/:address - 获取用户的 NFT
```

### 作业2：交易查询 API

```go
// TODO: 实现交易查询 API
// 1. GET /api/v1/transactions - 获取交易列表（支持分页）
// 2. GET /api/v1/transactions/:hash - 获取交易详情
// 3. GET /api/v1/transactions/address/:address - 获取地址的交易
```

## 🎯 检查点

- ✅ 能够创建 Gin Web 服务
- ✅ 掌握路由定义和分组
- ✅ 能够处理各种请求参数
- ✅ 能够返回 JSON 响应

## ⏭️ 下一课

[第9课：路由、中间件、参数绑定](./day9-gin-advanced.md)
