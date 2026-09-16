# 第31课（Week5 Day 2）：MySQL + Redis + 工程习惯

> 学习时间：6~8 小时 | 难度：⭐⭐⭐
> **日期**：9 月 13 日
> **目标**：昨天的接口能落库、能查历史、限流防刷、重复请求不把库打爆

## 🎯 交卷标准

- [ ] `conversations` + `messages` 两张表建好，能正常落库
- [ ] 同一个 `conversation_id` 第二次发请求时，能查到历史消息拼进 prompt
- [ ] Redis 限流：同一 user_id 1 秒最多 5 次请求，超过返回 `CodeRateLimited`
- [ ] Redis 缓存最近会话（最近 10 条），减少数据库查询
- [ ] 接口幂等：相同 `idempotency_key` 重复请求不重复扣费/不重复入库
- [ ] EXPLAIN 看一次慢查询，知道索引怎么建
- [ ] 整个对话流程能跑通：开始 → 多轮 → 查历史 → 结束

---

## 📋 本课目标

- 巩固 GORM 基础（[week2/day10](../week2/day10-gorm.md)）
- 巩固 go-redis 基础（[week3/day19](../week3/day19-redis-mq.md)）
- 学会按 user_id 限流（Redis 令牌桶/滑动窗口）
- 学会幂等设计（生产环境必备）
- 学会用 EXPLAIN 排查慢查询

## 1. 数据库设计

### 1.1 表结构（migrations/001_init.sql）

```sql
-- 会话表
CREATE TABLE conversations (
    id            BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    conversation_id VARCHAR(64) NOT NULL UNIQUE,    -- 业务 UUID，对外暴露
    user_id        VARCHAR(64) NOT NULL,
    title          VARCHAR(255) DEFAULT '',
    model          VARCHAR(64) DEFAULT '',
    created_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at     DATETIME NULL,
    INDEX idx_user_id_updated (user_id, updated_at),
    INDEX idx_conversation_id (conversation_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 消息表
CREATE TABLE messages (
    id              BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    conversation_id VARCHAR(64) NOT NULL,
    user_id         VARCHAR(64) NOT NULL,
    role            VARCHAR(16) NOT NULL,           -- system / user / assistant / tool
    content         MEDIUMTEXT NOT NULL,
    tokens_in       INT DEFAULT 0,                  -- 入 token
    tokens_out      INT DEFAULT 0,                  -- 出 token
    model           VARCHAR(64) DEFAULT '',
    trace_id        VARCHAR(64) DEFAULT '',         -- 链路追踪
    created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_conv_created (conversation_id, created_at),
    INDEX idx_user_id (user_id),
    INDEX idx_trace_id (trace_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 幂等记录表
CREATE TABLE idempotency_keys (
    id          BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    idem_key    VARCHAR(128) NOT NULL UNIQUE,
    user_id     VARCHAR(64) NOT NULL,
    response    MEDIUMTEXT,                  -- 缓存的响应体
    status      VARCHAR(16) NOT NULL,        -- pending / done / failed
    expires_at  DATETIME NOT NULL,
    created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_expires (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

> 💡 **索引说明**：
> - `idx_user_id_updated`：按用户查会话列表
> - `idx_conv_created`：按会话查消息历史
> - **永远给外键 / 频繁查询字段建索引**，但**不要给超长字段**（如 `content`）建索引

### 1.2 索引自检（EXPLAIN）

```sql
-- 模拟一次查询
EXPLAIN SELECT * FROM messages
WHERE conversation_id = 'abc123'
ORDER BY created_at ASC;

-- 看 type 列是不是 ref/range（走索引），如果是 ALL（全表扫）就完蛋了
-- 看 key 列用了哪个索引
-- 看 rows 列预估扫多少行
```

## 2. 依赖安装 + 连接初始化

```bash
go get gorm.io/gorm@latest
go get gorm.io/driver/mysql@latest
go get github.com/redis/go-redis/v9@latest
go get github.com/golang-migrate/migrate/v4 \
  -tool github.com/golang-migrate/migrate/v4/cmd/migrate@latest  # 可选，CLI 工具
