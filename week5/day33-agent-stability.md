# 第33课（Week5 Day 4）：Agent 编排 + 稳定性（超时/重试/降级/账单）

> 学习时间：6~8 小时 | 难度：⭐⭐⭐⭐
> **日期**：9 月 15 日
> **目标**：Agent 真能干活，模型挂了不整站挂，每个请求的成本可追溯

## 🎯 交卷标准

- [ ] 用户问 → Agent 决定要不要查 → 查工具 → 再回答，完整闭环
- [ ] 模型调用有超时（单次 30s，整体 60s）
- [ ] 网络抖动自动重试 2 次（指数退避）
- [ ] 模型挂了返回兜底文案，HTTP 200 而不是 500
- [ ] 每个请求的 token 用量 + 估算成本打到日志
- [ ] 账单查询接口 `GET /v1/admin/billing?user_id=xxx` 返回今日用量
- [ ] 写一个能讲清的"架构图"说给面试官听

---

## 📋 本课目标

- 掌握生产级超时控制（context + 整体超时）
- 学会带指数退避的重试
- 学会优雅降级（fallback 文案）
- 学会成本/账单打点
- 巩固 Agent 编排（Day 3 的基础上加边界保护）
- 给整个服务加可观测性（日志 + 错误捕获）

## 1. 稳定性三件套：超时、重试、降级

### 1.1 整体超时控制

```go
// internal/service/llm.go
package service

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrTimeout 超时错误
var ErrTimeout = errors.New("llm: request timeout")

// WithTimeout 给任意 LLM 调用包一层超时
func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return parent, func() {}
	}
	return context.WithTimeout(parent, timeout)
}
```

### 1.2 指数退避重试

```go
// internal/service/retry.go
package service

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"time"
)

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries  int           // 最大重试次数（不含首次）
	BaseDelay   time.Duration // 基础延迟
	MaxDelay    time.Duration // 最大延迟
	RetryOnErr  func(error) bool // 哪些错误重试
}

// DefaultRetryConfig 默认配置
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 2,
		BaseDelay:  500 * time.Millisecond,
		MaxDelay:   5 * time.Second,
		RetryOnErr: func(err error) bool {
			if err == nil {
				return false
			}
			// 4xx（参数错）不重试，5xx / 网络错误重试
			msg := err.Error()
			if errors.Is(err, context.Canceled) {
				return false // 客户端断连不重试
			}
			if errors.Is(err, context.DeadlineExceeded) {
				return true // 超时可以重试（可能是网络抖动）
			}
			if containsAny(msg, []string{"400", "401", "403", "404", "invalid_request"}) {
				return false
			}
			return true
		},
	}
}

// Retry 带指数退避的重试
func Retry(ctx context.Context, cfg RetryConfig, name string, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt <= cfg.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := cfg.BaseDelay * (1 << uint(attempt-1))
			if delay > cfg.MaxDelay {
				delay = cfg.MaxDelay
			}
			// 加 jitter（防止雪崩）
			jitter := time.Duration(rand.Int63n(int64(delay / 4)))
			delay += jitter

			log.Printf("[retry] %s 第 %d 次重试，等待 %v", name, attempt, delay)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := fn()
		if err == nil {
			return nil
		}
		lastErr = err
		if !cfg.RetryOnErr(err) {
			return err
		}
	}
	return fmt.Errorf("重试 %d 次后仍失败: %w", cfg.MaxRetries, lastErr)
}

func containsAny(s string, subs []string) bool {
	for _, sub := range subs {
		if contains(s, sub) {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
```

### 1.3 用法（在 LLM Client 里）

```go
// internal/service/llm.go 的 mockClient.Stream 改成：
func (m *mockClient) Stream(ctx context.Context, req ChatRequest, onChunk func(string) error) (*ChatResponse, error) {
	var fullText string
	err := Retry(ctx, DefaultRetryConfig(), "mock-stream", func() error {
		var err error
		fullText, err = m.simulateStream(ctx, req, onChunk)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &ChatResponse{
		Content: fullText,
		Usage:   Usage{PromptTokens: 100, CompletionTokens: 50, TotalTokens: 150},
	}, nil
}
```

### 1.4 降级（Fallback）

```go
// internal/service/fallback.go
package service

// FallbackMessage 模型挂了时的兜底
var FallbackMessages = []string{
	"抱歉，我这边有点忙，过会儿再来试试。",
	"系统正在升级中，请稍后再试，或联系人工客服。",
	"我暂时无法回答你的问题，你可以换个问题试试。",
}

// PickFallback 随机挑一条
func PickFallback() string {
	return FallbackMessages[time.Now().Unix()%int64(len(FallbackMessages))]
}
```

