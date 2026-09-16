# 第30课（Week5 Day 1）：Gin + SSE + 调假 LLM 跑通聊天接口

> 学习时间：6~8 小时 | 难度：⭐⭐⭐
> **日期**：9 月 12 日（今天）
> **目标**：本地能跑一个 Gin 服务，POST `/chat` 接请求、SSE 流式吐字返回

## 🎯 交卷标准（不达到别往下走）

- [ ] 项目 `chat-service` 在本地能 `go run` 起来
- [ ] `curl -N http://localhost:8080/v1/chat/stream` 能一个字一个字返回内容
- [ ] 代码分层清晰：`handler / service / repository`，每一层你知道在干什么
- [ ] 会读 `.env` / `os.Getenv` 加载配置
- [ ] 统一错误码（哪怕只是几个 if），不直接 `c.JSON(500, "崩了")`
- [ ] 能跟面试官 5 分钟讲清：路由怎么进来的 → middleware 干了啥 → handler 怎么 → service 怎么 → 返回为啥是 SSE

---

## 📋 本课目标

- 巩固 Gin 路由 + 中间件（[week2/day9](../week2/day9-gin-advanced.md) 已有基础）
- 学会 SSE（Server-Sent Events）流式响应写法
- 掌握 LLM 客户端的接口设计：哪怕今天调的是假实现，接口要像真接口
- 学会读环境变量加载配置
- 建立 chat-service 的工程目录骨架

## 1. 项目初始化

### 1.1 初始化 Module

```bash
cd D:\webProject\go
mkdir chat-service
cd chat-service
go mod init chat-service
```

### 1.2 安装依赖

```bash
# Gin
go get github.com/gin-gonic/gin@latest

# 配置加载（Viper 或 godotenv，二选一）
go get github.com/joho/godotenv@latest

# 日志
go get go.uber.org/zap@latest

# UUID
go get github.com/google/uuid@latest
```

> 💡 **为什么不用 Viper**：Viper 功能多但重，5 天速成用 godotenv 够，进组再切。

### 1.3 目录骨架

按 [Week5 README](./README.md) 里那张目录结构建好，今天至少要落地：

```
chat-service/
├── cmd/server/main.go
├── internal/
│   ├── handler/chat.go
│   ├── service/chat.go
│   ├── service/llm.go
│   └── config/config.go
├── .env.example
├── go.mod
└── README.md (先空着)
```

## 2. 配置加载（internal/config/config.go）

```go
package config

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig
	LLM      LLMConfig
	LogLevel string
}

type ServerConfig struct {
	Addr         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type LLMConfig struct {
	BaseURL      string        // 兼容 OpenAI / Anthropic 的网关地址
	APIKey       string        // 留空 = 走 mock
	MockMode     bool          // true 表示不真发请求
	Model        string        // 例如 gpt-4o-mini / claude-sonnet
	MaxTokens     int
	Temperature  float64
	StreamChunk  time.Duration // mock 模式下每个字间隔，模拟流式
	RequestTimeout time.Duration
}

func Load() (*Config, error) {
	// 加载 .env（找不到也不报错，方便生产用环境变量）
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Addr:         getEnv("SERVER_ADDR", ":8080"),
			ReadTimeout:  getDuration("SERVER_READ_TIMEOUT", "10s"),
			WriteTimeout: getDuration("SERVER_WRITE_TIMEOUT", "60s"), // SSE 长连接要放宽
		},
		LLM: LLMConfig{
			BaseURL:        getEnv("LLM_BASE_URL", "https://api.openai.com/v1"),
			APIKey:         getEnv("LLM_API_KEY", ""),
			MockMode:       getBool("LLM_MOCK_MODE", true),
			Model:          getEnv("LLM_MODEL", "gpt-4o-mini"),
			MaxTokens:      getInt("LLM_MAX_TOKENS", 1024),
			Temperature:    getFloat("LLM_TEMPERATURE", 0.7),
			StreamChunk:    getDuration("LLM_MOCK_STREAM_CHUNK", "30ms"),
			RequestTimeout: getDuration("LLM_REQUEST_TIMEOUT", "30s"),
		},
		LogLevel: getEnv("LOG_LEVEL", "info"),
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func getFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}

func getBool(key string, def bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func getDuration(key, def string) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
		if i, err := strconv.Atoi(v); err == nil {
			return time.Duration(i) * time.Second
		}
	}
	d, _ := time.ParseDuration(def)
	return d
}
```