```

### config.go 增加配置

```go
type MySQLConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type Config struct {
	Server   ServerConfig
	LLM      LLMConfig
	MySQL    MySQLConfig
	Redis    RedisConfig
	LogLevel string
}
```

```go
// .env
MYSQL_DSN=root:root123@tcp(127.0.0.1:3306)/chat_service?charset=utf8mb4&parseTime=True&loc=Local
MYSQL_MAX_OPEN_CONNS=50
MYSQL_MAX_IDLE_CONNS=10
MYSQL_CONN_MAX_LIFETIME=30m

REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
REDIS_DB=0
```

### infra/db/mysql.go

```go
package infra

import (
	"context"
	"fmt"
	"time"

	"chat-service/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewMySQL(cfg *config.MySQLConfig) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// 启动时 ping 一下
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("ping mysql: %w", err)
	}
	return db, nil
}
```

### infra/cache/redis.go

```go
package infra

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"chat-service/internal/config"
)

func NewRedis(cfg *config.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		PoolSize:     50,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return rdb, nil
}
```

## 3. 数据模型（internal/model/chat.go）

```go
package model

import "time"

// Conversation 会话
type Conversation struct {
	ID             uint64    `gorm:"primaryKey" json:"-"`
	ConversationID string    `gorm:"column:conversation_id;uniqueIndex;size:64" json:"conversation_id"`
	UserID         string    `gorm:"column:user_id;size:64;index" json:"user_id"`
	Title          string    `gorm:"column:title;size:255" json:"title"`
	Model          string    `gorm:"column:model;size:64" json:"model"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (Conversation) TableName() string { return "conversations" }

// Message 消息
type Message struct {
	ID             uint64    `gorm:"primaryKey" json:"-"`
	ConversationID  string    `gorm:"column:conversation_id;size:64;index:idx_conv_created,priority:1" json:"conversation_id"`
	UserID         string    `gorm:"column:user_id;size:64;index" json:"user_id"`
	Role           string    `gorm:"column:role;size:16" json:"role"`
	Content        string    `gorm:"column:content;type:mediumtext" json:"content"`
	TokensIn       int       `gorm:"column:tokens_in;default:0" json:"tokens_in"`
	TokensOut      int       `gorm:"column:tokens_out;default:0" json:"tokens_out"`
	Model          string    `gorm:"column:model;size:64" json:"model"`
	TraceID        string    `gorm:"column:trace_id;size:64;index" json:"trace_id"`
	CreatedAt      time.Time `json:"created_at"`
}

func (Message) TableName() string { return "messages" }

// IdempotencyKey 幂等键
type IdempotencyKey struct {
	ID        uint64    `gorm:"primaryKey" json:"-"`
	IdemKey   string    `gorm:"column:idem_key;uniqueIndex;size:128" json:"idem_key"`
	UserID    string    `gorm:"column:user_id;size:64" json:"user_id"`
	Response  string    `gorm:"column:response;type:mediumtext" json:"response"`
	Status    string    `gorm:"column:status;size:16" json:"status"`
	ExpiresAt time.Time `gorm:"column:expires_at;index" json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

func (IdempotencyKey) TableName() string { return "idempotency_keys" }
```

## 4. Repository 层（internal/repository）

### repository/conversation.go

```go
package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"chat-service/internal/model"
)

type ConversationRepo struct{ db *gorm.DB }

func NewConversationRepo(db *gorm.DB) *ConversationRepo {
	return &ConversationRepo{db: db}
}

