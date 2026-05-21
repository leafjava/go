// HashKey Web3 全栈面试题：Gin 如何做 JWT 认证 + 中间件？
// 对应课程：第11课 (week2/day11-jwt.md)

package main

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ============================================
// 1. JWT 工具函数
// ============================================

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
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "hashkey-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ParseToken 解析 JWT Token
func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ============================================
// 2. JWT 认证中间件（核心）
// ============================================

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
		claims, err := ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  "无效的 token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// 4. 将用户信息存入上下文（关键步骤）
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RequireRole 角色权限中间件
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code": 403,
				"msg":  "无权限访问",
			})
			c.Abort()
			return
		}

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

// ============================================
// 3. 用户模型
// ============================================

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Password string `gorm:"not null" json:"-"`
	Email    string `gorm:"uniqueIndex" json:"email"`
	Role     string `gorm:"default:'user'" json:"role"`
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

// ============================================
// 4. 主程序 - 路由配置
// ============================================

var db *gorm.DB

func main() {
	// 初始化数据库
	var err error
	db, err = gorm.Open(sqlite.Open("hashkey_jwt_demo.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	db.AutoMigrate(&User{})

	r := gin.Default()

	// 公开路由（不需要认证）
	r.POST("/register", register)
	r.POST("/login", login)

	// 需要认证的路由（应用 JWT 中间件）
	auth := r.Group("/api")
	auth.Use(JWTAuth()) // 关键：应用 JWT 中间件
	{
		auth.GET("/profile", getProfile)
		auth.PUT("/profile", updateProfile)

		// 需要管理员权限的路由（叠加角色中间件）
		admin := auth.Group("/admin")
		admin.Use(RequireRole("admin")) // 关键：叠加权限中间件
		{
			admin.GET("/users", listUsers)
			admin.DELETE("/users/:id", deleteUser)
		}
	}

	r.Run(":8080")
}

// ============================================
// 5. Handler 函数
// ============================================

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

	user := User{
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

// 登录（生成 Token）
func login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user User
	if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	// 生成 JWT Token（关键步骤）
	token, err := GenerateToken(user.ID, user.Username, user.Role)
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

// 获取个人信息（需要认证）
func getProfile(c *gin.Context) {
	// 从上下文获取用户信息（由 JWT 中间件设置）
	userID := c.GetUint("user_id")

	var user User
	if err := db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": user,
	})
}

// 更新个人信息（需要认证）
func updateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Email string `json:"email" binding:"email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := db.Model(&User{}).Where("id = ?", userID).Update("email", req.Email).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "更新成功",
	})
}

// 管理员：获取所有用户（需要 admin 权限）
func listUsers(c *gin.Context) {
	var users []User
	db.Find(&users)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": users,
	})
}

// 管理员：删除用户（需要 admin 权限）
func deleteUser(c *gin.Context) {
	id := c.Param("id")
	db.Delete(&User{}, id)

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "删除成功",
	})
}

/*
============================================
面试标准回答（背诵版）
============================================

问题：Gin 如何做 JWT 认证 + 中间件？

答：JWT 认证在 Gin 中主要通过自定义中间件实现，核心流程分为三步：

1. Token 生成（登录时）
   - 使用 golang-jwt/jwt 库创建 Claims，包含用户 ID、角色等信息
   - 设置过期时间（通常 24 小时）
   - 使用 HMAC-SHA256 算法签名，生成 token 返回给客户端

2. 中间件验证
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

3. 路由应用
   auth := r.Group("/api")
   auth.Use(middleware.JWTAuth())
   {
       auth.GET("/profile", getProfile)
   }

关键点：
- Token 存储在 Authorization: Bearer {token} Header 中
- 中间件通过 c.Set() 传递用户信息给后续 Handler
- 使用 c.Abort() 阻止未授权请求继续执行
- 生产环境密钥应从环境变量读取

进阶：权限控制
可以叠加角色中间件实现 RBAC：
admin := auth.Group("/admin")
admin.Use(middleware.RequireRole("admin"))

安全建议：
1. Token 过期时间不宜过长
2. 敏感操作考虑 Refresh Token 机制
3. 使用 HTTPS 传输
4. 密钥定期轮换
5. 考虑 Token 黑名单（Redis）处理登出

============================================
测试命令
============================================

# 1. 注册用户
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"123456","email":"test@example.com"}'

# 2. 登录获取 Token
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"123456"}'

# 3. 使用 Token 访问受保护路由
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# 4. 访问管理员路由（需要 admin 角色）
curl -X GET http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer ADMIN_TOKEN_HERE"

============================================
常见面试追问
============================================

Q1: JWT 和 Session 的区别？
A: JWT 无状态，服务端不存储，适合分布式系统，但无法主动失效；
   Session 有状态，服务端存储，可主动失效，但需要共享存储（Redis）

Q2: 如何实现 Token 刷新？
A: 使用双 Token 机制：
   - Access Token: 短期（15分钟），用于 API 访问
   - Refresh Token: 长期（7天），用于刷新 Access Token

Q3: JWT 被盗用怎么办？
A: 1. 使用 HTTPS 防止中间人攻击
   2. 设置短过期时间
   3. 实现 Token 黑名单（Redis）
   4. 记录设备指纹，异常登录提醒

Q4: 如何处理 Token 过期？
A: 前端拦截 401 响应，自动调用刷新接口重试原请求

*/
