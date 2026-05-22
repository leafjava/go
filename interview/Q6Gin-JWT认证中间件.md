# Q6: Gin 如何做 JWT 认证 + 中间件？对比 Spring Boot

## 目录

- [整体流程](#整体流程)
- [第一步：JWT 工具函数](#第一步jwt-工具函数)
- [第二步：JWT 认证中间件](#第二步jwt-认证中间件)
- [第三步：角色权限中间件](#第三步角色权限中间件)
- [第四步：路由分组应用](#第四步路由分组应用)
- [完整项目结构](#完整项目结构)
- [补充一：双 Token 方案](#补充一双-token-方案)
- [补充二：退出登录 Redis 黑名单](#补充二退出登录-redis-黑名单)
- [与 Spring Boot 对比](#与-spring-boot-对比)

---

## 整体流程

```
登录请求 /login
  │
  ▼
验证用户名密码 → 调用 GenerateToken() 生成 JWT → 返回 token 给客户端
  │
  ▼
后续请求（带 Authorization: Bearer <token>）
  │
  ▼
JWTAuth 中间件 → 解析 token → c.Set() 注入用户信息
  │
  ├── 普通接口 → 直接进入 Handler
  │
  └── 管理员接口 → RequireRole 中间件 → 校验角色 → 进入 Handler
```

四个步骤按顺序实现：**工具函数 → 认证中间件 → 权限中间件 → 路由分组**。

---

## 第一步：JWT 工具函数

负责 token 的**生成**和**解析**，通常封装在 `utils/jwt.go` 中。

```go
package utils

import (
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
)

var (
    JwtSecret         = []byte("your-secret-key-change-in-production")
    AccessTokenExpire = 15 * time.Minute  // 短期
)

// Claims 自定义 JWT 载荷
type Claims struct {
    UserID   uint   `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

// GenerateToken 生成 access token
func GenerateToken(userID uint, username, role string) (string, error) {
    claims := Claims{
        UserID:   userID,
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpire)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "gin-jwt-demo",
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(JwtSecret)
}

// ParseToken 解析 token，返回 Claims
func ParseToken(tokenStr string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{},
        func(t *jwt.Token) (interface{}, error) {
            // 确保签名算法一致
            if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, errors.New("unexpected signing method")
            }
            return JwtSecret, nil
        },
    )
    if err != nil {
        return nil, err
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token")
    }

    return claims, nil
}
```

关键设计点：

| 设计点 | 说明 |
|--------|------|
| `jwt.RegisteredClaims` | 内嵌标准 claims，自带 `ExpiresAt`、`IssuedAt` 等字段 |
| `SigningMethodHMAC` | HS256 对称加密，单服务够用；多服务场景换 RS256 |
| 角色写入 claims | 无需每次查库，直接取 token 中的角色做权限判断 |
| 返回指针类型 `*Claims` | 方便调用方判空，避免零值误判 |

---

## 第二步：JWT 认证中间件

验证 token，把用户信息存入 `gin.Context`，下游 handler 和中间件都能取到。

```go
package middleware

import (
    "net/http"
    "strings"

    "your-project/utils"

    "github.com/gin-gonic/gin"
)

func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 从 Header 取出 Authorization 字段
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"msg": "请先登录"})
            c.Abort() // 阻断后续处理
            return
        }

        // 2. 校验 "Bearer xxx" 格式
        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"msg": "token 格式错误"})
            c.Abort()
            return
        }

        // 3. 解析 token
        claims, err := utils.ParseToken(parts[1])
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"msg": "token 无效或已过期"})
            c.Abort()
            return
        }

        // 4. 将用户信息注入 context，下游通过 c.Get() 获取
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("role", claims.Role)

        c.Next() // 通过，继续执行后续 handler
    }
}
```

你可能会有疑问：

**Q: 为什么用 `c.Set()` 而不是 `context.WithValue()`？**

Gin 的 `c.Set()` 底层是基于 `gin.Context.Keys`（一个 `map[string]interface{}`），用于**请求级别**的数据传递。它比 Go 标准库的 `context.WithValue()` 访问更快，且生命周期正好是一个请求——请求结束，Context 被回收，数据自然消失，不存在 goroutine 泄漏问题。

**Q: `c.Abort()` 和 `return` 有什么区别？**

- `c.Abort()` 阻止后续中间件和 handler 执行
- `return` 只是退出当前函数，但 Gin 内部是通过 index 遍历中间件链的，不 return 的话代码会继续往下走到 `c.JSON` 后面的逻辑
- **两者要一起用**：先 `c.Abort()` 阻断链，再 `return` 退出函数

---

## 第三步：角色权限中间件

从 context 取出角色，判断是否有权限访问当前接口。

```go
// RequireRole 接受一个或多个角色，闭包返回 gin.HandlerFunc
func RequireRole(roles ...string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从 context 取出 JWTAuth 注入的角色信息
        role, exists := c.Get("role")
        if !exists {
            c.JSON(http.StatusForbidden, gin.H{"msg": "未认证"})
            c.Abort()
            return
        }

        // 判断当前用户角色是否在允许的角色列表中
        for _, r := range roles {
            if role.(string) == r {
                c.Next() // 匹配成功，放行
                return
            }
        }

        c.JSON(http.StatusForbidden, gin.H{"msg": "权限不足"})
        c.Abort()
    }
}
```

为什么设计成可变参数 `roles ...string`？调用时更灵活：

```go
// 单个角色
auth.Group("/admin", RequireRole("admin"))