// GetOrCreate 获取或创建会话
func (r *ConversationRepo) GetOrCreate(ctx context.Context, convID, userID, model string) (*model.Conversation, error) {
	if convID != "" {
		var conv model.Conversation
		err := r.db.WithContext(ctx).
			Where("conversation_id = ? AND user_id = ?", convID, userID).
			First(&conv).Error
		if err == nil {
			return &conv, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	// 新建
	if convID == "" {
		convID = newConvID()
	}
	conv := &model.Conversation{
		ConversationID: convID,
		UserID:         userID,
		Model:          model,
		Title:          "新会话",
	}
	if err := r.db.WithContext(ctx).Create(conv).Error; err != nil {
		return nil, err
	}
	return conv, nil
}

// ListByUser 查用户的会话列表（按更新时间倒序）
func (r *ConversationRepo) ListByUser(ctx context.Context, userID string, limit int) ([]model.Conversation, error) {
	var convs []model.Conversation
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("updated_at DESC").
		Limit(limit).
		Find(&convs).Error
	return convs, err
}

// Touch 更新会话时间（消息变化时调用）
func (r *ConversationRepo) Touch(ctx context.Context, convID string) error {
	return r.db.WithContext(ctx).
		Model(&model.Conversation{}).
		Where("conversation_id = ?", convID).
		Update("updated_at", time.Now()).Error
}

func newConvID() string {
	// 简单实现，进阶用 uuid
	return "conv_" + time.Now().Format("20060102150405") + "_" + randStr(6)
}

func randStr(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
```

### repository/message.go

```go
package repository

import (
	"context"

	"gorm.io/gorm"

	"chat-service/internal/model"
)

type MessageRepo struct{ db *gorm.DB }

func NewMessageRepo(db *gorm.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

// Insert 插入一条消息
func (r *MessageRepo) Insert(ctx context.Context, msg *model.Message) error {
	return r.db.WithContext(ctx).Create(msg).Error
}

// ListByConv 查会话的历史消息
func (r *MessageRepo) ListByConv(ctx context.Context, convID string, limit int) ([]model.Message, error) {
	var msgs []model.Message
	err := r.db.WithContext(ctx).
		Where("conversation_id = ?", convID).
		Order("created_at ASC").
		Limit(limit).
		Find(&msgs).Error
	return msgs, err
}
```

### repository/idempotency.go

```go
package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"chat-service/internal/model"
)

type IdempotencyRepo struct{ db *gorm.DB }

func NewIdempotencyRepo(db *gorm.DB) *IdempotencyRepo {
	return &IdempotencyRepo{db: db}
}

// ErrDuplicate 幂等键已存在
var ErrDuplicate = errors.New("duplicate idempotency key")

// Reserve 尝试占用一个幂等键（INSERT IGNORE 风格）
// 返回 (true, nil) 表示成功占用
// 返回 (false, nil) 表示已存在（需查旧记录）
func (r *IdempotencyRepo) Reserve(ctx context.Context, idemKey, userID string) (bool, error) {
	rec := model.IdempotencyKey{
		IdemKey:   idemKey,
		UserID:    userID,
		Status:    "pending",
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	err := r.db.WithContext(ctx).Create(&rec).Error
	if err == nil {
		return true, nil
	}
	if isDuplicateKeyError(err) {
		return false, nil
	}
	return false, err
}

// Complete 标记幂等键完成，存响应
func (r *IdempotencyRepo) Complete(ctx context.Context, idemKey, response string) error {
	return r.db.WithContext(ctx).
		Model(&model.IdempotencyKey{}).
		Where("idem_key = ?", idemKey).
		Updates(map[string]any{
			"status":   "done",
			"response": response,
		}).Error
}

// GetResponse 取幂等键对应的响应（重复请求时直接返回）
func (r *IdempotencyRepo) GetResponse(ctx context.Context, idemKey string) (string, error) {
	var rec model.IdempotencyKey
	err := r.db.WithContext(ctx).
		Where("idem_key = ? AND status = ?", idemKey, "done").
		First(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return rec.Response, nil
}

func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return contains(msg, "Duplicate entry") || contains(msg, "1062")
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
```

## 5. Redis 限流中间件

> 🎯 **滑动窗口**最常用也最准。今天先写最简版（INCR），Day 4 改成令牌桶。

```go
package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"chat-service/internal/handler"
)

// RateLimit 按 user_id 限流（窗口 = 1s, max = 5）
func RateLimit(rdb *redis.Client, max int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			userID = c.ClientIP() // 没传用户就用 IP
		}

		key := "ratelimit:user:" + userID
		ctx := c.Request.Context()

		// 经典做法：INCR + EXPIRE（第一次 INCR 时设置过期）
		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			c.Next() // Redis 挂了不要直接拒服务
			return
		}
		if count == 1 {
			rdb.Expire(ctx, key, window)
		}

		if count > int64(max) {
			c.AbortWithStatusJSON(http.StatusOK, handler.Response{
				Code:    handler.CodeRateLimited,
				Message: "请求过于频繁，请稍后再试",
			})
			return
		}
		c.Next()
	}
}
```

## 6. Redis 会话缓存

```go
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"chat-service/internal/model"
	"chat-service/internal/repository"
)

type CacheService struct {
	rdb    *redis.Client
	msgRepo *repository.MessageRepo
}

func NewCacheService(rdb *redis.Client, msgRepo *repository.MessageRepo) *CacheService {
	return &CacheService{rdb: rdb, msgRepo: msgRepo}
}

