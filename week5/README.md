# Week 5：5 天快速上手入职

> **目标岗位**：Golang + LLM 应用后端（成人视频站 + APP + Web3，Chat/SSE/Agent/RAG）
> **窗口**：9 月 12 日（今天）→ 9 月 16 日（Day 5）→ 9 月 17 日（入职）
> **每日时长**：6~8 小时
> **核心理念**：写得动，不是背得全；跑起来再说，文档能讲清每一层

## 🎯 这 5 天要交出的东西

5 天后，你的电脑里要有一个完整可演示的项目叫 `chat-service`（位置随便，建议放 `D:\webProject\go\chat-service`），含：

- 一个 Gin HTTP 服务，能 POST 聊天请求、SSE 流式返回
- 至少一张表（`conversations` + `messages`）把聊天记录落 MySQL
- Redis 做会话缓存 + 限流（按 user_id）
- 调用一个真实 LLM（OpenAI / Claude 兼容接口），代码能切 mock 模式
- 一个最小 RAG：本地几段文本 → 向量化（哪怕内存里硬编码相似度）→ 拼进 prompt
- 一个最小 Agent：用户问 → 判断要不要查工具 → 查 → 再回答
- Dockerfile + docker-compose，能 `docker compose up` 把 app + mysql + redis 起来
- 一个 README.md，3 分钟能讲清怎么跑、环境变量在哪、接口列表

## 📅 5 天日程

| Day | 文件 | 主题 | 每日交付 |
|-----|------|------|---------|
| 1 | [day30-gin-chat-sse.md](./day30-gin-chat-sse.md) | Gin + SSE + 调 LLM 假接口 | `POST /chat` SSE 跑通 |
| 2 | [day31-mysql-redis.md](./day31-mysql-redis.md) | MySQL + Redis + 工程习惯 | 聊天落库 + 缓存 + 限流 |
| 3 | [day32-llm-integration.md](./day32-llm-integration.md) | LLM 4 块（Chat / Prompt / Function Calling / SSE）+ 最小 RAG | 真调通一家 + 检索增强 demo |
| 4 | [day33-agent-stability.md](./day33-agent-stability.md) | Agent 编排 + 稳定性（超时/重试/降级/账单日志） | 一个最小 Agent + 限流账单可查 |
| 5 | [day34-docker-readme.md](./day34-docker-readme.md) | Docker + 读别人代码 + 给自己的项目写 README | 一键起项目 + 一页 README |

## 🛡️ 总原则

- **每天的交卷标准写在 Day 文件最上面**，没达到不要往下走
- **能跑通 > 代码漂亮**，面试会问的是"你写的哪一行在干什么"
- **遇到不会的立刻 Google/Bing**，比硬背强；可以参考的官方文档：[Gin](https://gin-gonic.com/)、[GORM](https://gorm.io/)、[go-redis](https://redis.uptrace.dev/)、[Anthropic API](https://docs.anthropic.com/)、[OpenAI API](https://platform.openai.com/docs)
- **17 号入职才需要会 PHP/K8s/Go-Zero**，这 5 天一个都不碰
- **通宵会废 17 号**，每天 6~8 小时够用，多出来的就改 README

## 📚 前置知识链接

如果某个基础概念忘了，直接翻回去看：

- Go 基础语法（变量、函数、错误处理）：[week1/day1-day5](../week1/)
- 结构体 / 接口 / 指针：[week1/day4-day5](../week1/)
- Slice / Map：[week1/day6](../week1/day6-collections.md)
- Goroutine / Channel / Context：[week1/day7](../week1/day7-concurrency.md)
- Gin 路由 / 中间件：[week2/day8-day9](../week2/)
- GORM：[week2/day10](../week2/day10-gorm.md)
- JWT：[week2/day11](../week2/day11-jwt.md)
- 配置 + 日志：[week2/day12](../week2/day12-config-log.md)
- Redis 缓存：[week3/day19](../week3/day19-redis-mq.md)
- 单元测试：[week4/day22](../week4/day22-testing.md)
- Docker：[week4/day23](../week4/day23-docker.md)

## 📂 最终目录结构

```
chat-service/
├── cmd/
│   └── server/
│       └── main.go              # 启动入口
├── internal/
│   ├── handler/                 # HTTP handler 层
│   │   └── chat.go
│   ├── service/                 # 业务逻辑层
│   │   ├── chat.go
│   │   ├── llm.go               # 调大模型
│   │   ├── agent.go             # Agent 编排
│   │   └── rag.go               # RAG 检索
│   ├── repository/              # 数据访问层
│   │   ├── conversation.go
│   │   └── message.go
│   ├── model/                   # 数据模型
│   │   └── chat.go
│   ├── middleware/              # 中间件
│   │   ├── auth.go
│   │   ├── ratelimit.go
│   │   └── trace.go
│   └── config/                  # 配置
│       └── config.go
├── migrations/                  # SQL 迁移
│   └── 001_init.sql
├── docs/
│   ├── README.md                # 怎么跑、接口列表、环境变量
│   └── ARCHITECTURE.md          # 架构图、各层职责
├── scripts/
│   └── seed.sql
├── .env.example
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

## ✅ 17 号早上能问同事的问题清单

入职第一周比"展示会多少框架"有用的事：

1. 仓库怎么起？依赖怎么装？
2. 环境变量在哪？谁能给？
3. 调大模型的那一层是哪个 package？
4. 错误码怎么返回？前端怎么识别？
5. SSE 长连接怎么写的？断线重连怎么做的？
6. 日志在哪里看？trace_id 怎么串？
7. 部署是 Docker 还是 K8s？怎么发版？

如果第 17 号你能直接照着自己的 README 找到答案，那这 5 天就没白练。

---

**💪 这周练的是「进组能干活」，不是「进组能带节奏」。先把 17 号第一周活过去，剩下两个月的「边学边交」是另一个课题。**
