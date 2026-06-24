package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		fmt.Printf("[%s] %s %s\n",
			time.Now().Format("2006-01-02 15:04:05"),
			c.Request.Method,
			c.Request.URL.Path,
		)

		c.Next()

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
			c.Abort() // 终止后续处理
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

	auth := r.Group("/api")
	auth.Use(AuthMiddleware())
	{
		auth.GET("/profile", func(c *gin.Context) {
			userID := c.GetInt("user_id")
			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		})
	}

	r.Run(":8082")
}
