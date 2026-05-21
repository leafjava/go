# 第9课：Gin 进阶 - 中间件、验证、错误处理

> 学习时间：3-4小时 | 难度：⭐⭐⭐

## 📋 本课目标

- 掌握 Gin 中间件的使用和自定义
- 学会参数验证和数据绑定
- 实现统一的错误处理和响应格式
- 理解 CORS 跨域处理

## 1. 中间件基础

### 内置中间件

```go
package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func main() {
    // gin.Default() 包含 Logger 和 Recovery 中间件
    r := gin.Default()
    
    // 或者使用 gin.New() 手动添加中间件
    r2 := gin.New()
    r2.Use(gin.Logger())    // 日志中间件
    r2.Use(gin.Recovery())  // 恢复中间件（panic 处理）
    
    r.GET("/", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "Hello"})
    })
    
    r.Run(":8080")
}
```

### 自定义中间件

```go
package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "net/http"
    "time"
)

// 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        // 处理请求前
        fmt.Printf("[%s] %s %s\n", 
            time.Now().Format("2006-01-02 15:04:05"),
            c.Request.Method,
            c.Request.URL.Path,
        )
        
        c.Next()  // 执行后续处理
        
        // 处理请求后
        latency := time.Since(start)
        fmt.Printf("耗时: %v\n", latency)
    }
}

// 认证中间件
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        
        if token == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "缺少认证令牌",
            })
            c.Abort()  // 终止后续处理
            return
        }
        
        // 验证 token（简化示例）
        if token != "Bearer valid-token" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "无效的令牌",
            })
            c.Abort()
            return
        }
        
        // 将用户信息存入上下文
        c.Set("user_id", 123)
        c.Next()
    }
}

func main() {
    r := gin.New()
    
    // 全局中间件
    r.Use(LoggerMiddleware())
    r.Use(gin.Recovery())
    
    // 公开路由
    r.GET("/public", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"message": "公开接口"})
    })
    
    // 需要认证的路由
    auth := r.Group("/api")
    auth.Use(AuthMiddleware())
    {
        auth.GET("/profile", func(c *gin.Context) {
            userID := c.GetInt("user_id")
            c.JSON(http.StatusOK, gin.H{
                "user_id": userID,
                "message": "个人信息",
            })
        })
    }
    
    r.Run(":8080")
}
```

## 2. CORS 跨域处理

```go
package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

// CORS 中间件
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Authorization, Accept, X-Requested-With")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }

        c.Next()
    }
}

func main() {
    r := gin.Default()
    r.Use(CORSMiddleware())
    
    r.GET("/api/data", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"data": "跨域请求成功"})
    })
    
    r.Run(":8080")
}
```

## 3. 参数验证

### 基本验证

```go
package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

type CreateUserRequest struct {
    Username string `json:"username" binding:"required,min=3,max=20"`
    Email    string `json:"email" binding:"required,email"`
    Age      int    `json:"age" binding:"required,gte=18,lte=100"`
    Password string `json:"password" binding:"required,min=6"`
}

func main() {
    r := gin.Default()
    
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
    
    r.Run(":8080")
}
```

### 自定义验证器

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/gin-gonic/gin/binding"
    "github.com/go-playground/validator/v10"
    "net/http"
    "strings"
)

// 自定义验证函数：验证以太坊地址
func validateEthAddress(fl validator.FieldLevel) bool {
    address := fl.Field().String()
    return len(address) == 42 && strings.HasPrefix(address, "0x")
}

type TransferRequest struct {
    From   string  `json:"from" binding:"required,eth_address"`
    To     string  `json:"to" binding:"required,eth_address"`
    Amount float64 `json:"amount" binding:"required,gt=0"`
}

func main() {
    r := gin.Default()
    
    // 注册自定义验证器
    if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
        v.RegisterValidation("eth_address", validateEthAddress)
    }
    
    r.POST("/transfer", func(c *gin.Context) {
        var req TransferRequest
        
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": err.Error(),
            })
            return
        }
        
        c.JSON(http.StatusOK, gin.H{
            "message": "转账成功",
            "data":    req,
        })
    })
    
    r.Run(":8080")
}
```

## 4. 统一响应格式

### 响应结构

```go
package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

// 统一响应结构
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}

// 成功响应
func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Code:    200,
        Message: "success",
        Data:    data,
    })
}

// 错误响应
func Error(c *gin.Context, code int, message string) {
    c.JSON(code, Response{
        Code:    code,
        Message: message,
    })
}

