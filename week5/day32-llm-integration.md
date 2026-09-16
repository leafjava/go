# 第32课（Week5 Day 3）：LLM 落地 —— Chat + Prompt + Function Calling + SSE + 最小 RAG

> 学习时间：6~8 小时 | 难度：⭐⭐⭐⭐
> **日期**：9 月 14 日
> **目标**：mock 切真模型，能跑通"提问 → 检索 → 带上下文回答"的最小 RAG

## 🎯 交卷标准

- [ ] `.env` 里 `LLM_MOCK_MODE=false` 后，真模型能调通（OpenAI 或 Claude 二选一）
- [ ] 多轮对话能记住前文（真模型 + 历史消息）
- [ ] Function Calling 能跑通：模型返回要调的工具，本地执行后回填再调用
- [ ] 最小 RAG 能跑通：本地几段文本 → 向量检索（哪怕内存版）→ 拼进 prompt → 模型带上下文回答
- [ ] SSE 流式输出仍然正常工作（真模型切流式）
- [ ] token 用量能在日志里看到，账单可查
- [ ] 能跟面试官 30 分钟讲清 LLM 应用层 4 大件

---

## 📋 本课目标

- 接入真实 LLM（OpenAI / Anthropic 兼容）
- 学会 Function Calling / Tool Use 的完整流程
- 学会最简 RAG：文档切分 + 向量化 + topK 检索 + 拼 prompt
- 理解 embedding 是什么、向量库（哪怕内存版）怎么用
- 为 Day 4 的 Agent 编排打基础

## 1. 选型：OpenAI 还是 Claude

| 维度 | OpenAI（gpt-4o-mini） | Anthropic（claude-haiku-4-5） |
|------|---------------------|-----------------------------|
| 价格 | 便宜 | 中等 |
| 速度 | 快 | 快 |
| 中文 | 良好 | 良好 |
| Function Calling | 稳定 | 稳定（叫 Tool Use） |
| 流式 | 原生支持 | 原生支持 |
| 文档 | 中文翻译好 | 英文为主 |
| Go SDK | 社区版（非官方） | 社区版（非官方） |