// 多个角色都能访问
auth.Group("/manage", RequireRole("admin", "moderator"))
```

如果想做更细粒度的权限（按操作/资源），可以扩展为：

```go
// 按权限码校验
func RequirePermission(codes ...string) gin.HandlerFunc { ... }
// 使用：RequirePermission("user:delete", "user:update")
```

---

## 第四步：路由分组应用

利用 Gin 的 `Group` 功能按层级嵌套中间件。

```go
package main

import (
    "your-project/middleware"
    "github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
    // ====================
    // 公开路由（无需登录）
    // ====================
    r.POST("/login", LoginHandler)
    r.POST("/register", RegisterHandler)

    // ====================
    // 需要登录的路由组
    // ====================
    api := r.Group("/api")
    api.Use(middleware.JWTAuth()) // 方式二：用 Use() 挂载中间件
    {
        api.GET("/profile", GetProfile)
        api.POST("/order", CreateOrder)

        // ====================
        // 需要管理员权限的子组
        // ====================
        admin := api.Group("/admin")
        admin.Use(middleware.RequireRole("admin"))
        {
            admin.GET("/users", ListUsers)
            admin.DELETE("/users/:id", DeleteUser)
        }
    }
}
```

两种挂载中间件的方式等效：

```go
// 方式一：在 Group 时直接传入
auth := r.Group("/api", middleware.JWTAuth())

// 方式二：用 .Use() 追加
api := r.Group("/api")
api.Use(middleware.JWTAuth())
```

最终路由效果：

| 路由 | 中间件链 | 谁可以访问 |
|------|---------|-----------|
| `/login` | 无 | 所有人 |
| `/api/profile` | `JWTAuth` | 已登录用户 |
| `/api/order` | `JWTAuth` | 已登录用户 |
| `/api/admin/users` | `JWTAuth` → `RequireRole("admin")` | 仅管理员 |
| `/api/admin/users/:id` | `JWTAuth` → `RequireRole("admin")` | 仅管理员 |

Gin 的中间件链执行顺序是**洋葱模型**：

```
        请求进入
          │
    ┌─────▼──────┐
    │  JWTAuth   │ ← 前置：解析 token
    │  c.Next()  │ ──────┐
    └────────────┘       │
                     ┌───▼───────┐
                     │ RequireRole│ ← 前置：校验角色
                     │  c.Next()  │ ──────┐
                     └────────────┘       │
                                    ┌─────▼──────┐
                                    │   Handler   │ ← 执行业务逻辑
                                    └─────┬──────┘
                     ┌────────────┐       │
                     │ RequireRole │ ← 后置：无操作
                     └────────────┘       │
    ┌────────────┐       │
    │  JWTAuth   │ ← 后置：无操作
    └────────────┘
          │
          ▼
      响应返回