> 💡 **降级原则**：
> 1. **不要让整个接口 500**——前端没法处理
> 2. **告知用户原因**——"模型暂时不可用"比"服务异常"清楚
> 3. **尽量返回结构化内容**——别让前端没法渲染

## 2. Agent 升级（生产版）

### 2.1 带边界保护的 Agent

```go
// internal/service/agent.go
package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// Agent 完整版
type Agent struct {
	llm       LLMClient
	tools     *ToolRegistry
	maxTurns  int
	timeout   time.Duration
	logger    *zap.Logger
}

func NewAgent(llm LLMClient, tools *ToolRegistry, logger *zap.Logger) *Agent {
	return &Agent{
		llm:      llm,
		tools:    tools,
		maxTurns: 5,
		timeout:  60 * time.Second,
		logger:   logger,
	}
}

// Run 主入口
func (a *Agent) Run(
	ctx context.Context,
	messages []Message,
	sysPrompt string,
	onChunk func(string) error,
) (*ChatResponse, error) {
	// 1. 整体超时
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	// 2. 拼 system prompt
	allMsgs := append([]Message{{Role: "system", Content: sysPrompt}}, messages...)

	// 3. Agent 主循环
	turn := 0
	var lastResp *ChatResponse
	for turn < a.maxTurns {
		turn++
		a.logger.Info("agent turn", zap.Int("turn", turn), zap.Int("messages", len(allMsgs)))

		// 3.1 调 LLM（带重试）
		var resp *ChatResponse
		var toolCalls []ToolCall
		err := Retry(ctx, DefaultRetryConfig(), "agent-llm", func() error {
			r, err := a.llm.Chat(ctx, ChatRequest{
				Model:    "",
				Messages: allMsgs,
				Stream:   false,
			})
			if err != nil {
				return err
			}
			// 真模型返回里要解析 tool_calls（这里简化）
			resp = r
			return nil
		})
		if err != nil {
			// 4. 兜底
			a.logger.Warn("agent llm failed, fallback", zap.Error(err))
			fb := PickFallback()
			if onChunk != nil {
				_ = onChunk(fb)
			}
			return &ChatResponse{Content: fb}, err
		}
		lastResp = resp

		// 3.2 模型没有要调工具 → 流式把 content 推给前端，结束
		if len(toolCalls) == 0 {
			if resp.Content != "" && onChunk != nil {
				_ = onChunk(resp.Content)
			}
			return resp, nil
		}

		// 3.3 有工具调用 → 逐个执行
		for _, tc := range toolCalls {
			tool, ok := a.tools.Get(tc.Name)
			if !ok {
				a.logger.Warn("unknown tool", zap.String("name", tc.Name))
				allMsgs = append(allMsgs, Message{Role: "tool", Content: fmt.Sprintf("未知工具: %s", tc.Name)})
				continue
			}

			// 单工具超时
			toolCtx, cancelTool := context.WithTimeout(ctx, 10*time.Second)
			result, err := tool.Handler(toolCtx, json.RawMessage(tc.Arguments))
			cancelTool()

			resultStr := fmt.Sprintf("%v", result)
			if err != nil {
				a.logger.Warn("tool failed", zap.String("tool", tc.Name), zap.Error(err))
				resultStr = fmt.Sprintf("工具执行失败: %v", err)
			}
			allMsgs = append(allMsgs, Message{Role: "tool", Content: resultStr})
		}
	}

	// 5. 超最大轮数
	a.logger.Warn("agent max turns reached", zap.Int("max", a.maxTurns))
	if lastResp != nil {
		return lastResp, nil
	}
	return &ChatResponse{Content: "抱歉，我没想清楚怎么回答你。"}, errors.New("agent max turns")
}

// ToolCall 工具调用（简化版）
type ToolCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}
```

> 💡 **生产 Agent 必备**：
> - **总超时**（防止用户问完跑了，服务还在跑）
> - **单工具超时**（一个工具卡了不阻塞整个流程）
> - **最大轮数**（防止死循环）
> - **降级**（模型挂时不整站挂）

### 2.2 真模型 + Tool Use 闭环

把 Day 3 的 OpenAI 工具调用循环搬过来，整合进 `Agent.Run`：

