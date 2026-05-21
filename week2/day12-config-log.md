# 第12课：配置管理 + 日志

> 学习时间：2-3小时 | 难度：⭐⭐

## 📋 本课目标

- 掌握 Viper 配置管理
- 学会使用环境变量
- 掌握 Zap 日志库
- 实现日志分级和轮转

## 1. Viper 配置管理

### 安装

```bash
go get github.com/spf13/viper
```

### 配置文件

#### config.yaml

```yaml
server:
  port: 8080
  mode: debug  # debug, release, test

database:
  driver: sqlite
  host: localhost
  port: 5432
  name: skillsbay
  user: postgres
  password: password
  max_idle_conns: 10
  max_open_conns: 100

blockchain:
  ethereum:
    rpc_url: https://eth.llamarpc.com
    chain_id: 1
  ton:
    rpc_url: https://toncenter.com/api/v2
    api_key: your-api-key

jwt:
  secret: your-secret-key-change-in-production
  expires_in: 86400  # 24 hours

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

log:
  level: info  # debug, info, warn, error
  file_path: logs/app.log
  max_size: 100  # MB
  max_backups: 3
  max_age: 28  # days
```

### 配置加载

#### config/config.go

```go
package config

import (
    "fmt"
    "github.com/spf13/viper"
)

type Config struct {
    Server     ServerConfig     `mapstructure:"server"`
    Database   DatabaseConfig   `mapstructure:"database"`
    Blockchain BlockchainConfig `mapstructure:"blockchain"`
    JWT        JWTConfig        `mapstructure:"jwt"`
    Redis      RedisConfig      `mapstructure:"redis"`
    Log        LogConfig        `mapstructure:"log"`
}

type ServerConfig struct {
    Port int    `mapstructure:"port"`
    Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
    Driver       string `mapstructure:"driver"`
    Host         string `mapstructure:"host"`
    Port         int    `mapstructure:"port"`
    Name         string `mapstructure:"name"`
    User         string `mapstructure:"user"`
    Password     string `mapstructure:"password"`
    MaxIdleConns int    `mapstructure:"max_idle_conns"`
    MaxOpenConns int    `mapstructure:"max_open_conns"`
}

type BlockchainConfig struct {
    Ethereum EthereumConfig `mapstructure:"ethereum"`
    TON      TONConfig      `mapstructure:"ton"`
}

type EthereumConfig struct {
    RPCURL  string `mapstructure:"rpc_url"`
    ChainID int    `mapstructure:"chain_id"`
}

type TONConfig struct {
    RPCURL string `mapstructure:"rpc_url"`
    APIKey string `mapstructure:"api_key"`
}

type JWTConfig struct {
    Secret    string `mapstructure:"secret"`
    ExpiresIn int    `mapstructure:"expires_in"`
}

type RedisConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    Password string `mapstructure:"password"`
    DB       int    `mapstructure:"db"`
}

type LogConfig struct {
    Level      string `mapstructure:"level"`
    FilePath   string `mapstructure:"file_path"`
    MaxSize    int    `mapstructure:"max_size"`
    MaxBackups int    `mapstructure:"max_backups"`
    MaxAge     int    `mapstructure:"max_age"`
}

var AppConfig *Config

// LoadConfig 加载配置
func LoadConfig(configPath string) error {
    viper.SetConfigFile(configPath)
    viper.SetConfigType("yaml")
    
    // 读取环境变量
    viper.AutomaticEnv()
    
    // 读取配置文件
    if err := viper.ReadInConfig(); err != nil {
        return fmt.Errorf("读取配置文件失败: %w", err)
    }
    
    // 解析配置
    if err := viper.Unmarshal(&AppConfig); err != nil {
        return fmt.Errorf("解析配置失败: %w", err)
    }
    
    return nil
}

// GetDatabaseDSN 获取数据库连接字符串
func (c *Config) GetDatabaseDSN() string {
    switch c.Database.Driver {
    case "mysql":
        return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
            c.Database.User,
            c.Database.Password,
            c.Database.Host,
            c.Database.Port,
            c.Database.Name,
        )
    case "postgres":
        return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
            c.Database.Host,
            c.Database.User,
            c.Database.Password,
            c.Database.Name,
            c.Database.Port,
        )
    case "sqlite":
        return c.Database.Name + ".db"
    default:
        return ""
    }
}
```