const recentMessagesKey = "conv:recent:%s"

// GetRecentMessages 读最近 N 条（带 Redis 缓存）
func (s *CacheService) GetRecentMessages(ctx context.Context, convID string, n int) ([]model.Message, error) {
	key := fmt.Sprintf(recentMessagesKey, convID)

	// 1. 查缓存
	cached, err := s.rdb.LRange(ctx, key, 0, int64(n-1)).Result()
	if err == nil && len(cached) > 0 {
		var msgs []model.Message
		for _, s := range cached {
			var m model.Message
			if err := json.Unmarshal([]byte(s), &m); err == nil {
				msgs = append(msgs, m)
			}
		}
		if len(msgs) > 0 {
			return msgs, nil
		}
	}

	// 2. 缓存未命中，查库
	msgs, err := s.msgRepo.ListByConv(ctx, convID, n)
	if err != nil {
		return nil, err
	}

	// 3. 回写缓存（5 分钟过期）
	pipe := s.rdb.Pipeline()
	pipe.Del(ctx, key)
	for _, m := range msgs {
		b, _ := json.Marshal(m)
		pipe.RPush(ctx, key, b)
	}
	pipe.Expire(ctx, key, 5*time.Minute)
	pipe.Exec(ctx)

	return msgs, nil
}

// PushMessage 新消息写进缓存头部
func (s *CacheService) PushMessage(ctx context.Context, m *model.Message) error {
	key := fmt.Sprintf(recentMessagesKey, m.ConversationID)
	b, _ := json.Marshal(m)
	pipe := s.rdb.Pipeline()
	pipe.LPush(ctx, key, b)
	pipe.LTrim(ctx, key, 0, 49) // 缓存最多 50 条
	pipe.Expire(ctx, key, 5*time.Minute)
	_, err := pipe.Exec(ctx)
	return err
}
```

## 7. Chat Service 升级（接 DB + Redis）

```go
// internal/service/chat.go
package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"chat-service/internal/handler"
	"chat-service/internal/model"
	"chat-service/internal/repository"
)

type ChatService struct {
	llm       LLMClient
	convRepo  *repository.ConversationRepo
	msgRepo   *repository.MessageRepo
	idemRepo  *repository.IdempotencyRepo
	cacheSvc  *CacheService
}

func NewChatService(
	llm LLMClient,
	convRepo *repository.ConversationRepo,
	msgRepo *repository.MessageRepo,
	idemRepo *repository.IdempotencyRepo,
	cacheSvc *CacheService,
) *ChatService {
	return &ChatService{
		llm:      llm,
		convRepo: convRepo,
		msgRepo:  msgRepo,
		idemRepo: idemRepo,
		cacheSvc: cacheSvc,
	}
}

// StreamRequest 完整请求
type StreamRequest struct {
	UserID         string             `json:"user_id"`
	ConversationID string             `json:"conversation_id"`
	Model          string             `json:"model"`
	Messages       []Message          `json:"messages"`
	IdempotencyKey string             `json:"idempotency_key"`
	TraceID        string             `json:"trace_id"`
}

