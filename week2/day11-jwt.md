# 第11课：JWT 认证 + 权限控制

## 📚 学习目标
- 理解 JWT 工作原理
- 在 Gin 中实现 JWT 认证中间件
- 掌握权限控制最佳实践

## 🎯 核心知识点

### 1. JWT 基础概念

JWT (JSON Web Token) 由三部分组成：
```
Header.Payload.Signature
```

- **Header**: 令牌类型和加密算法
- **Payload**: 用户信息（Claims）
- **Signature**: 签名，防止篡改

### 2. Gin JWT 认证完整实现

#### 安装依赖
```bash
go get -u github.com/golang-jwt/jwt/v5
go get -u github.com/gin-gonic/gin
```

#### 项目结构
```
project/
├── middleware/
│   └── auth.go          # JWT 中间件
├── models/
│   └── user.go          # 用户模型
├── utils/
│   └── jwt.go           # JWT 工具函数
└── main.go
```

## 💻 完整示例代码

### 1. JWT 工具函数 (utils/jwt.go)

```go
package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWT 密钥（生产环境应从环境变量读取）
var jwtSecret = []byte("your-secret-key-change-in-production")

// Claims 自定义声明
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT Token
func GenerateToken(userID uint, username, role string) (string, error) {
	// 设置过期时间（24小时）
	expirationTime := time.Now().Add(24 * time.Hour)

	// 创建声明
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "your-app-name",
		},
	}

	// 使用 HS256 算法生成 token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ParseToken 解析 JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	// 解析 token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名算法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	// 验证 token 并提取 claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken 刷新 Token（可选）
func RefreshToken(tokenString string) (string, error) {
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 如果 token 还有超过 1 小时才过期，不刷新
	if time.Until(claims.ExpiresAt.Time) > time.Hour {
		return tokenString, nil
	}

	// 生成新 token
	return GenerateToken(claims.UserID, claims.Username, claims.Role)
}
```

### 2. JWT 认证中间件 (middleware/auth.go)

```go
package middleware

import (
	"net/http"
	"strings"

	"your-project/utils"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT 认证中间件
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 从 Header 获取 token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "请求头中缺少 Authorization",
			})
			c.Abort()
			return
		}

		// 2. 验证 token 格式（Bearer token）
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "Authorization 格式错误，应为 Bearer {token}",
			})
			c.Abort()
			return
		}

		// 3. 解析 token
		claims, err := utils.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "无效的 token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 4. 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole 角色权限中间件
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户角色
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code": 403,
				"msg":  "无权限访问",
			})
			c.Abort()
			return
		}

		// 检查角色
		roleStr := userRole.(string)
		hasPermission := false
		for _, role := range roles {
			if roleStr == role {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"code": 403,
				"msg":  "权限不足",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
```

### 3. 用户模型 (models/user.go)

```go
package models

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Password string `gorm:"not null" json:"-"` // json:"-" 不返回密码
	Email    string `gorm:"uniqueIndex" json:"email"`
	Role     string `gorm:"default:'user'" json:"role"` // user, admin
}

// SetPassword 加密密码
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return nil
}

// CheckPassword 验证密码
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
	return err == nil
}
```

### 4. 主程序 (main.go)

```go
package main

import (
	"net/http"

	"your-project/middleware"
	"your-project/models"
	"your-project/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func main() {
	// 初始化数据库
	var err error
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	// 自动迁移
	db.AutoMigrate(&models.User{})

	// 创建路由
	r := gin.Default()

	// 公开路由
	r.POST("/register", register)
	r.POST("/login", login)

	// 需要认证的路由
	auth := r.Group("/api")
	auth.Use(middleware.JWTAuth())
	{
		auth.GET("/profile", getProfile)
		auth.PUT("/profile", updateProfile)

		// 需要管理员权限的路由
		admin := auth.Group("/admin")
		admin.Use(middleware.RequireRole("admin"))
		{
			admin.GET("/users", listUsers)
			admin.DELETE("/users/:id", deleteUser)
		}
	}

	r.Run(":8080")
}

// 注册
func register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Email    string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建用户
	user := models.User{
		Username: req.Username,
		Email:    req.Email,
		Role:     "user",
	}
	user.SetPassword(req.Password)

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "用户创建失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "注册成功",
		"data": user,
	})
}

// 登录
func login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 查询用户
	var user models.User
	if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 验证密码
	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 生成 token
	token, err := utils.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token 生成失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "登录成功",
		"data": gin.H{
			"token": token,
			"user":  user,
		},
	})
}

// 获取个人信息
func getProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": user,
	})
}

// 更新个人信息
func updateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Email string `json:"email" binding:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&models.User{}).Where("id = ?", userID).Update("email", req.Email).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})
}

// 管理员：获取所有用户
func listUsers(c *gin.Context) {
	var users []models.User
	db.Find(&users)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": users,
	})
}

// 管理员：删除用户
func deleteUser(c *gin.Context) {
	id := c.Param("id")
	db.Delete(&models.User{}, id)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}
