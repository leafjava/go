package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

//### 内置中间件
//func main() {
//	r := gin.Default()
//
//	r2 := gin.New()
//	r2.Use(gin.Logger())
//	r2.Use(gin.Recovery())
//
//	r.GET("/", func(c *gin.Context) {
//		c.JSON(http.StatusOK, gin.H{
//			"message": "Hello",
//		})
//	})
//
//	r.Run(":8080")
//}

//### 自定义中间件

// 日志中间件
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
		fmt.Printf("耗时:%v\n", latency)
	}
}

// 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的令牌",
			})
			c.Abort()
			return
		}

		if token != "Bearer valid-token" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "无效的令牌",
			})
			c.Abort()
			return
		}

		c.Set("user_id", 123)
		c.Next()
	}
}

func main() {
	r := gin.New()

	r.Use(LoggerMiddleware())
	r.Use(gin.Recovery())

	r.GET("/public", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "公开接口",
		})
	})

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
