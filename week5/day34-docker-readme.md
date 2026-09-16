# 第34课（Week5 Day 5）：Docker + 读别人代码 + 给自己的项目写 README

> 学习时间：6~8 小时 | 难度：⭐⭐⭐
> **日期**：9 月 16 日（入职前一天）
> **目标**：服务能 `docker compose up` 一键起，README 能让同事 3 分钟看懂项目

## 🎯 交卷标准

- [ ] `docker compose up` 能把 app + mysql + redis 一次性起来
- [ ] 容器之间能通信（app 访问到 mysql/redis）
- [ ] `.env` 切换真模型能跑通（在容器内）
- [ ] README.md 完整：怎么跑、目录结构、接口列表、环境变量、常见问题
- [ ] 扫一遍 PHP 基础语法，能看懂 Laravel 路由
- [ ] 整理一份"入职第一周 10 个问题"清单
- [ ] 17 号早上能问出有质量的问题，不是"在吗"

---

## 📋 本课目标

- 巩固 Docker 多阶段构建 + Compose（[week4/day23](../week4/day23-docker.md)）
- 学会写一份能交付的 README
- 快速入门 PHP 语法（看懂就行）
- 整理入职第一周的提问清单

## 1. Dockerfile（多阶段构建）

> 直接复用 [week4/day23](../week4/day23-docker.md) 的模板，适配 chat-service。

### chat-service/Dockerfile

```dockerfile
# ===== 阶段1：构建 =====
FROM golang:1.22-alpine AS builder

WORKDIR /build

# 国内镜像加速
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# 先 COPY go.mod 利用 Docker 缓存
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# 构建二进制，-ldflags="-s -w" 去调试信息减小体积
RUN go build -ldflags="-s -w" -o /build/chat-service ./cmd/server

# ===== 阶段2：运行 =====
FROM alpine:3.19

# 装 ca-certificates（如果调 HTTPS API）
RUN apk add --no-cache ca-certificates tzdata

ENV TZ=Asia/Shanghai

# 非 root 用户
RUN adduser -D -g '' appuser
WORKDIR /app

COPY --from=builder /build/chat-service /app/chat-service

USER appuser

EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app/chat-service"]
```

### .dockerignore

```
.git
.gitignore
.env
.env.local
*.md
.vscode
.idea
logs/
tmp/
*_test.go
Dockerfile
docker-compose.yml
```

## 2. docker-compose.yml

```yaml
version: '3.8'

services:
  # ===== 主应用 =====
  app:
    build:
      context: .
      dockerfile: Dockerfile
    image: chat-service:latest
    container_name: chat-service
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      # Server
      - SERVER_ADDR=:8080
      - LOG_LEVEL=info
      # LLM（默认 mock，进组后改 .env 切真接口）
      - LLM_MOCK_MODE=true
      - LLM_BASE_URL=${LLM_BASE_URL:-https://api.openai.com/v1}
      - LLM_API_KEY=${LLM_API_KEY:-}
      - LLM_MODEL=${LLM_MODEL:-gpt-4o-mini}
      - LLM_MAX_TOKENS=1024
      - LLM_TEMPERATURE=0.7
      - LLM_REQUEST_TIMEOUT=30s
      - LLM_PRICE_INPUT=${LLM_PRICE_INPUT:-0.00015}
      - LLM_PRICE_OUTPUT=${LLM_PRICE_OUTPUT:-0.0006}
      # MySQL
      - MYSQL_DSN=root:root123@tcp(mysql:3306)/chat_service?charset=utf8mb4&parseTime=True&loc=Local
      - MYSQL_MAX_OPEN_CONNS=50
      - MYSQL_MAX_IDLE_CONNS=10
      - MYSQL_CONN_MAX_LIFETIME=30m
      # Redis
      - REDIS_ADDR=redis:6379
      - REDIS_PASSWORD=
      - REDIS_DB=0
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_healthy
    networks:
      - chat-net

  # ===== MySQL =====
  mysql:
    image: mysql:8.0
    container_name: chat-mysql
    restart: unless-stopped
    environment:
      - MYSQL_ROOT_PASSWORD=root123
      - MYSQL_DATABASE=chat_service
      - TZ=Asia/Shanghai
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./migrations:/docker-entrypoint-initdb.d:ro
    command: --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-uroot", "-proot123"]
      interval: 10s
      timeout: 5s
      retries: 5
    networks:
      - chat-net

  # ===== Redis =====
  redis:
    image: redis:7-alpine
    container_name: chat-redis
    restart: unless-stopped
    command: redis-server --appendonly yes --requirepass ""
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 5
    networks:
      - chat-net

volumes:
  mysql_data:
  redis_data:

networks:
  chat-net:
    driver: bridge
```