```go
// 替换 3.1 调 LLM 那段
err := Retry(ctx, DefaultRetryConfig(), "agent-llm", func() error {
    // OpenAI 调用（带 tools）
    resp, err := a.openaiCli.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
        Model:    a.cfg.Model,
        Messages: convertMessages(allMsgs),
        Tools:    convertTools(a.tools.List()),
    })
    if err != nil { return err }

    if len(resp.Choices) == 0 {
        return errors.New("no choices")
    }

    msg := resp.Choices[0].Message
    lastResp = &ChatResponse{
        Content: msg.Content,
        Usage:   convertUsage(resp.Usage),
    }

    // 解析 tool_calls
    toolCalls = nil
    for _, tc := range msg.ToolCalls {
        toolCalls = append(toolCalls, ToolCall{
            Name:      tc.Function.Name,
            Arguments: tc.Function.Arguments,
        })
    }

    // 把模型回复加入历史
    allMsgs = append(allMsgs, Message{Role: "assistant", Content: msg.Content})
    return nil
})
```

## 3. 成本打点 + 账单

### 3.1 模型单价（配置化）

```go
// internal/config/config.go
type LLMConfig struct {
	// ... 原有字段
	PricePer1kInput  float64 // 输入 token 单价（美元）
	PricePer1kOutput float64 // 输出 token 单价（美元）
}

// .env
LLM_PRICE_INPUT=0.00015   # gpt-4o-mini: $0.15/M tokens
LLM_PRICE_OUTPUT=0.0006   # gpt-4o-mini: $0.60/M tokens
```

### 3.2 成本计算 + 打点

```go
// internal/service/billing.go
package service

import (
	"context"
	"time"

	"go.uber.org/zap"

	"chat-service/internal/config"
)

type BillingService struct {
	cfg    *config.Config
	rdb    *redis.Client
	logger *zap.Logger
}

func NewBillingService(cfg *config.Config, rdb *redis.Client, logger *zap.Logger) *BillingService {
	return &BillingService{cfg: cfg, rdb: rdb, logger: logger}
}

// Record 记录一次调用的成本
func (b *BillingService) Record(ctx context.Context, userID, model string, usage Usage, traceID string) {
	cost := b.calcCost(usage)

	b.logger.Info("llm usage",
		zap.String("user_id", userID),
		zap.String("model", model),
		zap.Int("prompt_tokens", usage.PromptTokens),
		zap.Int("completion_tokens", usage.CompletionTokens),
		zap.Float64("cost_usd", cost),
		zap.String("trace_id", traceID),
	)

	// 累计到 Redis（key 格式：billing:user:2026-09-15）
	today := time.Now().Format("2006-01-02")
	key := "billing:user:" + today + ":" + userID

	pipe := b.rdb.Pipeline()
	pipe.HIncrBy(ctx, key, "prompt_tokens", int64(usage.PromptTokens))
	pipe.HIncrBy(ctx, key, "completion_tokens", int64(usage.CompletionTokens))
	pipe.HIncrByFloat(ctx, key, "cost_usd", cost)
	pipe.Expire(ctx, key, 30*24*time.Hour) // 保留 30 天
	pipe.Exec(ctx)
}

// Query 查询用户今日账单
func (b *BillingService) Query(ctx context.Context, userID string) (map[string]string, error) {
	today := time.Now().Format("2006-01-02")
	key := "billing:user:" + today + ":" + userID
	return b.rdb.HGetAll(ctx, key).Result()
}

func (b *BillingService) calcCost(usage Usage) float64 {
	inCost := float64(usage.PromptTokens) / 1000.0 * b.cfg.LLM.PricePer1kInput
	outCost := float64(usage.CompletionTokens) / 1000.0 * b.cfg.LLM.PricePer1kOutput
	return inCost + outCost
}
```

### 3.3 在 ChatService 里接入

```go
func (s *ChatService) StreamChat(ctx context.Context, req StreamRequest, onChunk func(string) error) (*ChatResponse, error) {
	// ... 业务逻辑

	resp, err := s.llm.Stream(ctx, streamReq, func(chunk string) error {
		return onChunk(chunk)
	})
	if err != nil {
		return nil, handler.NewBizError(handler.CodeUpstreamFailed, "模型调用失败", err)
	}

	// ★ 关键：每次调用都打点
	s.billingSvc.Record(ctx, req.UserID, req.Model, resp.Usage, req.TraceID)

	// ... 落库
}
```

### 3.4 账单查询 API