> 💡 **推荐路径**：用 [sashabaranov/go-openai](https://github.com/sashabaranov/go-openai) 这个**社区主流 SDK**，它兼容 OpenAI 接口。如果公司走 Anthropic，可以切到 [anthropic-sdk-go](https://github.com/anthropics/anthropic-sdk-go)（今天两者写法都演示，但只挑一个跑通即可）。

## 2. 接入 OpenAI（推荐先跑通这家）

### 2.1 安装

```bash
go get github.com/sashabaranov/go-openai@latest
```

### 2.2 配置文件更新

```bash
# .env
LLM_MOCK_MODE=false
LLM_BASE_URL=https://api.openai.com/v1
LLM_API_KEY=sk-xxx    # 从 https://platform.openai.com/api-keys 拿
LLM_MODEL=gpt-4o-mini
```

> 💡 **国内访问**：`LLM_BASE_URL` 可以指向代理网关（如 `https://api.openai-proxy.com/v1`），或公司自己部署的中转网关。

### 2.3 实现 OpenAI 客户端（internal/service/llm_openai.go）

```go
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/sashabaranov/go-openai"

	"chat-service/internal/config"
)

type openaiClient struct {
	cfg config.LLMConfig
	cli *openai.Client
}

func NewOpenAIClient(cfg config.LLMConfig) LLMClient {
	c := openai.NewClientWithConfig(openai.ClientConfig{
		APIKey:    cfg.APIKey,
		BaseURL:   cfg.BaseURL,
		OrgID:     "",
		HTTPClient: defaultHTTPClient(cfg.RequestTimeout),
	})
	return &openaiClient{cfg: cfg, cli: c}
}

func (o *openaiClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	msgs := toOpenAIMessages(req.Messages)
	resp, err := o.cli.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       pickModel(o.cfg, req.Model),
		Messages:    msgs,
		MaxTokens:   req.MaxTokens,
		Temperature: float32(req.Temperature),
	})
	if err != nil {
		return nil, fmt.Errorf("openai chat: %w", err)
	}
	if len(resp.Choices) == 0 {
		return nil, errors.New("no choices in response")
	}
	return &ChatResponse{
		Content: resp.Choices[0].Message.Content,
		Usage: Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}, nil
}

func (o *openaiClient) Stream(ctx context.Context, req ChatRequest, onChunk func(chunk string) error) (*ChatResponse, error) {
	msgs := toOpenAIMessages(req.Messages)
	stream, err := o.cli.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       pickModel(o.cfg, req.Model),
		Messages:    msgs,
		MaxTokens:   req.MaxTokens,
		Temperature: float32(req.Temperature),
		Stream:      true,
	})
	if err != nil {
		return nil, fmt.Errorf("openai stream: %w", err)
	}
	defer stream.Close()

	var full string
	for {
		resp, err := stream.Recv()
		if errors.Is(err, openai.ErrStreamClosed) {
			break
		}
		if err != nil {
			return nil, err
		}
		if len(resp.Choices) == 0 {
			continue
		}
		delta := resp.Choices[0].Delta.Content
		if delta == "" {
			continue
		}
		full += delta
		if onChunk != nil {
			if err := onChunk(delta); err != nil {
				return nil, err
			}
		}
	}

	// 注：流式 API 的 Usage 不一定返回（OpenAI 在最后一个 chunk 带）
	return &ChatResponse{
		Content: full,
		Usage:   Usage{PromptTokens: 0, CompletionTokens: 0, TotalTokens: 0},
	}, nil
}

func toOpenAIMessages(msgs []Message) []openai.ChatCompletionMessage {
	out := make([]openai.ChatCompletionMessage, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, openai.ChatCompletionMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}
	return out
}

func pickModel(cfg config.LLMConfig, override string) string {
	if override != "" {
		return override
	}
	return cfg.Model
}

func defaultHTTPClient(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}
```

### 2.4 工厂方法切换

修改 `NewLLMClient`：

```go
func NewLLMClient(cfg *config.Config) LLMClient {
	if cfg.LLM.MockMode {
		return NewMockClient(cfg.LLM)
	}
	// 按 base URL 选 provider
	if strings.Contains(cfg.LLM.BaseURL, "anthropic") {
		return NewAnthropicClient(cfg.LLM)
	}
	return NewOpenAIClient(cfg.LLM)
}
```

> 💡 这就是 Day 1 设计的回报——`LLMClient` 接口在前面，现在切换 provider **业务层一行不改**。

## 3. Anthropic（Claude）实现（可选）

如果你入职的是 Anthropic 兼容场景，参考这份。**先跑通一家，另一家对照改即可**。

```go
// internal/service/llm_anthropic.go
package service

import (
	"context"
	"errors"

	"github.com/anthropics/anthropic-sdk-go"
)

type anthropicClient struct {
	cfg config.LLMConfig
	cli anthropic.Client
}

func NewAnthropicClient(cfg config.LLMConfig) LLMClient {
	c := anthropic.NewClient(
		anthropic.WithAPIKey(cfg.APIKey),
		// anthropic.WithBaseURL(cfg.BaseURL), // 国内代理
	)
	return &anthropicClient{cfg: cfg, cli: c}
}

func (a *anthropicClient) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	resp, err := a.cli.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: int64(req.MaxTokens),
		Messages:  toAnthropicMessages(req.Messages),
	})
	if err != nil {
		return nil, err
	}
	content := ""
	for _, b := range resp.Content {
		if b.Type == "text" {
			content += b.Text
		}
	}
	return &ChatResponse{
		Content: content,
		Usage: Usage{
			PromptTokens:     int(resp.Usage.InputTokens),
			CompletionTokens: int(resp.Usage.OutputTokens),
			TotalTokens:      int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
		},
	}, nil
}

func (a *anthropicClient) Stream(ctx context.Context, req ChatRequest, onChunk func(string) error) (*ChatResponse, error) {
	stream := a.cli.Messages.NewStreaming(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaudeHaiku4_5,
		MaxTokens: int64(req.MaxTokens),
		Messages:  toAnthropicMessages(req.Messages),
	})
	defer stream.Close()

	var full string
	for stream.Next() {
		event := stream.Current()
		if event.Type == "content_block_delta" {
			text := event.Delta.Text
			full += text
			if onChunk != nil {
				if err := onChunk(text); err != nil {
					return nil, err
				}
			}
		}
	}
	if err := stream.Err(); err != nil {
		return nil, err
	}
	return &ChatResponse{Content: full}, nil
}

func toAnthropicMessages(msgs []Message) []anthropic.MessageParam {
	out := make([]anthropic.MessageParam, 0, len(msgs))
	for _, m := range msgs {
		if m.Role == "system" {
			continue // Anthropic system 单独传
		}
		out = append(out, anthropic.MessageParam{
			Role:    m.Role,
			Content: []anthropic.ContentBlock{{Type: "text", Text: m.Content}},
		})
	}
	return out
}
```

> ⚠️ Anthropic 的 system prompt 要单独处理（不能放在 messages 里），`messages` 只放 user/assistant。

## 4. Prompt 工程基础

### 4.1 三段式 Prompt 结构

```go
func buildSystemPrompt() string {
    return `你是「叶子助手」，一家成人视频网站 + APP 的智能客服。

【回答规则】
1. 只回答与产品相关的问题（会员、播放、支付、账号、内容分类）
2. 不知道就说不知道，不要编
3. 涉及支付、退款、账号安全，必须引导到人工客服
4. 回答简洁，控制在 100 字以内
5. 不要重复用户的问题

【当前可用工具】
- search_faq(query): 在 FAQ 知识库中检索
- get_user_orders(user_id): 查询用户订单

【回答格式】
直接给出答案，不要加 "你好"、"请问还有什么问题" 这种寒暄。`
}
```

### 4.2 用法（service 层组装）

```go
func (s *ChatService) buildMessages(userMsgs []Message) []Message {
    return append([]Message{
        {Role: "system", Content: buildSystemPrompt()},
    }, userMsgs...)
}
```

> 💡 **拆 system / user 是工业级标配**。System 放规则，User 放真实输入，分开调 token 统计更准，也方便运营改 prompt 不动业务。

### 4.3 进阶：Few-shot

```go
func buildSystemPrompt() string {
    return `你是叶子助手。回答示例：

示例 1:
用户: 怎么开通会员？
你: 进入【我的】-【会员中心】，选择套餐并支付即可。会员开通后立即生效。

示例 2:
用户: 视频加载不出来
你: 请尝试：1. 切换网络 2. 清除 APP 缓存 3. 升级到最新版 APP。如仍有问题请联系人工客服。

现在请回答用户的问题。`
}
```

## 5. Function Calling / Tool Use

> 🎯 **JD 里"Function Calling"对应的就是这块**。流程：
>
> ```
> 用户问 → 调模型 → 模型决定"需要调工具"
>                       ↓
>                  返回 tool_calls
>                       ↓
> Go 里执行工具（查 DB、调 API）
>                       ↓
> 把工具结果回填给模型
>                       ↓
> 模型基于工具结果生成最终回答
> ```

### 5.1 定义工具

```go
// internal/service/tools.go
package service

import (
	"context"
	"encoding/json"
	"fmt"
)

// Tool 工具定义
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"` // JSON Schema
	Handler     func(ctx context.Context, args json.RawMessage) (any, error)
}