### 启动命令

```bash
# 一键起
docker compose up -d

# 看日志
docker compose logs -f app

# 跑迁移（如果有 migrate 工具）
docker compose exec app sh -c "echo 'migrations done'"

# 停
docker compose down

# 停 + 删数据（危险）
docker compose down -v
```

### 验证

```bash
# 1. 看 health
curl http://localhost:8080/healthz
# {"status":"ok","time":1726234567}

# 2. 发请求
curl -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -H "X-User-ID: docker_test" \
  -d '{"messages":[{"role":"user","content":"docker 起来了"}]}'
```

## 3. .env.example（最终版）

```bash
# ===== Server =====
SERVER_ADDR=:8080
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=60s
LOG_LEVEL=info

# ===== LLM =====
# 切真接口时改 false
LLM_MOCK_MODE=true
LLM_BASE_URL=https://api.openai.com/v1
LLM_API_KEY=
LLM_MODEL=gpt-4o-mini
LLM_MAX_TOKENS=1024
LLM_TEMPERATURE=0.7
LLM_REQUEST_TIMEOUT=30s
LLM_MOCK_STREAM_CHUNK=30ms
# 单价（美元 / 1k tokens）
LLM_PRICE_INPUT=0.00015
LLM_PRICE_OUTPUT=0.0006

# ===== MySQL =====
MYSQL_DSN=root:root123@tcp(127.0.0.1:3306)/chat_service?charset=utf8mb4&parseTime=True&loc=Local
MYSQL_MAX_OPEN_CONNS=50
MYSQL_MAX_IDLE_CONNS=10
MYSQL_CONN_MAX_LIFETIME=30m

# ===== Redis =====
REDIS_ADDR=127.0.0.1:6379
REDIS_PASSWORD=
REDIS_DB=0
```

## 4. README.md（交付级别）

> 🎯 **这份 README 就是你入职第一周的工作说明书**，同事照着 3 分钟能上手。

### chat-service/README.md

```markdown
# chat-service

> LLM 聊天后端服务（Gin + MySQL + Redis + SSE）

## 📦 包含什么

- HTTP 接口（REST + SSE）
- MySQL 存储会话和消息
- Redis 做会话缓存 + 用户限流 + 账单
- LLM 客户端（OpenAI / Anthropic / Mock 三选一）
- Agent 编排（Function Calling + 多轮工具）
- 最小 RAG（内存向量 + cosine 相似度，生产应换 pgvector）
- 结构化日志（zap）+ trace_id 全链路
- 单用户账单 / 模型成本打点

## 🚀 快速开始

### 方式一：Docker Compose（推荐）

```bash
# 1. 克隆
git clone <repo>
cd chat-service

# 2. 配环境变量
cp .env.example .env
# 编辑 .env，至少把 LLM_API_KEY 填上（或者保持 MOCK_MODE=true）

# 3. 起服务
docker compose up -d

# 4. 验证
curl http://localhost:8080/healthz
# {"status":"ok","time":1726234567}

# 5. 发请求
curl -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -H "X-User-ID: test_user" \
  -d '{"messages":[{"role":"user","content":"你好"}]}'
```

### 方式二：本地 Go 跑（开发用）

```bash
# 1. 起 MySQL + Redis（用 Docker 或本地装）
docker run -d -p 3306:3306 -e MYSQL_ROOT_PASSWORD=root123 -e MYSQL_DATABASE=chat_service mysql:8.0
docker run -d -p 6379:6379 redis:7-alpine