### .env.example

```bash
# Server
SERVER_ADDR=:8080
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=60s
LOG_LEVEL=info

# LLM（默认走 mock，把 LLM_MOCK_MODE 改 false 就能切真接口）
LLM_MOCK_MODE=true
LLM_BASE_URL=https://api.openai.com/v1
LLM_API_KEY=
LLM_MODEL=gpt-4o-mini
LLM_MAX_TOKENS=1024
LLM_TEMPERATURE=0.7
LLM_MOCK_STREAM_CHUNK=30ms
LLM_REQUEST_TIMEOUT=30s
```

> 💡 **关键设计**：`MockMode` 字段。今天 mock 是为了不烧 token 跑通流程，明天换真接口只要改环境变量。

## 3. 统一错误处理（internal/handler/common.go）

```go
package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrCode 业务错误码
type ErrCode int

const (
	CodeSuccess        ErrCode = 0
	CodeBadRequest     ErrCode = 40000
	CodeUnauthorized   ErrCode = 40100
	CodeInternal       ErrCode = 50000
	CodeUpstreamFailed ErrCode = 50200
	CodeRateLimited    ErrCode = 42900
)

// BizError 业务错误
type BusinessError struct {
	Code    ErrCode
	Message string
	Err     error
}

func (e *BusinessError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *BusinessError) Unwrap() error { return e.Err }

func NewBizError(code ErrCode, msg string, err error) *BusinessError {
	return &BusinessError{Code: code, Message: msg, Err: err}
}

// 快捷构造
func BadRequest(msg string) *BusinessError  { return NewBizError(CodeBadRequest, msg, nil) }
func Unauthorized(msg string) *BusinessError { return NewBizError(CodeUnauthorized, msg, nil) }
func Internal(err error) *BusinessError     { return NewBizError(CodeInternal, "服务暂时不可用", err) }

// Response 统一响应结构
type Response struct {
	Code    ErrCode    `json:"code"`
	Message string     `json:"message"`
	Data    any        `json:"data,omitempty"`
	TraceID string     `json:"trace_id,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// Fail 写错误响应（业务错误）
func Fail(c *gin.Context, err error) {
	var biz *BusinessError
	if errors.As(err, &biz) {
		c.JSON(http.StatusOK, Response{
			Code:    biz.Code,
			Message: biz.Message,
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Code:    CodeInternal,
		Message: "服务暂时不可用",
	})
}
```

> 💡 **设计取舍**：业务错误一律返回 HTTP 200，错误码放在 body 里。前端只判 `code`。这是国内后端常规做法（参考字节、阿里）。

## 4. LLM 客户端接口（internal/service/llm.go）

> 🎯 这是这 5 天的核心。今天调假实现，明天换真实现。

```go
package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"chat-service/internal/config"
)