// StreamChat 流式聊天（完整版）
func (s *ChatService) StreamChat(
	ctx context.Context,
	req StreamRequest,
	onChunk func(chunk string) error,
) (*ChatResponse, error) {
	// 1. 幂等检查
	if req.IdempotencyKey != "" {
		ok, err := s.idemRepo.Reserve(ctx, req.IdempotencyKey, req.UserID)
		if err != nil {
			return nil, handler.Internal(err)
		}
		if !ok {
			// 已存在，查旧响应直接返回
			cached, _ := s.idemRepo.GetResponse(ctx, req.IdempotencyKey)
			if cached != "" {
				for _, c := range cached {
					if err := onChunk(string(c)); err != nil {
						return nil, err
					}
				}
				return &ChatResponse{Content: cached}, nil
			}
		}
	}

	// 2. 拿 / 建会话
	conv, err := s.convRepo.GetOrCreate(ctx, req.ConversationID, req.UserID, req.Model)
	if err != nil {
		return nil, handler.Internal(err)
	}

	// 3. 读历史消息，拼到 prompt 里
	history, err := s.cacheSvc.GetRecentMessages(ctx, conv.ConversationID, 20)
	if err != nil {
		return nil, handler.Internal(err)
	}
	allMessages := mergeMessages(history, req.Messages)

	// 4. 调 LLM
	var fullContent string
	streamReq := ChatRequest{
		Model:    req.Model,
		Messages: allMessages,
		Stream:   true,
	}

	resp, err := s.llm.Stream(ctx, streamReq, func(chunk string) error {
		fullContent += chunk
		return onChunk(chunk)
	})
	if err != nil {
		return nil, handler.NewBizError(handler.CodeUpstreamFailed, "模型调用失败", err)
	}

	// 5. 落库：用户消息 + 助手回复
	now := time.Now()
	for _, m := range req.Messages {
		if m.Role == "user" {
			s.msgRepo.Insert(ctx, &model.Message{
				ConversationID: conv.ConversationID,
				UserID:         req.UserID,
				Role:           m.Role,
				Content:        m.Content,
				Model:          req.Model,
				TraceID:        req.TraceID,
				CreatedAt:      now,
			})
		}
	}

	assistantMsg := &model.Message{
		ConversationID: conv.ConversationID,
		UserID:         req.UserID,
		Role:           "assistant",
		Content:        resp.Content,
		TokensIn:       resp.Usage.PromptTokens,
		TokensOut:      resp.Usage.CompletionTokens,
		Model:          req.Model,
		TraceID:        req.TraceID,
		CreatedAt:      now,
	}
	s.msgRepo.Insert(ctx, assistantMsg)
	s.cacheSvc.PushMessage(ctx, assistantMsg)
	s.convRepo.Touch(ctx, conv.ConversationID)

	// 6. 记幂等
	if req.IdempotencyKey != "" {
		s.idemRepo.Complete(ctx, req.IdempotencyKey, resp.Content)
	}

	return resp, nil
}

func mergeMessages(history []model.Message, newMsgs []Message) []Message {
	var all []Message
	for _, h := range history {
		all = append(all, Message{Role: h.Role, Content: h.Content})
	}
	all = append(all, newMsgs...)
	return all
}

// 防止 uuid 包被删除
var _ = uuid.New
```

## 8. 路由更新（main.go）

```go
import (
    "chat-service/internal/infra"
    "chat-service/internal/middleware"
)

func main() {
    cfg, _ := config.Load()
    logger, _ := zap.NewProduction()
    defer logger.Sync()

    // 初始化基础设施
    db, err := infra.NewMySQL(&cfg.MySQL)
    if err != nil { logger.Fatal("mysql", zap.Error(err)) }
    rdb, err := infra.NewRedis(&cfg.Redis)
    if err != nil { logger.Fatal("redis", zap.Error(err)) }

    // Repository
    convRepo := repository.NewConversationRepo(db)
    msgRepo := repository.NewMessageRepo(db)
    idemRepo := repository.NewIdempotencyRepo(db)

    // Service
    llmClient := service.NewLLMClient(cfg)
    cacheSvc := service.NewCacheService(rdb, msgRepo)
    chatSvc := service.NewChatService(llmClient, convRepo, msgRepo, idemRepo, cacheSvc)
    chatHandler := handler.NewChatHandler(chatSvc)

    // Gin
    gin.SetMode(gin.ReleaseMode)
    r := gin.New()
    r.Use(gin.Recovery())
    r.Use(requestLogger(logger))

    // 鉴权（占位，从 header 读 user_id）
    r.Use(middleware.AuthFromHeader())

    r.GET("/healthz", handler.HealthCheck)

    v1 := r.Group("/v1")
    v1.Use(middleware.RateLimit(rdb, 5, time.Second))
    {
        v1.POST("/chat", chatHandler.NonStreamChat)
        v1.POST("/chat/stream", chatHandler.StreamChat)
        v1.GET("/conversations", chatHandler.ListConversations)
        v1.GET("/conversations/:id/messages", chatHandler.ListMessages)
    }

    srv := &http.Server{
        Addr:    cfg.Server.Addr,
        Handler: r,
    }
    // ... 启动省略
}

// AuthFromHeader 简单鉴权中间件（仅作演示）
func AuthFromHeader() gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.GetHeader("X-User-ID")
        if userID == "" {
            userID = "anonymous" // 真实环境会强制要求登录
        }
        c.Set("user_id", userID)
        c.Next()
    }
}
```

## 9. 测试场景

### 9.1 启动依赖

没有 MySQL/Redis 的同学用 Docker 起：

```bash
# 启动 MySQL
docker run -d --name mysql-dev -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=root123 \
  -e MYSQL_DATABASE=chat_service \
  mysql:8.0

