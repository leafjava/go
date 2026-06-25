package main

import (
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 1. 请求日志中间件 — 记录请求方法、路径、耗时
// ============================================================
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()        // ① 请求进来时记下时间
		method := c.Request.Method // ② 请求方法 GET/POST/...
		path := c.Request.URL.Path // ③ 请求路径

		c.Next() // ④ 执行后续中间件 + 处理函数

		elapsed := time.Since(start) // ⑤ 处理完算耗时
		log.Printf("[日志] %s %s — 耗时: %v", method, path, elapsed)
	}
}

// ============================================================
// 2. 认证中间件 — JWT Token 验证
// ============================================================
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// 检查是否有 Authorization 头
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少认证令牌，请先登录",
			})
			c.Abort()
			return
		}

		// 检查格式："Bearer xxxxx"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "令牌格式错误",
			})
			c.Abort()
			return
		}

		token := parts[1]

		// 验证 token（实际项目用 jwt-go 库解析，这里做简单校验）
		if token == "" || len(token) < 10 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "令牌无效或已过期",
			})
			c.Abort()
			return
		}

		// 模拟从 token 解析出用户信息，存入上下文供后续使用
		c.Set("user_id", "user_123")
		c.Set("username", "张三")

		c.Next()
	}
}

// ============================================================
// 3. 权限中间件 — 角色检查
// ============================================================
func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 优先从上下文取角色，取不到就从请求头取
		role, exists := c.Get("role")
		if !exists {
			role = c.GetHeader("X-User-Role")
		}

		if role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    403,
				"message": "权限不足，需要 " + requiredRole + " 角色",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ============================================================
// 4. 限流中间件 — 滑动窗口 IP 限流
// ============================================================
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

	// 清理过期记录（滑动窗口核心）
	if times, exists := rl.requests[key]; exists {
		var valid []time.Time
		for _, t := range times {
			if now.Sub(t) < rl.window {
				valid = append(valid, t)
			}
		}
		rl.requests[key] = valid
	}

	// 判断是否超限
	if len(rl.requests[key]) >= rl.limit {
		return false
	}

	// 记录本次请求
	rl.requests[key] = append(rl.requests[key], now)
	return true
}

func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)
	return func(c *gin.Context) {
		key := c.ClientIP()
		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后再试",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ============================================================
// 5. CORS 跨域中间件
// ============================================================
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Authorization, Accept, X-Requested-With, X-User-Role")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")

		// 预检请求直接返回 204，不走到后续处理函数
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// ============================================================
// main — 组装中间件 + 注册路由
// ============================================================
func main() {
	r := gin.Default()

	// ---------- 全局中间件（每个请求都会经过）----------
	r.Use(LoggerMiddleware())                   // ① 日志
	r.Use(CORSMiddleware())                     // ② 跨域
	r.Use(RateLimitMiddleware(20, time.Minute)) // ③ 限流（每IP每分钟20次）

	// ---------- 公开路由（无需登录）----------
	r.GET("/api/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "公开数据，无需登录",
		})
	})

	// ---------- 需要认证的路由组 ----------
	auth := r.Group("/api")
	auth.Use(AuthMiddleware()) // ④ 认证
	{
		// 普通用户接口
		auth.GET("/profile", func(c *gin.Context) {
			username, _ := c.Get("username")
			c.JSON(http.StatusOK, gin.H{
				"code":    200,
				"message": "个人资料",
				"user":    username,
			})
		})

		// 管理员接口（认证 + 权限）
		admin := auth.Group("/admin")
		admin.Use(RoleMiddleware("admin")) // ⑤ 权限：必须是 admin
		{
			admin.GET("/dashboard", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"code":    200,
					"message": "管理员面板",
					"data":    gin.H{"total_users": 1024, "total_orders": 5678},
				})
			})

			admin.DELETE("/users/:id", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"code":    200,
					"message": "用户 " + c.Param("id") + " 已删除",
				})
			})
		}
	}

	r.Run(":8081")
}