// Message 消息结构（兼容 OpenAI / Anthropic）
type Message struct {
	Role    string `json:"role"`    // system / user / assistant
	Content string `json:"content"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
	Stream      bool      `json:"stream,omitempty"`
}

// ChatResponse 非流式响应
type ChatResponse struct {
	Content string
	Usage   Usage
}

type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// LLMClient LLM 客户端接口（设计先于实现）
type LLMClient interface {
	// Chat 非流式聊天
	Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)

	// Stream 流式聊天，通过 onChunk 回调逐片返回；返回完整内容 + usage
	Stream(ctx context.Context, req ChatRequest, onChunk func(chunk string) error) (*ChatResponse, error)
}

// NewLLMClient 工厂方法（根据配置选 mock 还是 real）
func NewLLMClient(cfg *config.Config) LLMClient {
	if cfg.LLM.MockMode {
		return NewMockClient(cfg.LLM)
	}
	// Day 3 才接真接口
	return NewMockClient(cfg.LLM)
}

// ============ Mock 实现 ============

type mockClient struct {
	cfg config.LLMConfig
}

func NewMockClient(cfg config.LLMConfig) *mockClient {
	return &mockClient{cfg: cfg}
}

func (m *mockClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	fullText, err := m.simulateStream(ctx, req, nil)
	if err != nil {
		return nil, err
	}
	return &ChatResponse{
		Content: fullText,
		Usage: Usage{
			PromptTokens:     estimateTextTokens(joinMessages(req.Messages)),
			CompletionTokens: estimateTextTokens(fullText),
			TotalTokens:      estimateTextTokens(fullText) * 2,
		},
	}, nil
}

func (m *mockClient) Stream(ctx context.Context, req ChatRequest, onChunk func(chunk string) error) (*ChatResponse, error) {
	fullText, err := m.simulateStream(ctx, req, onChunk)
	if err != nil {
		return nil, err
	}
	return &ChatResponse{
		Content: fullText,
		Usage:   Usage{PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150},
	}, nil
}

// simulateStream 按 m.cfg.StreamChunk 一个字一个字吐
func (m *mockClient) simulateStream(ctx context.Context, req ChatRequest, onChunk func(string) error) (string, error) {
	reply := m.fakeAnswer(req)
	var sb strings.Builder

	// 按 rune 切，兼容中文
	runes := []rune(reply)
	for _, r := range runes {
		select {
		case <-ctx.Done():
			return sb.String(), ctx.Err()
		case <-time.After(m.cfg.StreamChunk):
		}

		sb.WriteRune(r)
		if onChunk != nil {
			if err := onChunk(string(r)); err != nil {
				return sb.String(), err
			}
		}
	}
	return sb.String(), nil
}

// fakeAnswer 根据用户最后一条消息生成假回答
func (m *mockClient) fakeAnswer(req ChatRequest) string {
	last := lastUserMessage(req.Messages)
	if last == "" {
		return "你好，我是 mock，可以问我任何问题。"
	}

	// 简单模板回复
	templates := []string{
		"你说的是：「%s」。这是个有趣的问题，让我想想。",
		"关于「%s」我理解你的意思，不过我只是个 mock，需要 Day 3 才接真模型。",
		"「%s」——好的，已收到。当前在 mock 模式，正式上线前请切换 LLM_MOCK_MODE=false。",
	}
	idx := len(last) % len(templates)
	return fmt.Sprintf(templates[idx], truncate(last, 50))
}

func lastUserMessage(msgs []Message) string {
	for i := len(msgs) - 1; i >= 0; i-- {
		if msgs[i].Role == "user" {
			return msgs[i].Content
		}
	}
	return ""
}

func joinMessages(msgs []Message) string {
	var sb strings.Builder
	for _, m := range msgs {
		sb.WriteString(m.Role)
		sb.WriteString(": ")
		sb.WriteString(m.Content)
		sb.WriteString("\n")
	}
	return sb.String()
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

// estimateTextTokens 粗略估算 token 数（中英文混合，1 token ≈ 1.5 字符）
func estimateTextTokens(s string) int {
	if s == "" {
		return 0
	}
	return len([]rune(s)) * 2 / 3
}
```

> 💡 **关键设计**：接口 `LLMClient` 在前，Mock 在后。明天加真实现时**业务层一行都不用改**，这就是依赖接口的好处（参考 [week4/day22 接口 Mock 模式](../week4/day22-testing.md)）。

## 5. Chat 业务层（internal/service/chat.go）

```go
package service

import (
	"context"
	"errors"

	"chat-service/internal/handler"
)

// ChatService 聊天服务（业务编排）
type ChatService struct {
	llm LLMClient
}

func NewChatService(llm LLMClient) *ChatService {
	return &ChatService{llm: llm}
}

// SendChat 非流式（先不做，Day 3 接 Agent 时加）
// 这里先留接口位

// StreamChat 流式聊天
// 每一行（chunk）通过 onChunk 回调返出去（handler 层转 SSE）
func (s *ChatService) StreamChat(
	ctx context.Context,
	req ChatRequest,
	onChunk func(chunk string) error,
) (*ChatResponse, error) {
	if len(req.Messages) == 0 {
		return nil, handler.BadRequest("messages 不能为空")
	}

	// 兜底参数
	if req.Model == "" {
		req.Model = "gpt-4o-mini"
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 1024
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}

	resp, err := s.llm.Stream(ctx, req, onChunk)
	if err != nil {
		// 区分：客户端断连（Abort）不算错误
		if errors.Is(err, context.Canceled) {
			return resp, nil
		}
		return nil, handler.NewBizError(handler.CodeUpstreamFailed, "上游模型调用失败", err)
	}
	return resp, nil
}
```

## 6. SSE Handler（internal/handler/chat.go）

> 🎯 **今天的难点就在这**。SSE 写法面试必问，得能手写。

```go
package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"chat-service/internal/service"
)

type ChatHandler struct {
	chatSvc *service.ChatService
}

func NewChatHandler(chatSvc *service.ChatService) *ChatHandler {
	return &ChatHandler{chatSvc: chatSvc}
}

// StreamChatRequest 客户端请求
type StreamChatRequest struct {
	Messages []service.Message `json:"messages" binding:"required,min=1"`
	Model    string            `json:"model"`
}

// StreamChat POST /v1/chat/stream —— 真正的 SSE 流式接口
func (h *ChatHandler) StreamChat(c *gin.Context) {
	var req StreamChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, BadRequest("请求参数错误: "+err.Error()))
		return
	}

	// === SSE 响应头 ===
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // 关闭 Nginx 缓冲（部署时用得到）
	c.Writer.WriteHeader(200)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		Fail(c, Internal(fmt.Errorf("response writer 不支持 flush")))
		return
	}

	// === 业务调用 ===
	onChunk := func(chunk string) error {
		// SSE 协议：每个 event 用 \n\n 结束，data 是 JSON
		payload, _ := json.Marshal(map[string]any{
			"type":    "chunk",
			"content": chunk,
		})
		if _, err := fmt.Fprintf(c.Writer, "data: %s\n\n", payload); err != nil {
			return err
		}
		flusher.Flush() // 关键：写完就 flush，否则会等缓冲
		return nil
	}

	chatReq := service.ChatRequest{
		Model:    req.Model,
		Messages: req.Messages,
		Stream:   true,
	}

	resp, err := h.chatSvc.StreamChat(c.Request.Context(), chatReq, onChunk)
	if err != nil {
		// 流式过程中出错，发一个 error 事件让前端知道
		errPayload, _ := json.Marshal(map[string]any{
			"type":    "error",
			"message": err.Error(),
		})
		fmt.Fprintf(c.Writer, "event: error\ndata: %s\n\n", errPayload)
		flusher.Flush()
		return
	}

	// === 末尾发一个 done 事件，带 usage ===
	donePayload, _ := json.Marshal(map[string]any{
		"type":  "done",
		"usage": resp.Usage,
	})
	fmt.Fprintf(c.Writer, "event: done\ndata: %s\n\n", donePayload)
	flusher.Flush()
}

// NonStreamChat POST /v1/chat —— 非流式，方便 curl 测试
func (h *ChatHandler) NonStreamChat(c *gin.Context) {
	var req StreamChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Fail(c, BadRequest("请求参数错误: "+err.Error()))
		return
	}

	resp, err := h.chatSvc.StreamChat(c.Request.Context(), service.ChatRequest{
		Model:    req.Model,
		Messages: req.Messages,
		Stream:   false,
	}, nil)
	if err != nil {
		Fail(c, err)
		return
	}
	OK(c, resp)
}

// HealthCheck GET /healthz —— K8s liveness 探针用
func HealthCheck(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok", "time": time.Now().Unix()})
}

// 防止 goimports 把 io 这个包删除
var _ = io.Discard
```

## 7. 主程序（cmd/server/main.go）

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"chat-service/internal/config"
	"chat-service/internal/handler"
	"chat-service/internal/service"
)