# 2. 跑迁移
mysql -h 127.0.0.1 -u root -proot123 chat_service < migrations/001_init.sql

# 3. 启动服务
cp .env.example .env
go run cmd/server/main.go
```

## 🗂 目录结构

```
chat-service/
├── cmd/server/             # 入口
├── internal/
│   ├── handler/            # HTTP handler（解析参数、返回响应）
│   ├── service/            # 业务逻辑（聊天、LLM、Agent、RAG、账单）
│   ├── repository/         # 数据访问（GORM）
│   ├── model/              # 数据库模型
│   ├── middleware/         # Gin 中间件（日志、限流、鉴权、recovery）
│   ├── infra/              # 基础设施（MySQL、Redis 初始化）
│   └── config/             # 配置加载
├── migrations/             # SQL 迁移脚本
├── docker-compose.yml
├── Dockerfile
├── .env.example
└── README.md
```

## 🌐 接口列表

### Chat

| Method | Path | 说明 | 是否需要鉴权 |
|--------|------|------|------------|
| POST | `/v1/chat` | 非流式聊天 | 是 |
| POST | `/v1/chat/stream` | 流式聊天（SSE） | 是 |
| GET | `/v1/conversations` | 当前用户的会话列表 | 是 |
| GET | `/v1/conversations/:id/messages` | 会话消息列表 | 是 |

### Admin

| Method | Path | 说明 |
|--------|------|------|
| GET | `/v1/admin/billing?user_id=xxx` | 查询用户今日账单 |

### System

| Method | Path | 说明 |
|--------|------|------|
| GET | `/healthz` | 健康检查（K8s liveness probe） |

## 🔧 环境变量

完整列表见 [.env.example](.env.example)。

**必填项**：
- `LLM_API_KEY`：调真模型时必填（mock 模式不用）
- `LLM_BASE_URL`：模型 API 地址（公司网关 / OpenAI / Anthropic）
- `MYSQL_DSN`：MySQL 连接串
- `REDIS_ADDR`：Redis 地址

## 🛠 接口示例

### 1. 流式聊天

```bash
curl -N -X POST http://localhost:8080/v1/chat/stream \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_001" \
  -H "X-Trace-ID: trace_xxx" \
  -d '{
    "messages": [
      {"role": "user", "content": "你好"}
    ]
  }'
```

**响应**（SSE 格式，每行一个事件）：

```
data: {"type":"chunk","content":"你"}

data: {"type":"chunk","content":"好"}

...

event: done
data: {"type":"done","usage":{"prompt_tokens":12,"completion_tokens":3}}
```

### 2. 多轮对话

```bash
# 第一轮
curl -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_001" \
  -d '{
    "messages": [{"role":"user","content":"我叫叶子"}]
  }'
# 响应里带 conversation_id

# 第二轮
curl -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_001" \
  -d '{
    "conversation_id": "conv_xxx",
    "messages": [{"role":"user","content":"我叫什么？"}]
  }'
```

### 3. 幂等

```bash
# 相同 idempotency_key 重复请求直接返回上次结果
curl -X POST http://localhost:8080/v1/chat \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user_001" \
  -d '{
    "idempotency_key": "idem_xxx",
    "messages": [{"role":"user","content":"测试"}]
  }'
```

## 🧱 架构图

```
        ┌────────────────────────────────────────┐
        │          Chat Service (Go/Gin)         │
        └────────────────────────────────────────┘
                  │              │           │
       ┌──────────┘              │           └──────────┐
       ▼                         ▼                      ▼
   ┌────────┐             ┌──────────────┐      ┌──────────────┐
   │ MySQL  │             │    Redis     │      │   LLM API    │
   │(会话/消│             │ (缓存/限流/   │      │ (OpenAI/     │
   │  息)   │             │   账单)        │      │  Anthropic)  │
   └────────┘             └──────────────┘      └──────────────┘
