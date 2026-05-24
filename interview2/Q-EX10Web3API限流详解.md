# Web3 API 限流详解：令牌桶 + 滑动窗口

---

## 前置知识：什么是限流？

限流（Rate Limiting）就是**限制"多快"**——比如规定每个用户**每秒最多请求 10 次**，超过就拒绝。

### 为什么 Web3 项目更需要限流？

```
Web3 后端面临的问题：

1. RPC 节点有频率限制
   Infura / Alchemy 的免费额度每秒只允许几十次请求
   一个用户刷页面就可能打爆

2. 提现/转账不能重复提交
   用户脚本连点三次"提现"→ 不能扣三次钱

3. 链上 Gas 竞争
   有人写脚本无限循环查 Gas 价格 → 等于 DDoS

4. 女巫攻击（Sybil Attack）
   一个攻击者用脚本注册 1000 个钱包刷空投
```

限流就是第一道防线。

---

## 一、令牌桶（Token Bucket）—— 轻量级方案

### 是什么？

想象一个桶，系统以固定速度往桶里放令牌：

```
      每秒放 10 个令牌
          ↓  ↓  ↓
    ┌─────────────────┐
    │  🎫🎫🎫🎫🎫   │  ← 桶最多装 20 个（满了就溢出）
    │  🎫🎫🎫🎫🎫   │
    └─────────────────┘
          ↓
    每个请求来，要拿走一个令牌才能通过
    桶里没令牌 → 请求被拒绝（返回 429）
```

**优点**：允许"突发"——平时没人用，桶里攒了 20 个令牌，突然来 20 个请求也能全通过。

### Go 实现：使用 x/time/rate 包

```go
package main

import (
    "fmt"
    "net/http"
    "sync"
    "time"

    "golang.org/x/time/rate"
)

// ============ 每用户一个限流器 ============

type RateLimiter struct {
    mu       sync.Mutex
    limiters map[string]*rate.Limiter // key: 用户ID/IP，value: 限流器
}

func NewRateLimiter() *RateLimiter {
    return &RateLimiter{
        limiters: make(map[string]*rate.Limiter),
    }
}

// getUserLimiter 获取或创建用户的限流器
func (rl *RateLimiter) getUserLimiter(userID string) *rate.Limiter {
    rl.mu.Lock()
    defer rl.mu.Unlock()

    limiter, exists := rl.limiters[userID]
    if !exists {
        // rate.NewLimiter(r, b)
        //   r = 每秒放多少个令牌（速率）
        //   b = 桶容量（允许的最大突发量）
        limiter = rate.NewLimiter(10, 20) // 每秒10个，桶容量20
        rl.limiters[userID] = limiter
    }
    return limiter
}

// ============ HTTP 中间件 ============

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 从请求中获取用户标识（实际项目从 JWT / API Key 取）
        userID := r.Header.Get("X-User-ID")
        if userID == "" {
            userID = r.RemoteAddr // 降级：用 IP 地址
        }

        limiter := rl.getUserLimiter(userID)

        // Allow() 尝试拿走一个令牌
        // 有令牌 → true，没令牌 → false
        if !limiter.Allow() {
            w.WriteHeader(http.StatusTooManyRequests) // 429
            w.Write([]byte(`{"error": "请求太频繁，请稍后再试"}`))
            return
        }

        next.ServeHTTP(w, r)
    })
}

// ============ 启动服务 ============

func main() {
    rl := NewRateLimiter()

    // 定期清理不活跃用户的限流器，防止内存泄漏
    go func() {
        for {
            time.Sleep(10 * time.Minute)
            rl.mu.Lock()
            for id, limiter := range rl.limiters {
                // 做一些清理逻辑...
                _ = id
                _ = limiter
            }
            rl.mu.Unlock()
        }
    }()

    mux := http.NewServeMux()
    mux.HandleFunc("/api/balance", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"balance": "1.5 ETH"}`))
    })

    fmt.Println("服务启动在 :8080")
    http.ListenAndServe(":8080", rl.Middleware(mux))
}
```

### 令牌桶参数怎么设？

```go
// Web3 常见场景的参数建议：

// 普通 API（查余额、查行情）
rate.NewLimiter(20, 40)  // 每秒 20 次，允许突发到 40

// 敏感接口（提现、转账）
rate.NewLimiter(1, 3)    // 每秒 1 次，最多突发 3 次