func main() {
	// 1. 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	// 2. 初始化日志
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// 3. 初始化依赖
	llmClient := service.NewLLMClient(cfg)
	chatSvc := service.NewChatService(llmClient)
	chatHandler := handler.NewChatHandler(chatSvc)

	// 4. 路由
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger(logger))

	r.GET("/healthz", handler.HealthCheck)

	v1 := r.Group("/v1")
	{
		v1.POST("/chat", chatHandler.NonStreamChat)
		v1.POST("/chat/stream", chatHandler.StreamChat)
	}

	// 5. HTTP 服务 + 优雅关闭
	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	go func() {
		logger.Info("server starting", zap.String("addr", cfg.Server.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("listen", zap.Error(err))
		}
	}()

	// 等待信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("shutdown", zap.Error(err))
	}
	logger.Info("bye")
}

// requestLogger 简单请求日志中间件
func requestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		logger.Info("request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("lat", time.Since(start)),
		)
	}
}
```

> 💡 **优雅关闭**：SSE 长连接必须能优雅关闭，否则客户端会看到莫名其妙的断线。`srv.Shutdown(ctx)` 等现有连接把当前 chunk 吐完再退。

## 8. 启动 & 测试

### 8.1 启动

```bash
cd D:\webProject\go\chat-service
cp .env.example .env
go mod tidy
go run cmd/server/main.go
```

输出：
```
{"level":"info","msg":"server starting","addr":":8080"}
```

### 8.2 非流式测试

```bash
curl -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"你是谁"}]}'
```

预期：
```json
{"code":0,"message":"success","data":{"content":"你说的是：「你是谁」...","usage":{...}}}
```

### 8.3 流式测试（重点）

```bash
curl -N -X POST http://localhost:8080/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"讲个笑话"}]}'
```

预期（一行一行打印，每行 30ms 一个字）：
```
data: {"type":"chunk","content":"你"}