```

## 🐛 常见问题

### Q: 流式响应卡住 / 收不到完整响应？

检查 Nginx 反向代理是否开启了缓冲。加 `proxy_buffering off;` 或在响应里设置 `X-Accel-Buffering: no`（代码里已加）。

### Q: 模型调用超时？

调整 `LLM_REQUEST_TIMEOUT`，默认 30s。如果模型真的很慢，可以加一个二级缓存层（Redis 缓存最近回答）。

### Q: 账单查询不到？

1. 检查 Redis 连接（`docker compose logs redis`）
2. 检查 `LLM_MOCK_MODE` 是否为 false（mock 也会计费但价格不准）
3. 账单按日期分 key，时区用 `Asia/Shanghai`

### Q: 进组后发现环境变量命名不一样？

公司项目通常用 `viper` 或 `configmap`，把 `internal/config/config.go` 的 `Load()` 函数替换掉即可，业务层不用动。

## 📚 相关文档

- [Gin 官方文档](https://gin-gonic.com/)
- [GORM 文档](https://gorm.io/)
- [go-redis 文档](https://redis.uptrace.dev/)
- [OpenAI API](https://platform.openai.com/docs)
- [Anthropic API](https://docs.anthropic.com/)
- [Go OpenAI SDK](https://github.com/sashabaranov/go-openai)

## 📝 License

MIT
```

## 5. ARCHITECTURE.md（架构说明）

```markdown
# 架构说明

## 1. 分层

- **Handler**：HTTP 层，解析参数、调用 Service、返回响应
- **Service**：业务编排，调 LLM、组装 prompt、管理会话
- **Repository**：数据访问层，对 GORM 封装
- **Model**：数据模型定义

调用方向：`Handler → Service → Repository`，**禁止反向调用**。

## 2. 关键流程：流式聊天

```
1. 用户 POST /v1/chat/stream
2. 中间件链：日志 → 鉴权 → 限流
3. Handler 绑定 JSON 请求体
4. Service.StreamChat:
   a. 幂等检查（DB）
   b. GetOrCreate 会话（DB）
   c. 读历史消息（Redis 缓存 → 兜底查 DB）
   d. 拼 prompt：system + history + RAG 检索结果
   e. Agent.Run（带超时 / 重试 / 降级）
      - LLM.Chat 调模型（带 Retry）
      - 如需工具 → 执行 → 回填 → 继续循环
   f. 落库（user + assistant 消息）
   g. 账单打点（Redis HIncrBy）
   h. 标记幂等完成
5. Service 通过 onChunk 回调吐字给 Handler
6. Handler 转 SSE 协议 → 写回客户端
```

## 3. 接口设计原则

### LLMClient 接口

```go
type LLMClient interface {
    Chat(ctx, req) (*ChatResponse, error)
    Stream(ctx, req, onChunk) (*ChatResponse, error)
}
```

- **接口先于实现**：今天 Mock，明天换 OpenAI，后天换 Anthropic，**业务层零改动**
- **ctx 必须穿透**：超时、取消、trace_id 全靠它

### Repository 接口

- 每个表一个 Repo，方法命名 Get/Create/List/Update/Delete
- ctx 必传，不传视为代码 bug
- 复杂查询写在 Repo 私有方法，业务层不写 SQL

## 4. 配置加载

- 用 `godotenv` 加载 `.env`
- 所有配置在 `internal/config/config.go` 集中定义
- 业务层只能依赖 `*config.Config`，不直接读环境变量

## 5. 错误处理

业务错误一律 `*handler.BusinessError`，含 Code + Message：

```go
return nil, handler.BadRequest("参数错误")
return nil, handler.NewBizError(handler.CodeUpstreamFailed, "模型挂了", err)
```

Handler 层统一转 HTTP 响应，**业务错误一律返回 200，错误码在 body 里**。

## 6. 日志规范

- 用 zap，禁止 `fmt.Println` / `log.Println`
- 必带字段：`trace_id`、`user_id`、`path`、`latency`
- 业务日志用 `middleware.LoggerFromCtx(ctx)` 取带字段的 logger

## 7. 部署清单

生产部署前必做：

- [ ] `GIN_MODE=release`
- [ ] `LOG_LEVEL=info` 或 `warn`
- [ ] 真实模型 API key 放进 Secret Manager（不入库）
- [ ] MySQL / Redis 加白名单
- [ ] 反向代理（Nginx）关闭 SSE 缓冲
- [ ] 健康检查 `/healthz` 接 K8s probe
- [ ] 日志接 ELK / Loki
- [ ] 监控告警：错误率、延迟、token 成本
```