### 使用配置

```go
package main

import (
    "fmt"
    "log"
    "your-project/config"
)

func main() {
    // 加载配置
    if err := config.LoadConfig("config.yaml"); err != nil {
        log.Fatal("加载配置失败:", err)
    }
    
    // 使用配置
    fmt.Printf("服务器端口: %d\n", config.AppConfig.Server.Port)
    fmt.Printf("数据库: %s\n", config.AppConfig.GetDatabaseDSN())
    fmt.Printf("以太坊 RPC: %s\n", config.AppConfig.Blockchain.Ethereum.RPCURL)
}
```

## 2. 环境变量

### .env 文件

```bash
# .env
SERVER_PORT=8080
SERVER_MODE=debug

DB_HOST=localhost
DB_PORT=5432
DB_NAME=skillsbay
DB_USER=postgres
DB_PASSWORD=secret

JWT_SECRET=your-secret-key
JWT_EXPIRES_IN=86400

ETH_RPC_URL=https://eth.llamarpc.com
TON_RPC_URL=https://toncenter.com/api/v2
```

### 加载环境变量

```bash
go get github.com/joho/godotenv
```

```go
package main

import (
    "github.com/joho/godotenv"
    "log"
    "os"
)

func main() {
    // 加载 .env 文件
    if err := godotenv.Load(); err != nil {
        log.Println("未找到 .env 文件")
    }
    
    // 读取环境变量
    port := os.Getenv("SERVER_PORT")
    dbHost := os.Getenv("DB_HOST")
    jwtSecret := os.Getenv("JWT_SECRET")
    
    log.Printf("端口: %s, 数据库: %s\n", port, dbHost)
}
```

## 3. Zap 日志库

### 安装

```bash
go get -u go.uber.org/zap
go get -u gopkg.in/natefinch/lumberjack.v2  # 日志轮转
```

### 日志初始化

#### logger/logger.go

```go
package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
    "gopkg.in/natefinch/lumberjack.v2"
    "os"
)

var Logger *zap.Logger

// InitLogger 初始化日志
func InitLogger(level, filePath string, maxSize, maxBackups, maxAge int) error {
    // 日志级别
    var zapLevel zapcore.Level
    switch level {
    case "debug":
        zapLevel = zapcore.DebugLevel
    case "info":
        zapLevel = zapcore.InfoLevel
    case "warn":
        zapLevel = zapcore.WarnLevel
    case "error":
        zapLevel = zapcore.ErrorLevel
    default:
        zapLevel = zapcore.InfoLevel
    }
    
    // 编码器配置
    encoderConfig := zapcore.EncoderConfig{
        TimeKey:        "time",
        LevelKey:       "level",
        NameKey:        "logger",
        CallerKey:      "caller",
        MessageKey:     "msg",
        StacktraceKey:  "stacktrace",
        LineEnding:     zapcore.DefaultLineEnding,
        EncodeLevel:    zapcore.CapitalColorLevelEncoder,  // 彩色输出
        EncodeTime:     zapcore.ISO8601TimeEncoder,
        EncodeDuration: zapcore.SecondsDurationEncoder,
        EncodeCaller:   zapcore.ShortCallerEncoder,
    }
    
    // 文件输出
    fileWriter := &lumberjack.Logger{
        Filename:   filePath,
        MaxSize:    maxSize,    // MB
        MaxBackups: maxBackups,
        MaxAge:     maxAge,     // days
        Compress:   true,
    }
    
    // 控制台输出
    consoleWriter := zapcore.AddSync(os.Stdout)
    
    // 创建 Core
    core := zapcore.NewTee(
        zapcore.NewCore(
            zapcore.NewJSONEncoder(encoderConfig),
            zapcore.AddSync(fileWriter),
            zapLevel,
        ),
        zapcore.NewCore(
            zapcore.NewConsoleEncoder(encoderConfig),
            consoleWriter,
            zapLevel,
        ),
    )
    
    // 创建 Logger
    Logger = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
    
    return nil
}

// Sync 刷新日志缓冲
func Sync() {
    if Logger != nil {
        Logger.Sync()
    }
}
```

