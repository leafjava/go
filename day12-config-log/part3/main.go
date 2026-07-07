package main

import (
	"part2/logger"
	"part2/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	logger.InitLogger("info", "logs/app.log", 100, 3, 28)
	defer logger.Sync()

	// 创建 Gin 引擎
	r := gin.New()

	// 挂载自定义中间件
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RecoveryMiddleware())

	// 测试路由
	r.GET("/ping", func(c *gin.Context) {
		logger.Logger.Info("处理请求",
			zap.String("path", c.Request.URL.Path),
		)
		c.JSON(200, gin.H{"message": "pong"})
	})

	// 测试 panic 恢复
	r.GET("/panic", func(c *gin.Context) {
		panic("模拟 panic")
	})

	// 启动服务
	logger.Logger.Info("服务启动",
		zap.String("port", "8083"),
		zap.String("mode", "debug"),
	)
	if err := r.Run(":8083"); err != nil {
		logger.Logger.Fatal("服务启动失败", zap.Error(err))
	}
}
