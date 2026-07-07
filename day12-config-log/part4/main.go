package main

import (
	"fmt"
	"part4/config"
	"part4/logger"
	"part4/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 1. 加载配置
	if err := config.LoadConfig("config.yaml"); err != nil {
		panic("加载配置失败: " + err.Error())
	}

	// 2. 初始化日志
	if err := logger.InitLogger(
		config.AppConfig.Log.Level,
		config.AppConfig.Log.FilePath,
		config.AppConfig.Log.MaxSize,
		config.AppConfig.Log.MaxBackups,
		config.AppConfig.Log.MaxAge,
	); err != nil {
		panic("初始化日志失败: " + err.Error())
	}
	defer logger.Sync()

	// 3. 设置 Gin 模式
	gin.SetMode(config.AppConfig.Server.Mode)

	// 4. 创建路由
	r := gin.New()
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RecoveryMiddleware())

	// 5. 定义路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 6. 启动服务
	addr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	logger.Logger.Info("服务启动",
		zap.String("addr", addr),
		zap.String("mode", config.AppConfig.Server.Mode),
	)

	if err := r.Run(addr); err != nil {
		logger.Logger.Fatal("服务启动失败", zap.Error(err))
	}
}