# 启动 Redis
docker run -d --name redis-dev -p 6379:6379 redis:7-alpine

# 跑迁移（可选）
mysql -h 127.0.0.1 -u root -proot123 chat_service < migrations/001_init.sql
```

### 9.2 测试多轮对话

```bash
# 第一轮
curl -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_001" \
  -d '{
    "messages":[{"role":"user","content":"我叫叶子"}]
  }'

# 第二轮（传 conversation_id）
curl -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_001" \
  -d '{
    "conversation_id":"conv_xxx",
    "messages":[{"role":"user","content":"我叫什么？"}]
  }'
```

虽然 mock 不会真的"记住"，但**消息已经落库**了。Day 3 接真模型后就能看到效果。

### 9.3 测试限流

```bash
# 1 秒内连发 10 次（中间不要 sleep）
for i in $(seq 1 10); do
  curl -s -X POST http://localhost:8080/v1/chat \
    -H "Content-Type: application/json" \
    -H "X-User-ID: user_002" \
    -d '{"messages":[{"role":"user","content":"ping"}]}' &
done
wait
```

应该看到部分请求返回 `{"code":42900,"message":"请求过于频繁"}`。

### 9.4 测试幂等

```bash
curl -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_003" \
  -d '{
    "idempotency_key":"idem_test_001",
    "messages":[{"role":"user","content":"重复请求测试"}]
  }'

# 立即再发一次同样的请求，应该直接返回上次的响应
curl -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_003" \
  -d '{
    "idempotency_key":"idem_test_001",
    "messages":[{"role":"user","content":"重复请求测试"}]
  }'
```

## 10. Day 2 复盘清单

| 检查点 | 完成 |
|-------|------|
| MySQL 连接成功，表已建好 | ☐ |
| Redis 连接成功 | ☐ |
| 聊天消息能落 `messages` 表 | ☐ |
| 多轮对话历史能查出来 | ☐ |
| 限流：超过阈值返回 42900 | ☐ |
| 幂等：相同 key 重复请求不重复扣 | ☐ |
| EXPLAIN 看过一次查询计划 | ☐ |
| 日志里能看到 user_id 和 trace_id | ☐ |

## 📝 Day 2 作业

### 作业 1：加 List Conversations 和 List Messages 接口

```go
// handler
v1.GET("/conversations", chatHandler.ListConversations)
v1.GET("/conversations/:id/messages", chatHandler.ListMessages)
```

返回格式：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [...],
    "total": 10
  }
}
```

### 作业 2：加一个清理 idempotency_keys 的定时任务

```go
// 每小时跑一次，删除 24 小时前的
go func() {
    ticker := time.NewTicker(time.Hour)
    defer ticker.Stop()
    for range ticker.C {
        db.Where("expires_at < ?", time.Now()).Delete(&model.IdempotencyKey{})
    }
}()
```

### 作业 3：研究一下"防重放"

如果同一请求在 1 秒内发 10 次（哪怕 idempotency_key 不同），上游模型会被打爆。加一个**业务级防重放**：用 `user_id + 内容哈希` 当 key，2 秒内重复直接返回上次结果。

提示：
```go
import "crypto/sha256"
// 用 SHA256(user_id + 消息拼接) 做 Redis key，TTL=2s
```

## 💡 重点提示

1. **历史消息一定要裁剪**：不要无脑塞 100 条 prompt，token 爆炸又费钱。**保留最近 10~20 条 + 摘要**是行业标准
2. **限流按用户 + 按接口分开**：全局限流太粗，登录、聊天、上传应该各有限额
3. **幂等键设计要带用户**：`user_001:idem_xxx`，防止不同用户撞 key
4. **WriteTimeout 还是不要设太短**：SSE 长连接，Day 1 已经踩过坑了
5. **MySQL 慢查询要会看**：`SHOW PROCESSLIST` 看实时，`slow_query_log` 看历史

## ⏭️ Day 3 预告

明天把这套接上**真模型**（OpenAI / Claude 二选一），加 Function Calling 工具调用，加最小 RAG（本地文本检索），把 mock 切到生产路径。

---

**💪 Day 2 跑通，你就有了"记忆"和"防御"。明天开始让模型真的"开口说话"。**