```go
// internal/handler/billing.go
func (h *BillingHandler) GetBilling(c *gin.Context) {
    userID := c.Query("user_id")
    if userID == "" {
        userID = c.GetString("user_id") // 从 AuthMiddleware 取
    }
    if userID == "" {
        Fail(c, BadRequest("user_id 必填"))
        return
    }

    data, err := h.billingSvc.Query(c.Request.Context(), userID)
    if err != nil {
        Fail(c, Internal(err))
        return
    }
    OK(c, data)
}
```

```go
// main.go 注册
admin := v1.Group("/admin")
admin.Use(middleware.RequireAdmin()) // Day 5 再写，今天先不做鉴权
{
    admin.GET("/billing", billingHandler.GetBilling)
}
```

## 4. 结构化日志（zap + trace_id）

### 4.1 请求日志中间件（升级版）

```go
// internal/middleware/logger.go
package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

const TraceIDKey = "trace_id"

func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// trace_id 优先从 header 取，没就生成
		traceID := c.GetHeader("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.NewString()
		}
		c.Set(TraceIDKey, traceID)
		c.Header("X-Trace-ID", traceID)

		c.Next()

		logger.Info("http",
			zap.String("trace_id", traceID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_id", c.GetString("user_id")),
		)
	}
}
```

### 4.2 业务日志助手

```go
// internal/middleware/context.go
package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// WithLogger 把 zap logger 放进 context
func WithLogger(base *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        traceID := c.GetString(TraceIDKey)
        userID := c.GetString("user_id")
        logger := base.With(
            zap.String("trace_id", traceID),
            zap.String("user_id", userID),
        )
        ctx := context.WithValue(c.Request.Context(), loggerKey{}, logger)
        c.Request = c.Request.WithContext(ctx)
        c.Next()
    }
}

// LoggerFromCtx 从 context 取 logger
func LoggerFromCtx(ctx context.Context) *zap.Logger {
    if l, ok := ctx.Value(loggerKey{}).(*zap.Logger); ok {
        return l
    }
    return zap.NewNop()
}

type loggerKey struct{}
```

### 4.3 用法（Service 层）

```go
func (s *ChatService) StreamChat(ctx context.Context, req StreamRequest, onChunk func(string) error) (*ChatResponse, error) {
    logger := middleware.LoggerFromCtx(ctx)
    logger.Info("chat start", zap.Int("msg_count", len(req.Messages)))

    // ... 业务

    logger.Info("chat done",
        zap.Int("input_tokens", resp.Usage.PromptTokens),
        zap.Int("output_tokens", resp.Usage.CompletionTokens),
    )
}
```

## 5. 错误分类（生产标准）

### 5.1 错误码体系（最终版）

```go
// internal/handler/common.go 补充
const (
    CodeSuccess         ErrCode = 0
    CodeBadRequest      ErrCode = 40000
    CodeUnauthorized    ErrCode = 40100
    CodeForbidden       ErrCode = 40300
    CodeNotFound        ErrCode = 40400
    CodeRateLimited     ErrCode = 42900
    CodeInternal        ErrCode = 50000
    CodeUpstreamTimeout ErrCode = 50400  // ★ 新增
    CodeUpstreamFailed  ErrCode = 50200  // ★ 上游失败
    CodeLLMFallback     ErrCode = 50300  // ★ 降级命中
)
```

### 5.2 错误捕获（兜底）

```go
// internal/middleware/recovery.go
package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"chat-service/internal/handler"
)

func Recovery(logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        defer func() {
            if err := recover(); err != nil {
                logger.Error("panic recovered",
                    zap.Any("error", err),
                    zap.String("path", c.Request.URL.Path),
                    zap.String("stack", string(debug.Stack())),
                )
                c.AbortWithStatusJSON(http.StatusOK, handler.Response{
                    Code:    handler.CodeInternal,
                    Message: "服务内部错误",
                })
            }
        }()
        c.Next()
    }
}
```

## 6. 完整的 ChatService 流程

```
用户请求 (含 trace_id)
  │
  ▼
RequestLogger 中间件（生成/透传 trace_id）
  │
  ▼
Auth 中间件（设 user_id）
  │
  ▼
RateLimit 中间件（Redis INCR 限流）
  │
  ▼
Handler: 绑定参数 → 调 Service
  │
  ▼
Service: StreamChat
  │   ├─ 1. 幂等检查（DB）
  │   ├─ 2. 拿/建会话（DB）
  │   ├─ 3. 拼 prompt（System + History + RAG 检索）
  │   ├─ 4. Agent.Run（带超时 / 重试 / 降级）
  │   │     ├─ LLM.Chat（带 Retry）
  │   │     ├─ 工具执行（带单工具超时）
  │   │     └─ 失败 → Fallback 文案
  │   ├─ 5. 落库（user/assistant 消息）
  │   ├─ 6. 计费打点（Redis HIncr）
  │   └─ 7. 幂等标记完成
  │
  ▼
Handler: SSE 输出
  │
  ▼
客户端
```