```

---

## Handler 如何取用户信息

```go
func GetProfile(c *gin.Context) {
    // 直接从 context 取值，不需要查库
    userID, _ := c.Get("user_id")
    username, _ := c.Get("username")
    role, _ := c.Get("role")

    c.JSON(200, gin.H{
        "user_id":  userID,
        "username": username,
        "role":     role,
    })
}
```

为了方便，通常会封装一个取值工具函数，避免到处写类型断言：

```go
// utils/context.go
func GetUserID(c *gin.Context) uint {
    id, _ := c.Get("user_id")
    return id.(uint)
}

func GetRole(c *gin.Context) string {
    role, _ := c.Get("role")
    return role.(string)
}
```

---

## 完整项目结构

```
├── main.go                  # 入口，初始化路由
├── utils/
│   ├── jwt.go               # JWT 工具函数（GenerateToken / ParseToken）
│   └── context.go           # gin.Context 取值工具函数（GetUserID / GetRole）
├── middleware/
│   ├── auth.go              # JWTAuth 认证中间件
│   └── permission.go        # RequireRole 权限中间件
├── handler/
│   ├── auth.go              # 登录、注册 handler
│   ├── user.go              # 用户相关 handler
│   └── admin.go             # 管理员 handler
└── routes/
    └── router.go            # 路由分组定义
```

---

## 补充一：双 Token 方案

access token 过期时间短（15 分钟），refresh token 过期时间长（7 天）。前者暴露风险低，后者用来无感刷新。

```go
// utils/jwt.go 中新增

var (
    AccessTokenExpire  = 15 * time.Minute
    RefreshTokenExpire = 7 * 24 * time.Hour
)

// GenerateAccessToken 短期 token，用于日常请求鉴权
func GenerateAccessToken(userID uint, username, role string) (string, error) {
    claims := Claims{
        UserID:   userID,
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenExpire)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            Issuer:    "gin-jwt-demo",
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(JwtSecret)
}