// ToolRegistry 工具注册表
type ToolRegistry struct {
	tools map[string]*Tool
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{tools: make(map[string]*Tool)}
}

func (r *ToolRegistry) Register(t *Tool) {
	r.tools[t.Name] = t
}

func (r *ToolRegistry) Get(name string) (*Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

func (r *ToolRegistry) List() []*Tool {
	out := make([]*Tool, 0, len(r.tools))
	for _, t := range r.tools {
		out = append(out, t)
	}
	return out
}
```

### 5.2 注册 FAQ 搜索工具

```go
// internal/service/tools/faq.go
package service

import (
	"context"
	"encoding/json"
)

func RegisterFAQTool(reg *ToolRegistry, rag *RAGService) {
	reg.Register(&Tool{
		Name: "search_faq",
		Description: "在 FAQ 知识库中检索与用户问题相关的条目。返回最相关的 3 条。",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "用户的问题关键词",
				},
			},
			"required": []string{"query"},
		},
		Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
			var p struct {
				Query string `json:"query"`
			}
			if err := json.Unmarshal(args, &p); err != nil {
				return nil, err
			}
			hits, err := rag.Search(ctx, p.Query, 3)
			if err != nil {
				return nil, err
			}
			return hits, nil
		},
	})
}
```

### 5.3 调用流程（OpenAI）

```go
// internal/service/agent.go
package service

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/sashabaranov/go-openai"
)