```

## 🧪 测试示例

### 1. 注册用户
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456",
    "email": "test@example.com"
  }'
```

### 2. 登录获取 Token
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "123456"
  }'
```

### 3. 使用 Token 访问受保护路由
```bash
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## 📝 面试标准回答

### 问题：Gin 如何做 JWT 认证 + 中间件？

**标准回答：**

JWT 认证在 Gin 中主要通过自定义中间件实现，核心流程分为三步：

**1. Token 生成（登录时）**
- 使用 `golang-jwt/jwt` 库创建 Claims，包含用户 ID、角色等信息
- 设置过期时间（通常 24 小时）
- 使用 HMAC-SHA256 算法签名，生成 token 返回给客户端

**2. 中间件验证**
```go
func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 从 Header 提取 Bearer token
        authHeader := c.GetHeader("Authorization")
        
        // 解析并验证 token
        claims, err := ParseToken(token)
        
        // 将用户信息存入 Context
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

**3. 路由应用**
```go
auth := r.Group("/api")
auth.Use(middleware.JWTAuth())
{
    auth.GET("/profile", getProfile)
}
```

**关键点：**
- Token 存储在 `Authorization: Bearer {token}` Header 中
- 中间件通过 `c.Set()` 传递用户信息给后续 Handler
- 使用 `c.Abort()` 阻止未授权请求继续执行
- 生产环境密钥应从环境变量读取，不要硬编码

**进阶：权限控制**
可以叠加角色中间件实现 RBAC：
```go
admin := auth.Group("/admin")
admin.Use(middleware.RequireRole("admin"))
```

**安全建议：**
1. Token 过期时间不宜过长
2. 敏感操作考虑 Refresh Token 机制
3. 使用 HTTPS 传输
4. 密钥定期轮换
5. 考虑 Token 黑名单（Redis）处理登出

## 🎯 常见面试追问

### Q1: JWT 和 Session 的区别？
**答：**
- **JWT**: 无状态，服务端不存储，适合分布式系统，但无法主动失效
- **Session**: 有状态，服务端存储，可主动失效，但需要共享存储（Redis）

### Q2: 如何实现 Token 刷新？
**答：**
使用双 Token 机制：
- **Access Token**: 短期（15分钟），用于 API 访问
- **Refresh Token**: 长期（7天），用于刷新 Access Token
- Refresh Token 存储在 HttpOnly Cookie 中更安全

### Q3: JWT 被盗用怎么办？
**答：**
1. 使用 HTTPS 防止中间人攻击
2. 设置短过期时间
3. 实现 Token 黑名单（Redis）
4. 记录设备指纹，异常登录提醒
5. 敏感操作二次验证

### Q4: 如何处理 Token 过期？
**答：**
前端拦截 401 响应，自动调用刷新接口：
```javascript
axios.interceptors.response.use(
  response => response,
  async error => {
    if (error.response.status === 401) {
      const newToken = await refreshToken()
      // 重试原请求
    }
  }
)
```

## 🔥 生产级优化

### 1. Token 黑名单（登出功能）
```go
// 使用 Redis 存储已登出的 token
func Logout(c *gin.Context) {
    token := c.GetHeader("Authorization")
    // 将 token 加入黑名单，过期时间与 token 一致
    redis.Set("blacklist:"+token, "1", 24*time.Hour)
}

// 中间件检查黑名单
func JWTAuth() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        if redis.Exists("blacklist:" + token) {
            c.JSON(401, gin.H{"error": "token 已失效"})
            c.Abort()
            return
        }
        // ... 其他验证
    }
}
```

### 2. 多设备管理
```go
type Claims struct {
    UserID   uint
    DeviceID string // 设备唯一标识
    jwt.RegisteredClaims
}

// 限制同时登录设备数
func LimitDevices(userID uint, deviceID string) error {
    key := fmt.Sprintf("user:%d:devices", userID)
    devices := redis.SMembers(key)
    
    if len(devices) >= 3 { // 最多3个设备
        // 踢出最早的设备
        oldestDevice := devices[0]
        redis.SRem(key, oldestDevice)
    }
    
    redis.SAdd(key, deviceID)
    return nil
}
```

### 3. 性能优化
```go
// 使用 sync.Pool 复用 Claims 对象
var claimsPool = sync.Pool{
    New: func() interface{} {
        return &Claims{}
    },
}

func ParseToken(tokenString string) (*Claims, error) {
    claims := claimsPool.Get().(*Claims)
    defer claimsPool.Put(claims)
    // ... 解析逻辑
}
```

## 📚 作业

1. 实现完整的 JWT 认证系统（注册、登录、权限控制）
2. 添加 Refresh Token 机制
3. 使用 Redis 实现 Token 黑名单
4. 编写单元测试覆盖所有中间件逻辑

## 🔗 参考资料

- [golang-jwt/jwt 官方文档](https://github.com/golang-jwt/jwt)
- [JWT 官网](https://jwt.io/)
- [Gin 中间件最佳实践](https://gin-gonic.com/docs/examples/custom-middleware/)