### 使用日志

```go
package main

import (
    "your-project/logger"
    "go.uber.org/zap"
)

func main() {
    // 初始化日志
    logger.InitLogger("info", "logs/app.log", 100, 3, 28)
    defer logger.Sync()
    
    // 使用日志
    logger.Logger.Info("服务启动",
        zap.String("port", "8080"),
        zap.String("mode", "debug"),
    )
    
    logger.Logger.Debug("调试信息", zap.Int("user_id", 123))
    
    logger.Logger.Warn("警告信息", zap.String("reason", "余额不足"))
    
    logger.Logger.Error("错误信息",
        zap.String("error", "数据库连接失败"),
        zap.Int("retry", 3),
    )
    
    // 结构化日志
    logger.Logger.Info("用户登录",
        zap.String("username", "linshen"),
        zap.String("ip", "192.168.1.1"),
        zap.Int64("timestamp", 1704067200),
    )
}
```

### Gin 集成日志中间件

```go
package middleware

import (
    "github.com/gin-gonic/gin"
    "go.uber.org/zap"
    "time"
    "your-project/logger"
)

// LoggerMiddleware Gin 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path
        query := c.Request.URL.RawQuery
        
        c.Next()
        
        latency := time.Since(start)
        
        logger.Logger.Info("HTTP请求",
            zap.String("method", c.Request.Method),
            zap.String("path", path),
            zap.String("query", query),
            zap.Int("status", c.Writer.Status()),
            zap.Duration("latency", latency),
            zap.String("ip", c.ClientIP()),
            zap.String("user_agent", c.Request.UserAgent()),
        )
    }
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                logger.Logger.Error("Panic恢复",
                    zap.Any("error", err),
                    zap.String("path", c.Request.URL.Path),
                )
                c.JSON(500, gin.H{"error": "服务器内部错误"})
            }
        }()
        c.Next()
    }
}
```

## 4. 完整示例

### main.go

```go
package main

import (
    "fmt"
    "github.com/gin-gonic/gin"
    "your-project/config"
    "your-project/logger"
    "your-project/middleware"
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
```

## 📝 作业

### 作业1：多环境配置

创建 `homework/day12/multi-env`：

```go
// TODO: 实现多环境配置
// 1. config.dev.yaml - 开发环境
// 2. config.prod.yaml - 生产环境
// 3. config.test.yaml - 测试环境
// 4. 根据环境变量自动加载对应配置
```

### 作业2：日志分析工具

```go
// TODO: 实现日志分析工具
// 1. 统计不同级别的日志数量
// 2. 统计最慢的 10 个请求
// 3. 统计错误最多的接口
// 4. 生成日志报告
```

### 作业3：配置热更新

```go
// TODO: 实现配置热更新
// 1. 监听配置文件变化
// 2. 自动重新加载配置
// 3. 通知相关模块更新
```

## 🎯 检查点

- ✅ 能够使用 Viper 管理配置
- ✅ 掌握环境变量的使用
- ✅ 能够使用 Zap 记录日志
- ✅ 实现日志轮转和分级
- ✅ 集成日志到 Gin 项目

## ⏭️ 下一课

[第13-14课：实战项目 - SkillsBay API](./day13-14-project.md)