// AgentRun 工具调用循环（最多 5 轮）
func AgentRun(
	ctx context.Context,
	cli *openai.Client,
	tools *ToolRegistry,
	model string,
	messages []openai.ChatCompletionMessage,
	maxTurns int,
	onChunk func(string) error,
) (string, error) {
	for turn := 0; turn < maxTurns; turn++ {
		// 1. 调模型
		resp, err := cli.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
			Tools:    toOpenAITools(tools.List()),
		})
		if err != nil {
			return "", err
		}
		if len(resp.Choices) == 0 {
			return "", errors.New("no choices")
		}
		msg := resp.Choices[0].Message

		// 2. 把模型回复加进消息历史
		messages = append(messages, msg)

		// 3. 没有工具调用 → 把 content 返回
		if len(msg.ToolCalls) == 0 {
			if msg.Content != "" && onChunk != nil {
				_ = onChunk(msg.Content)
			}
			return msg.Content, nil
		}

		// 4. 有工具调用 → 逐个执行
		for _, tc := range msg.ToolCalls {
			tool, ok := tools.Get(tc.Function.Name)
			if !ok {
				messages = append(messages, openai.ChatCompletionMessage{
					Role:       "tool",
					Content:    fmt.Sprintf("错误：未知工具 %s", tc.Function.Name),
					ToolCallID: tc.ID,
				})
				continue
			}

			result, err := tool.Handler(ctx, json.RawMessage(tc.Function.Arguments))
			resultJSON, _ := json.Marshal(map[string]any{"result": result})
			if err != nil {
				resultJSON, _ = json.Marshal(map[string]any{"error": err.Error()})
			}

			// 5. 工具结果回填
			messages = append(messages, openai.ChatCompletionMessage{
				Role:       "tool",
				Content:    string(resultJSON),
				ToolCallID: tc.ID,
			})
		}
	}
	return "", errors.New("max turns reached")
}