// RPC 代理（转发到 Infura/Alchemy）
rate.NewLimiter(50, 100) // 留余量，不超过上游限制
```

### 令牌桶的优缺点

```
✅ 优点：
  - 纯内存，极快（纳秒级）
  - 允许突发流量
  - 代码简单

❌ 缺点：
  - 重启丢失所有限流状态
  - 多实例部署时，每个实例独立计数
    → 实例A限流了10次，实例B也限流了10次
    → 实际用户发了 20 次
  - 不适合对精确度要求高的场景（如提现）
```

---

## 二、滑动窗口（Sliding Window）—— 精确型方案

### 是什么？

不再用"桶"和"令牌"，改用**时间窗口**：

```
固定窗口的问题：
  窗口：每分钟最多 10 次
  用户在 12:00:59 发了 10 次
  用户在 12:01:00 又发 10 次
  → 2 秒内发了 20 次，但没有违反"每分钟 10 次"的规则！

滑动窗口：
  窗口：最近 60 秒内最多 10 次
  任何时候都是看"过去 60 秒"的请求数
  → 完美解决边界问题
```

### Redis 实现：用 ZSET（有序集合）

```
ZSET 结构：
  key: rate_limit:user:123
  ┌─────────────────────────┐
  │  score (时间戳)  member  │
  ├─────────────────────────┤
  │  1700000000     req_1   │
  │  1700000001     req_2   │
  │  1700000002     req_3   │
  │  ...                    │
  └─────────────────────────┘

每次请求：
  1. ZREMRANGEBYSCORE 删除窗口外的旧记录
  2. ZCARD 统计窗口内剩余记录数
  3. 如果 < 阈值，ZADD 添加当前请求时间戳 → 放行
  4. 如果 ≥ 阈值 → 拒绝
```

### Go + Redis 完整实现

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/redis/go-redis/v9"
)

// ============ 滑动窗口限流器 ============

type SlidingWindowLimiter struct {
    client *redis.Client
}

func NewSlidingWindowLimiter(client *redis.Client) *SlidingWindowLimiter {
    return &SlidingWindowLimiter{client: client}
}

// Allow 检查是否允许请求
// userID: 用户标识
// limit:  窗口内最大请求数
// window: 窗口大小（如 60 秒）
func (l *SlidingWindowLimiter) Allow(
    ctx context.Context,
    userID string,
    limit int64,
    window time.Duration,
) (bool, error) {

    now := time.Now().UnixMilli()            // 当前时间（毫秒）
    windowStart := now - window.Milliseconds() // 窗口左边界

    key := fmt.Sprintf("rate_limit:%s", userID)

    // Lua 脚本：清理 + 计数 + 判断 + 添加，四步原子执行
    script := `
        local key = KEYS[1]
        local now = tonumber(ARGV[1])
        local window_start = tonumber(ARGV[2])
        local limit = tonumber(ARGV[3])
        local request_id = ARGV[4]

        -- 1. 清理窗口外的旧记录
        redis.call("ZREMRANGEBYSCORE", key, 0, window_start)

        -- 2. 统计窗口内的请求数
        local count = redis.call("ZCARD", key)

        -- 3. 判断是否超限
        if count >= limit then
            return 0  -- 拒绝
        end

        -- 4. 添加当前请求（score=时间戳, member=唯一ID 防重复）
        redis.call("ZADD", key, now, request_id)

        -- 5. 设过期时间防止 key 永久占用内存
        redis.call("EXPIRE", key, math.ceil(window / 1000) + 1)

        return 1  -- 放行
    `

    // 用纳秒时间戳+随机数作为 member，防止同一毫秒的请求被覆盖
    requestID := fmt.Sprintf("%d", now)

    result, err := l.client.Eval(ctx, script,
        []string{key},
        now,        // ARGV[1]
        windowStart, // ARGV[2]
        limit,      // ARGV[3]
        requestID,  // ARGV[4]
    ).Result()

    if err != nil {
        return false, fmt.Errorf("限流检查失败: %w", err)
    }

    return result.(int64) == 1, nil
}

// ============ HTTP 中间件 ============

func (l *SlidingWindowLimiter) Middleware(
    limit int64,
    window time.Duration,
) func(next http.Handler) http.Handler {

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            userID := r.Header.Get("X-User-ID")
            if userID == "" {
                userID = r.RemoteAddr
            }

            allowed, err := l.Allow(r.Context(), userID, limit, window)
            if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                return
            }
            if !allowed {
                w.Header().Set("Retry-After", fmt.Sprintf("%.0f", window.Seconds()))
                w.WriteHeader(http.StatusTooManyRequests) // 429
                w.Write([]byte(`{"error": "请求太频繁"}`))
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}

// ============ 使用示例 ============

func main() {
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    limiter := NewSlidingWindowLimiter(client)

    mux := http.NewServeMux()

    // 普通查询：每分钟 100 次
    mux.HandleFunc("/api/balance", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte(`{"balance": "1.5 ETH"}`))
    })

    // 提现接口：每分钟 3 次
    mux.Handle("/api/withdraw",
        limiter.Middleware(3, 1*time.Minute)(
            http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Write([]byte(`{"status": "ok"}`))
            }),
        ),
    )

    // 行情查询：每 10 秒 20 次
    mux.Handle("/api/price",
        limiter.Middleware(20, 10*time.Second)(
            http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.Write([]byte(`{"BTC": 68000}`))
            }),
        ),
    )

    fmt.Println("服务启动在 :8080")
    http.ListenAndServe(":8080", mux)
}
```