## 6. PHP 扫盲（看懂 Laravel 路由就行）

> 你的岗位 JD 提了 PHP 存量。入职后大概率要看懂 Laravel 项目的路由和 Controller，能改 prompt 不动框架。

### 6.1 30 分钟速览

**Laravel 路由**：

```php
// routes/web.php
Route::get('/users', [UserController::class, 'index']);
Route::post('/users', [UserController::class, 'store']);
Route::put('/users/{id}', [UserController::class, 'update']);
Route::delete('/users/{id}', [UserController::class, 'destroy']);

// API 路由组
Route::middleware('auth:sanctum')->prefix('api/v1')->group(function () {
    Route::post('/chat', [ChatController::class, 'chat']);
});
```

**Laravel Controller**：

```php
namespace App\Http\Controllers;

use Illuminate\Http\Request;

class ChatController extends Controller
{
    public function chat(Request $request)
    {
        $userId = $request->header('X-User-ID');
        $messages = $request->input('messages');

        // 调服务层
        $reply = app(ChatService::class)->handle($userId, $messages);

        return response()->json([
            'code' => 0,
            'message' => 'success',
            'data' => ['content' => $reply],
        ]);
    }
}
```

**Laravel 中间件**：

```php
namespace App\Http\Middleware;

use Closure;

class TraceIdMiddleware
{
    public function handle($request, Closure $next)
    {
        $traceId = $request->header('X-Trace-ID') ?? Str::uuid();
        $request->attributes->set('trace_id', $traceId);
        $response = $next($request);
        $response->headers->set('X-Trace-ID', $traceId);
        return $response;
    }
}
```

> 💡 **看懂就够**：Laravel 的「路由 + Controller + 中间件」三层和 Gin 是一回事，名字不一样而已。

### 6.2 速成对比表

| 概念 | Laravel (PHP) | Gin (Go) |
|------|---------------|----------|
| 路由 | `Route::post('/path', ...)` | `r.POST("/path", ...)` |
| Controller | 类 + 方法 | handler 函数 |
| 中间件 | `Route::middleware()` | `r.Use()` |
| 请求参数 | `$request->input('xxx')` | `c.ShouldBindJSON(&req)` |
| 返回 JSON | `response()->json([...])` | `c.JSON(200, ...)` |
| 数据库 ORM | Eloquent | GORM |
| 缓存 | Redis facade | go-redis |
| 配置 | `.env` + `config()` | godotenv + os.Getenv |

## 7. 入职第一周 10 个问题清单

> 🎯 **这 10 个问题问完，你第一周就能干活**。

### 关于仓库和工程
1. **仓库怎么起？** 依赖怎么装？是 monorepo 还是多仓？
2. **环境变量在哪？** 是 K8s ConfigMap 还是 Vault？谁能给？
3. **本地开发怎么连数据库？** 用 Docker compose 还是直连测试库？

### 关于业务
4. **核心业务流是什么？** 用户进来会触发哪些动作？
5. **调大模型的那一层是哪个 package？** 哪个文件？
6. **SSE 长连接怎么写的？** 断线重连怎么做的？

### 关于质量
7. **错误码怎么设计？** 前端怎么识别错误？
8. **日志在哪看？** 是 stdout 还是 ELK？trace_id 怎么传？
9. **单测覆盖率门槛多少？** 集成测试怎么跑？

### 关于人和流程
10. **我第一周领什么活？** （主动要"改 prompt、加日志、补一个 tool"这种小票）

## 8. 模拟入职练习

### 8.1 读一遍你的 README

让一个朋友（或自己模拟）从头读 [README.md](./README.md)，回答：