func main() {
    r := gin.Default()
    
    r.GET("/users", func(c *gin.Context) {
        users := []string{"Alice", "Bob", "Charlie"}
        Success(c, gin.H{"users": users})
    })
    
    r.GET("/error", func(c *gin.Context) {
        Error(c, http.StatusBadRequest, "参数错误")
    })
    
    r.Run(":8080")
}
```

## 5. 统一错误处理

```go
package main

import (
    "errors"
    "github.com/gin-gonic/gin"
    "net/http"
)

// 自定义错误类型
type AppError struct {
    Code    int
    Message string
}

func (e *AppError) Error() string {
    return e.Message
}

// 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        
        // 检查是否有错误
        if len(c.Errors) > 0 {
            err := c.Errors.Last().Err
            
            // 根据错误类型返回不同响应
            if appErr, ok := err.(*AppError); ok {
                c.JSON(appErr.Code, gin.H{
                    "code":    appErr.Code,
                    "message": appErr.Message,
                })
            } else {
                c.JSON(http.StatusInternalServerError, gin.H{
                    "code":    500,
                    "message": "服务器内部错误",
                })
            }
        }
    }
}

func main() {
    r := gin.New()
    r.Use(gin.Logger())
    r.Use(gin.Recovery())
    r.Use(ErrorHandler())
    
    r.GET("/error", func(c *gin.Context) {
        // 返回自定义错误
        c.Error(&AppError{
            Code:    400,
            Message: "参数验证失败",
        })
    })
    
    r.GET("/panic", func(c *gin.Context) {
        // 模拟 panic
        panic("something went wrong")
    })
    
    r.Run(":8080")
}
```

## 6. 限流中间件

```go
package main

import (
    "github.com/gin-gonic/gin"
    "net/http"
    "sync"
    "time"
)

// 简单的限流器
type RateLimiter struct {
    requests map[string][]time.Time
    mu       sync.Mutex
    limit    int
    window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
    return &RateLimiter{
        requests: make(map[string][]time.Time),
        limit:    limit,
        window:   window,
    }
}

func (rl *RateLimiter) Allow(key string) bool {
    rl.mu.Lock()
    defer rl.mu.Unlock()
    
    now := time.Now()
    
    // 清理过期请求
    if requests, exists := rl.requests[key]; exists {
        var valid []time.Time
        for _, t := range requests {
            if now.Sub(t) < rl.window {
                valid = append(valid, t)
            }
        }
        rl.requests[key] = valid
    }
    
    // 检查是否超过限制
    if len(rl.requests[key]) >= rl.limit {
        return false
    }
    
    // 记录新请求
    rl.requests[key] = append(rl.requests[key], now)
    return true
}

// 限流中间件
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 使用 IP 作为限流 key
        key := c.ClientIP()
        
        if !limiter.Allow(key) {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "请求过于频繁，请稍后再试",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

func main() {
    r := gin.Default()
    
    // 每分钟最多 10 个请求
    limiter := NewRateLimiter(10, time.Minute)
    r.Use(RateLimitMiddleware(limiter))
    
    r.GET("/api/data", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{"data": "success"})
    })
    
    r.Run(":8080")
}
```

## 📝 作业

### 作业1：完整的中间件系统

创建 `homework/day9/middleware-system`：

```go
// TODO: 实现以下中间件
// 1. 请求日志中间件（记录请求时间、方法、路径、耗时）
// 2. 认证中间件（JWT 验证）
// 3. 权限中间件（角色检查）
// 4. 限流中间件（IP 限流）
// 5. CORS 中间件
```

### 作业2：参数验证系统

```go
// TODO: 实现自定义验证器
// 1. 以太坊地址验证
// 2. TON 地址验证
// 3. 交易哈希验证
// 4. 金额范围验证
```

### 作业3：统一响应和错误处理

```go
// TODO: 实现统一的响应格式和错误处理
// 1. 成功响应（Success）
// 2. 错误响应（Error）
// 3. 分页响应（Paginated）
// 4. 自定义错误类型
```

## 🎯 检查点

- ✅ 能够创建和使用自定义中间件
- ✅ 掌握参数验证和自定义验证器
- ✅ 实现统一的响应格式
- ✅ 处理 CORS 跨域问题
- ✅ 实现限流功能

## ⏭️ 下一课

[第10课：GORM 数据库操作](./day10-gorm.md)