### 为什么用 ZSET？

```
ZSET（有序集合）的特点：
  - 每个元素有 score（分数），按 score 排序
  - 我们用时间戳当 score
  → ZREMRANGEBYSCORE 可以高效删除"时间戳小于某值"的所有记录
  → 天生适合"过去 N 秒的窗口"

如果用 List 或 Set：
  - List：只能从头/尾操作，删中间很麻烦
  - Set：没有排序，清理旧记录只能遍历
```

---

## 三、两套方案对比

| | 令牌桶（x/time/rate） | 滑动窗口（Redis ZSET） |
|---|---|---|
| 性能 | 极快（内存） | 快（网络调 Redis） |
| 精确度 | 中等（允许突发） | 高（任意时间窗口精确统计） |
| 多实例 | ❌ 各自计数 | ✅ Redis 共享状态 |
| 重启 | ❌ 数据丢失 | ✅ Redis 持久化 |
| 实现复杂度 | 低 | 中（需 Lua 脚本） |
| 适合场景 | 普通 API、查行情 | 提现、转账、发空投 |

---

## 四、Web3 进阶：分级限流策略

实际项目不是"一刀切"，而是**按用户等级 + 接口敏感度**组合限流：

```go
// 限流配置表
var rateLimitConfig = map[string]struct {
    Limit  int64
    Window time.Duration
}{
    // 未认证用户
    "/api/price":     {20, 10 * time.Second}, // 每10秒20次
    "/api/balance":   {10, 60 * time.Second}, // 每分钟10次

    // 已认证用户（查自己的数据）
    "/api/orders":    {30, 60 * time.Second}, // 每分钟30次

    // 敏感操作——全部走滑动窗口
    "/api/withdraw":  {3, 60 * time.Second},  // 每分钟3次
    "/api/transfer":  {5, 60 * time.Second},  // 每分钟5次
}

// 用户等级影响限流阈值
func getLimitMultiplier(userTier string) int64 {
    switch userTier {
    case "vip":
        return 5  // VIP 5倍限额
    case "pro":
        return 2  // Pro 2倍限额
    default:
        return 1
    }
}
```

---

## 五、面试口述话术（30 秒版）

> "两种方案。轻量级用令牌桶——Go 的 `x/time/rate` 包，给每个用户维护一个限流器，每秒放 10 个令牌，桶容量 20 允许少量突发，超过就返回 429。对精确度要求高的比如提现接口，用 Redis 滑动窗口——ZSET 存请求时间戳，每次先清理窗口外的旧记录，统计窗口内请求数，超过阈值就拒绝，Lua 脚本保证原子性。"

---

## 六、要点速记

- 令牌桶（轻量级）——Go `x/time/rate` 包，每用户一个限流器，每秒放 N 个令牌，桶容量允许少量突发，超限返回 429
- Redis 滑动窗口（精确型）——ZSET 存请求时间戳，每次清理窗口外旧记录再计数，适合提现等敏感接口
- Lua 脚本保原子——滑动窗口的清理+计数+判断打包成一个 Lua 脚本，避免并发问题
- 多实例部署必须用 Redis 方案——令牌桶纯内存，多实例各自计数等于没限流