func toOpenAITools(tools []*Tool) []openai.Tool {
	out := make([]openai.Tool, 0, len(tools))
	for _, t := range tools {
		out = append(out, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}
	return out
}
```

> 💡 **关键点**：循环 + 工具回填是 Agent 核心逻辑。生产环境 Day 4 才加超时、重试、并发限制。

## 6. 最小 RAG

### 6.1 文档准备

```go
// internal/service/rag.go
package service

import (
	"context"
	"math"
	"sort"
	"strings"
	"sync"

	"chat-service/internal/config"
)

// Document 文档块
type Document struct {
	ID       string  `json:"id"`
	Content  string  `json:"content"`
	Source   string  `json:"source"` // 文件名 / URL
	Vector   []float32 `json:"-"`
}

// RAGService 最小 RAG（内存版，生产用向量库）
type RAGService struct {
	mu        sync.RWMutex
	documents []Document
	embedFn   func(ctx context.Context, text string) ([]float32, error)
}

func NewRAGService(cfg *config.Config) *RAGService {
	r := &RAGService{}
	r.loadDefaultDocs()
	return r
}

// loadDefaultDocs 加载默认 FAQ（生产从 DB / 文件读）
func (r *RAGService) loadDefaultDocs() {
	defaultDocs := []Document{
		{ID: "faq_001", Source: "faq.md", Content: "Q: 怎么开通会员？\nA: 进入【我的】-【会员中心】，选择套餐并支付。会员开通后立即生效。"},
		{ID: "faq_002", Source: "faq.md", Content: "Q: 视频加载不出来怎么办？\nA: 1. 检查网络 2. 清除 APP 缓存 3. 升级到最新版本 4. 切换 WiFi / 4G"},
		{ID: "faq_003", Source: "faq.md", Content: "Q: 怎么退款？\nA: 会员开通 7 天内未观看任何内容可全额退款，请联系人工客服。"},
		{ID: "faq_004", Source: "faq.md", Content: "Q: 账号被盗怎么办？\nA: 立即在【设置】-【账号安全】修改密码，并联系客服冻结账号。"},
		{ID: "faq_005", Source: "faq.md", Content: "Q: 视频支持下载吗？\nA: 付费会员支持离线下载，进入视频详情页点击下载按钮。"},
	}

	// 简单分块（按段落）+ 哈希做伪向量
	for _, d := range defaultDocs {
		chunks := splitByParagraph(d.Content)
		for i, c := range chunks {
			doc := Document{
				ID:      d.ID + "_" + itoa(i),
				Source:  d.Source,
				Content: c,
				Vector:  fakeEmbedding(c), // 真接 OpenAI Embedding 替换这里
			}
			r.documents = append(r.documents, doc)
		}
	}
}

// splitByParagraph 简单按段落切
func splitByParagraph(s string) []string {
	parts := strings.Split(s, "\n")
	out := make([]string, 0)
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// Search topK 检索
func (r *RAGService) Search(ctx context.Context, query string, topK int) ([]Document, error) {
	queryVec := fakeEmbedding(query)

	r.mu.RLock()
	defer r.mu.RUnlock()

	type scored struct {
		doc   Document
		score float64
	}
	scores := make([]scored, 0, len(r.documents))
	for _, d := range r.documents {
		scores = append(scores, scored{doc: d, score: cosineSim(queryVec, d.Vector)})
	}
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	out := make([]Document, 0, topK)
	for i := 0; i < topK && i < len(scores); i++ {
		out = append(out, scores[i].doc)
	}
	return out, nil
}

// ============ 假 Embedding（生产请换 OpenAI text-embedding-3-small）============
// 这里用文本哈希 + 字符统计做伪向量，**仅用于演示**。
// 真实场景必须用 embedding 模型，否则检索效果极差。

func fakeEmbedding(text string) []float32 {
	// 256 维向量
	vec := make([]float32, 256)
	// 简单的字符频率特征
	for i, r := range text {
		vec[int(r)%256] += 1
		vec[(i*7)%256] += 0.1
	}
	// 归一化
	var norm float32
	for _, v := range vec {
		norm += v * v
	}
	if norm > 0 {
		norm = float32(1.0 / math.Sqrt(float64(norm)))
	}
	for i := range vec {
		vec[i] *= norm
	}
	return vec
}

func cosineSim(a, b []float32) float64 {
	if len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i] * b[i])
		na += float64(a[i] * a[i])
		nb += float64(b[i] * b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	digits := "0123456789"
	out := ""
	for i > 0 {
		out = string(digits[i%10]) + out
		i /= 10
	}
	return out
}
```

### 6.2 真 Embedding（OpenAI）

```go
// 真接入时用这个替换 fakeEmbedding
func (r *RAGService) realEmbedding(ctx context.Context, text string) ([]float32, error) {
	resp, err := r.embedCli.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Model: openai.AdaEmbeddingV3, // text-embedding-3-small
		Input: []string{text},
	})
	if err != nil {
		return nil, err
	}
	return resp.Data[0].Embedding, nil
}
```

> ⚠️ **fakeEmbedding 是演示用，语义检索效果为 0**。真接入后：
>
> - 文档 embedding 提前算好（离线）
> - 用户 query embedding 实时算
> - 向量库用 pgvector / Milvus / Pinecone / Qdrant，**别自己造**

### 6.3 RAG 拼 Prompt

```go
func (s *RAGService) BuildContextPrompt(query string, topK int) (string, error) {
	docs, err := s.Search(context.Background(), query, topK)
	if err != nil {
		return "", err
	}
	if len(docs) == 0 {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString("【参考资料】\n")
	for i, d := range docs {
		fmt.Fprintf(&sb, "[%d] %s\n", i+1, d.Content)
	}
	sb.WriteString("\n请基于以上参考资料回答用户问题。如果资料里没有，直接说不知道，不要编。\n\n")
	sb.WriteString("【用户问题】\n")
	sb.WriteString(query)
	return sb.String(), nil
}
```

## 7. 把 RAG 接到 Service

```go
// chat.go 里改造
func (s *ChatService) StreamChat(ctx context.Context, req StreamRequest, onChunk func(string) error) (*ChatResponse, error) {
	// ... 前面的幂等、拿会话、省略

	// === 新增：用 RAG 检索增强 ===
	var userQuery string
	for _, m := range req.Messages {
		if m.Role == "user" {
			userQuery = m.Content
			break
		}
	}

	systemPrompt := buildSystemPrompt()
	if userQuery != "" && s.ragSvc != nil {
		ragCtx, err := s.ragSvc.BuildContextPrompt(userQuery, 3)
		if err == nil && ragCtx != "" {
			systemPrompt += "\n\n" + ragCtx
		}
	}

	// 拼消息：system + 历史 + 当前
	allMessages := []Message{{Role: "system", Content: systemPrompt}}
	for _, h := range history {
		allMessages = append(allMessages, Message{Role: h.Role, Content: h.Content})
	}
	allMessages = append(allMessages, req.Messages...)

	// 调 LLM
	streamReq := ChatRequest{Model: req.Model, Messages: allMessages, Stream: true}
	resp, err := s.llm.Stream(ctx, streamReq, func(chunk string) error {
		return onChunk(chunk)
	})
	// ... 落库逻辑同上
}
```

## 8. 启动 & 验证

```bash
# 编辑 .env 把 LLM_MOCK_MODE 改 false
LLM_MOCK_MODE=false
LLM_API_KEY=sk-xxx

# 启动
go run cmd/server/main.go
```

```bash
# 测试 RAG（不带 query 时模型也得知道自己是 FAQ 助手）
curl -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_001" \
  -d '{"messages":[{"role":"user","content":"怎么开通会员？"}]}'
```

预期回答应该提到「我的 → 会员中心 → 选择套餐」——这是 RAG 注入了 FAQ 内容的效果。

```bash
# 测试流式
curl -N -X POST http://localhost:8080/v1/chat/stream \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_001" \
  -d '{"messages":[{"role":"user","content":"视频加载不出来"}]}'
```

## 9. Day 3 复盘清单

| 检查点 | 完成 |
|-------|------|
| 真模型能调通 | ☐ |
| 多轮对话能记住前文 | ☐ |
| FAQ 检索能查到相关条目 | ☐ |
| 模型回答用到了检索内容 | ☐ |
| Function Calling 工具能调 | ☐ |
| 流式输出正常 | ☐ |
| 日志能看到 token 用量 | ☐ |

## 📝 Day 3 作业

### 作业 1：加第二个工具 `get_user_orders`

```go
// 内部写死 3 个 mock 订单，按 user_id 返回
func RegisterOrdersTool(reg *ToolRegistry) {
    reg.Register(&Tool{
        Name: "get_user_orders",
        Description: "查询当前用户的订单列表",
        Parameters: ...,
        Handler: func(ctx context.Context, args json.RawMessage) (any, error) {
            return []map[string]any{
                {"id": "ord_001", "amount": "98.00", "status": "paid"},
                {"id": "ord_002", "amount": "198.00", "status": "refunded"},
            }, nil
        },
    })
}
```

测试：「我有哪些订单？」→ 模型调工具 → 返回订单列表 → 生成自然语言回答。

### 作业 2：给 RAG 加 metadata 过滤

```go
type Document struct {
    ID       string
    Content  string
    Source   string
    Category string  // 新增：会员/支付/账号/技术
    Vector   []float32
}

// Search 支持 category 过滤
func (r *RAGService) Search(ctx context.Context, query string, topK int, category string) ([]Document, error) {
    // ...
}
```

### 作业 3：把 mock 改假向量换成真 embedding

如果有余力：
1. 加 `EMBEDDING_API_KEY` 配置
2. 把 `fakeEmbedding` 换成 OpenAI Embeddings 调用
3. 启动时一次性把所有文档 embedding 算完
4. 重新跑测试，看效果差别（应该会有惊喜）

## 💡 重点提示

1. **Function Calling 一定要带工具描述**：模型靠 description 决定要不要调，写不好它会乱调或漏调
2. **工具参数用 JSON Schema**：写错模型会按 schema 校验失败
3. **RAG 不是万能**：召回不准、加 context 浪费 token、prompt 越长越慢。**生产先验证不用 RAG 也能解决 80% 问题，再用 RAG 补剩下 20%**
4. **Embedding 不要存 JSON**：存二进制（pgvector、`bytea`）省空间
5. **模型选小不选大**：gpt-4o-mini / haiku 性价比高，先上线再优化

## ⏭️ Day 4 预告

明天在今天的基础上加**稳定性**：超时、重试、降级、限流账单日志。再把 Agent 编排做得更完整——能"思考要不要查工具、查几次、查不到怎么办"。

---

**💪 Day 3 是这周最难的一天，过了这一关你就有了一个"真能回答"的 AI 服务。**