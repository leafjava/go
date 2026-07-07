package main

import (
	"fmt"
	"part1/middleware"
	"skillsbay-api/database"
	"skillsbay-api/handlers"

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

	// 3. 初始化数据库
	if err := database.InitDB(); err != nil {
		panic("初始化数据库失败: " + err.Error())
	}

	// 4. 创建路由
	gin.SetMode(config.AppConfig.Server.Mode)
	r := gin.New()

	// 5. 全局中间件
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.CORSMiddleware())

	// 6. 初始化处理器
	authHandler := &handlers.AuthHandler{}
	walletHandler := &handlers.WalletHandler{}

	// 7. 公开路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	r.POST("/api/v1/auth/register", authHandler.Register)
	r.POST("/api/v1/auth/login", authHandler.Login)

	// 8. 需要认证的路由
	auth := r.Group("/api/v1")
	auth.Use(middleware.JWTAuth())
	{
		// 钱包路由
		wallets := auth.Group("/wallets")
		{
			wallets.POST("", walletHandler.CreateWallet)
			wallets.GET("", walletHandler.GetWallets)
			wallets.GET("/:id", walletHandler.GetWallet)
		}
	}

	// 9. 启动服务
	addr := fmt.Sprintf(":%d", config.AppConfig.Server.Port)
	logger.Logger.Info("服务启动", zap.String("addr", addr))

	if err := r.Run(addr); err != nil {
		logger.Logger.Fatal("服务启动失败", zap.Error(err))
	}
}