- 项目是干啥的？（30 秒能说清吗）
- 怎么起服务？（5 分钟能跑起来吗）
- 怎么调通模型？（能照着做吗）

如果任何一个环节卡壳，回去改 README。

### 8.2 给自己面试一次

照着 [week1/sexAI-interview20260908.md](../week1/sexAI-interview20260908.md) 那份面试问题列表，自己回答 30 分钟。重点回答：

- 「多 Agent 怎么设计？」—— 用你这周写的 Agent.Run 来讲
- 「RAG 流程是什么？」—— 用 rag.go + tools/faq.go 来讲
- 「Go error 处理？」—— 用 Service 层的 error 链来讲
- 「MySQL 主从？」—— 老实说不熟，承诺两周内补

### 8.3 准备 5 个你打算问同事的问题

**好的问题示例**：
- 「我看 ChatService 里调用了 Client 层的 Generate，请问这个 client 是公司自己封装的还是用了开源 SDK？」
- 「数据库里 messages 表已经有 1000w 条了，分页查询时你们一般用什么策略？」
- 「我看日志里有些奇怪的 trace_id 不是 UUID 格式，是不是有什么内部约定？」

**不好的问题示例**：
- ❌ 「这个项目是干啥的？」（自己读 README）
- ❌ 「K8s 怎么用？」（自己看文档）
- ❌ 「PHP 怎么写？」（自己学）

## 9. Day 5 复盘清单

| 检查点 | 完成 |
|-------|------|
| `docker compose up` 一次起 | ☐ |
| 容器内能调通 mock 模型 | ☐ |
| 容器内能调通真模型 | ☐ |
| README 完整且能照着跑 | ☐ |
| 10 个入职问题准备好 | ☐ |
| PHP 路由和 Controller 能看懂 | ☐ |
| 整个项目能在 5 分钟内讲清架构 | ☐ |

## 📝 Day 5 作业

### 作业 1：把 5 天学的写一篇 800 字总结

写到 `docs/REFLECTION.md`，包含：

1. 你学会了什么
2. 哪些还是不懂（**诚实列出来**，进组前看一遍避免翻车）
3. 入职第一周的目标

### 作业 2：跑一遍完整 demo 录屏

用 `curl` 或 Postman 录一段 3 分钟视频：

1. `docker compose up`
2. `curl /v1/chat`（mock 模式）
3. 改 `.env` 切真模型，再 `curl`
4. `curl /v1/admin/billing` 看账单

面试时拿这个视频当佐证，比讲 30 分钟更有说服力。

### 作业 3：给你的项目加 GitHub Actions

```yaml
# .github/workflows/ci.yml
name: CI
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: root123
          MYSQL_DATABASE: chat_service
        ports: ['3306:3306']
      redis:
        image: redis:7-alpine
        ports: ['6379:6379']
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go test ./... -race -coverprofile=coverage.out
      - run: go vet ./...
```

## 💡 重点提示

1. **README 是面试作品**：写完请朋友读一遍，能不能 3 分钟上手就知道质量
2. **.env.example 必须完整**：同事第一次跑项目，90% 卡在环境变量
3. **docker-compose 一定要能跑**：哪怕只在本地跑通，CI 上能跑就是加分
4. **诚实列不懂**：比硬编强 10 倍，面试官阅人无数一眼看穿
5. **5 天练的不是背题**：是"出问题知道往哪查"的能力

## ⏭️ 入职第一周预告

**Day 1**：看仓库、起本地、找"调大模型"那层、领第一个 ticket。
**Day 2~3**：改 prompt、加 tool、补日志。做出第一个 PR。
**Day 4~5**：求代码 review、问问题、摸清部署流程。
**周末**：写第一周周报。

> 记住：这 5 天练的不是让你"来了就能带节奏"，是让你"来了不慌"。
> 入职后前两个月的"边学边交"是另一个课题，但有了这 5 天的底子，你不会被它压垮。

---

**🎉 5 天结束。17 号见。**

**💪 接下来就是真刀真枪的职场了。**