## 7. 测试场景

### 7.1 测试超时降级

```bash
# 把 LLM 切 mock，stream chunk 改 1s，故意触发超时
# 修改 .env：
LLM_MOCK_STREAM_CHUNK=200ms
LLM_REQUEST_TIMEOUT=2s

# 发个长文本测试
curl -N -X POST http://localhost:8080/v1/chat/stream \
  -H "Content-Type: application/json" \
  -d '{"messages":[{"role":"user","content":"讲个超长故事"}]}'
```

预期看到：前几秒有 chunk 输出，然后降级文案「抱歉，我这边有点忙……」。

### 7.2 测试账单

```bash
# 发 5 个请求
for i in 1 2 3 4 5; do
  curl -X POST http://localhost:8080/v1/chat \
    -H "Content-Type: application/json" \
    -H "X-User-ID: user_001" \
    -d "{\"messages\":[{\"role\":\"user\",\"content\":\"test $i\"}]}"
done

# 查询今日账单
curl http://localhost:8080/v1/admin/billing?user_id=user_001
```

预期：
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "completion_tokens": "250",
    "cost_usd": "0.00015",
    "prompt_tokens": "500"
  }
}
```

### 7.3 测试重试

```bash
# 在 mock 客���端里加一个概率失败
# (Day 3 的 fakeAnswer 里 10% 概率返回 error)
# 然后发 100 个请求，看日志里有没有 "retry" 字样
```

## 8. Day 4 复盘清单

| 检查点 | 完成 |
|-------|------|
| Agent 能跑通多轮工具调用 | ☐ |
| 超时能命中并降级 | ☐ |
| 重试逻辑生效（看日志） | ☐ |
| 账单查询能查到数据 | ☐ |
| trace_id 全链路串联 | ☐ |
| panic 能被 recovery 捕获 | ☐ |
| 服务挂了一个接口不整站挂 | ☐ |

## 📝 Day 4 作业

### 作业 1：给 Agent 加一个"决策日志"

```go
// 每次决策都打日志
logger.Info("agent decision",
    zap.Int("turn", turn),
    zap.String("decision", "use_tool"),  // 或 "final_answer"
    zap.String("tool", tc.Name),
    zap.String("reason", "用户问的是支付问题，需要查订单"),
)
```

目的：面试时能演示"Agent 决策过程"，对面试官特别有说服力。

### 作业 2：实现一个"模型不可用"开关

```go
// Redis 里一个 flag，运维可以手动切
// key: feature:llm_disabled
// 值为 "true" 时直接返回降级文案，不调模型
func (s *ChatService) StreamChat(...) {
    disabled, _ := s.rdb.Get(ctx, "feature:llm_disabled").Result()
    if disabled == "true" {
        fb := PickFallback()
        onChunk(fb)
        return &ChatResponse{Content: fb}, nil
    }
    // ... 正常流程
}
```

测试：`redis-cli SET feature:llm_disabled true`，发请求看降级。

### 作业 3：加一个每日预算超限保护

```go
// 单用户单日预算上限（默认 5 美元）
const DailyBudgetLimit = 5.0

func (s *ChatService) checkBudget(ctx context.Context, userID string) error {
    data, _ := s.billingSvc.Query(ctx, userID)
    cost, _ := strconv.ParseFloat(data["cost_usd"], 64)
    if cost >= DailyBudgetLimit {
        return errors.New("今日调用预算已用完，请明天再来")
    }
    return nil
}
```

## 💡 重点提示

1. **超时是必须的**：不设超时，模型卡住整个连接就废了，账单也会爆
2. **重试要看错误类型**：参数错误重试 100 次也没用
4. **降级要诚实**：前端要知道"现在是兜底回答"，别假装模型在工作
5. **账单数据要分开存**：用户粒度 + 模型粒度 + 时间粒度，运营要看
6. **结构化日志 = debug 神器**：JSON 日志 + trace_id，能串起整条链路

## ⏭️ Day 5 预告

明天把整个服务 Docker 化，加一个 mysql + redis 的 docker-compose。给你的项目写 README。然后扫一遍 PHP 语法（能看懂 Laravel 路由就行）。最后整理一份"入职第一周要问的问题清单"。

---

**💪 Day 4 跑完，你的服务就有"工业级"的雏形了。**