data: {"type":"chunk","content":"说"}

data: {"type":"chunk","content":"的"}

data: {"type":"chunk","content":"是"}

...

event: done
data: {"type":"done","usage":{...}}
```

## 9. Day 1 复盘清单

| 检查点 | 完成 |
|-------|------|
| 项目能 `go run` 起来 | ☐ |
| `/healthz` 返回 200 | ☐ |
| `/v1/chat` 非流式工作 | ☐ |
| `/v1/chat/stream` SSE 一字一字吐 | ☐ |
| 错误参数返回统一错误码 | ☐ |
| Ctrl+C 能优雅关闭（无残留进程） | ☐ |
| 能讲清 handler→service→llmClient 三层各自做什么 | ☐ |

## 📝 Day 1 作业

### 作业 1：模拟"超时打断"

在 handler 里加一个测试：客户端发请求后 3 秒内断开（`curl --max-time 3`），服务端不应该 panic、不应该有 goroutine 泄漏。

```bash
# 测试
curl --max-time 1 -N -X POST http://localhost:8080/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"讲个长一点的故事"}]}'
```

观察服务端日志是不是 `client disconnected` 类似输出，进程没崩。

### 作业 2：加一个简单的 trace_id 中间件

```go
// internal/middleware/trace.go
func TraceID() gin.HandlerFunc {
    return func(c *gin.Context) {
        traceID := c.GetHeader("X-Trace-ID")
        if traceID == "" {
            traceID = uuid.NewString()
        }
        c.Set("trace_id", traceID)
        c.Header("X-Trace-ID", traceID)
        c.Next()
    }
}
```

注册到全局，并在响应 / 日志里带上。**这是 Day 4 账单日志的前置**。

### 作业 3：能讲清这五件事

1. 为什么 SSE 协议用 `data: xxx\n\n` 这种格式？（HTTP 长连接 + 分隔）
2. 为什么要 `flusher.Flush()`？（Gin 默认会缓冲，不 flush 客户端看不到）
3. 为什么要单独设置 `X-Accel-Buffering: no`？（Nginx 反代会缓冲 SSE）
4. `WriteTimeout: 60s` 是给谁的？（整个响应超时，SSE 不能设太短）
5. Mock 客户端和真客户端为啥共用同一个接口？（依赖倒置，Day 3 一行不改切换）

## 💡 重点提示

1. **SSE 是流式响应的"最低成本"方案**，比 WebSocket 简单，比轮询优雅，AI 对话场景首选
2. **接口先于实现**：今天写的 `LLMClient` 接口，Day 3 接真 OpenAI / Claude 时业务层零改动
4. **WriteTimeout 一定要长**：SSE 连接可能要开几分钟，设 60s 就太短，先写 0（不限）也可以
5. **优雅关闭必加**：K8s 滚动更新时如果不优雅关，正在聊天的用户会看到 502

## ⏭️ Day 2 预告

明天把这套接到 MySQL（聊天记录落库）和 Redis（会话缓存 + 限流），把"聊天"变成"有记忆的聊天"，还要做幂等防重复请求。

---

**💪 Day 1 走完，你就有了一个能面试拿出来讲���项目骨架。明天开始加"血肉"。**