// GenerateRefreshToken 长期 token，只用于换取新的 access token
func GenerateRefreshToken(userID uint) (string, error) {
    claims := jwt.RegisteredClaims{
        Subject:   fmt.Sprintf("%d", userID),
        ExpiresAt: jwt.NewNumericDate(time.Now().Add(RefreshTokenExpire)),
        IssuedAt:  jwt.NewNumericDate(time.Now()),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(JwtSecret)
}
```

刷新接口 handler：

```go
// POST /refresh
func RefreshTokenHandler(c *gin.Context) {
    var req struct {
        RefreshToken string `json:"refresh_token" binding:"required"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"msg": "参数错误"})
        return
    }

    // 解析 refresh token
    claims, err := utils.ParseToken(req.RefreshToken)
    // 只从 subject 拿 userID，后续查库拿最新角色
    if err != nil {
        c.JSON(401, gin.H{"msg": "refresh token 无效"})
        return
    }

    // 查库拿到用户最新信息（角色可能已变更）
    user := model.FindByID(claims.UserID)
    if user == nil {
        c.JSON(401, gin.H{"msg": "用户不存在"})
        return
    }

    // 生成新的 access token
    newAccessToken, _ := utils.GenerateAccessToken(user.ID, user.Username, user.Role)

    c.JSON(200, gin.H{
        "access_token": newAccessToken,
    })
}
```

流程对比：

```
单 token（简单场景）:
  登录 → 发一个 token → 过期就重新登录

双 token（生产推荐）:
  登录 → 发 access token(15min) + refresh token(7天)
  请求 API → 带 access token
  access token 过期 → 前端用 refresh token 调 /refresh → 拿新的 access token
  refresh token 也过期 → 重新登录
```

---

## 补充二：退出登录（Redis 黑名单）

JWT 是无状态的，签发后无法直接"作废"。解决方案是维护一个 Redis 黑名单。

### 思路

```
用户点击退出
  │
  ▼
后端把该 token 加入 Redis 黑名单，TTL = token 剩余有效时间
  │
  ▼
后续请求到达 JWTAuth 中间件
  │
  ▼
解析 token → 查 Redis 黑名单 → 命中 → 拒绝访问
```

### 代码实现

退出登录 handler：

```go
// POST /logout
func LogoutHandler(c *gin.Context) {
    tokenStr := extractToken(c) // 从 Authorization header 取出 token

    // 解析获取剩余有效时间
    claims, err := utils.ParseToken(tokenStr)
    if err != nil {
        c.JSON(200, gin.H{"msg": "已退出"}) // 即使解析失败也不报错
        return
    }

    // 计算剩余有效时间
    remaining := time.Until(claims.ExpiresAt.Time)
    if remaining <= 0 {
        c.JSON(200, gin.H{"msg": "已退出"})
        return
    }

    // 加入 Redis 黑名单，key: "blacklist:<token>", TTL: 剩余有效时间
    rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
    rdb.Set(c, "blacklist:"+tokenStr, "1", remaining)

    c.JSON(200, gin.H{"msg": "退出成功"})
}

func extractToken(c *gin.Context) string {
    authHeader := c.GetHeader("Authorization")
    parts := strings.SplitN(authHeader, " ", 2)
    if len(parts) == 2 && parts[0] == "Bearer" {
        return parts[1]
    }
    return ""
}
```

在 `JWTAuth` 中间件增加黑名单检查：

```go
func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        // ... 省略格式校验 ...

        tokenStr := parts[1]

        // ★ 新增：查 Redis 黑名单
        rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
        exists, _ := rdb.Exists(c, "blacklist:"+tokenStr).Result()
        if exists > 0 {
            c.JSON(http.StatusUnauthorized, gin.H{"msg": "token 已失效，请重新登录"})
            c.Abort()
            return
        }

        claims, err := utils.ParseToken(tokenStr)
        // ... 后续逻辑不变 ...
    }
}
```

### 为什么 TTL 设为 token 剩余有效时间？

token 过期后 Redis 里的记录也会自动过期删除，不会造成内存堆积。不需要手动清理。

### 生产环境建议

生产环境建议把 Redis 连接封装成单例，不要在中间件里每次 `redis.NewClient()`：

```go
// 在 JWTAuth 中注入已初始化的 redis client
func JWTAuth(rdb *redis.Client) gin.HandlerFunc { ... }
```

---

## 与 Spring Boot 对比

| 维度 | Spring Boot | Go Gin |
|------|------------|--------|
| 核心依赖 | `spring-boot-starter-security` + `jjwt` | `golang-jwt/jwt` |
| 拦截机制 | `OncePerRequestFilter` 过滤器 | `gin.HandlerFunc` 中间件函数 |
| 注入上下文 | `SecurityContextHolder.getContext()` | `c.Set()` / `c.Get()` |
| 权限控制 | `@PreAuthorize("hasRole('ADMIN')")` 注解 | `RequireRole("admin")` 中间件函数 |
| 路由分组 | `antMatchers("/api/**").authenticated()` | `r.Group("/api", JWTAuth())` |
| 配置方式 | 配置类 + 注解驱动 | 纯代码驱动，函数组合 |
| 框架重量 | 重：自动配置 + AOP + 依赖注入 | 轻：一个函数返回 `gin.HandlerFunc` |

**本质是一样的**：拦截器取 Token → 解析 → 注入上下文 → 链式调用。Gin 少了注解和 AOP，用函数组合替代，更轻量更直接。
