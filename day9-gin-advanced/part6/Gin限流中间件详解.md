# Gin 限流中间件（Rate Limiter）详解

## 1. 整体架构

```
请求 → RateLimitMiddleware → 检查次数 → 没超 → c.Next() → 处理函数 → 响应
                │                        │
                │                        └── 超了 → 返回 429 + c.Abort()
                │
                ▼
           RateLimiter
                │
    内存 map[IP → 请求时间列表]
```

## 2. 数据结构 `RateLimiter`

```go
type RateLimiter struct {
    requests map[string][]time.Time  // IP → 该 IP 的请求时间列表
    mu       sync.Mutex              // 互斥锁，防止并发冲突
    limit    int                     // 最大允许次数（示例：2 次）
    window   time.Duration           // 时间窗口（示例：1 分钟）
}
```

每个字段用 Vue 类比：

```js
const state = reactive({
    requests: {                         // map[string][]time.Time
        '127.0.0.1': [时间1, 时间2],     // 这个 IP 最近的请求时间
    },
    mu: null,                           // 锁（JS 单线程不需要）
    limit: 2,                           // 每分钟最多 2 次
    window: 60 * 1000,                  // 时间窗口 1 分钟
})
```

## 3. 构造函数 `NewRateLimiter`

```go
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
    return &RateLimiter{
        requests: make(map[string][]time.Time),  // 初始化空的 map
        limit:    limit,                          // 2 次
        window:   window,                         // 1 分钟
    }
}
```

`make(...)` 分配内存，不写就 nil，写入时会 panic。

调用：

```go
limiter := NewRateLimiter(2, time.Minute)
//                         ↑    ↑
//                       2次   1分钟 → 每个 IP 每分钟最多 2 次
```

## 4. 核心方法 `Allow(key string) bool`

分三段理解：

### 4.1 加锁

```go
func (rl *RateLimiter) Allow(key string) bool {
    rl.mu.Lock()           // ① 上锁：同一时间只能一个协程进入
    defer rl.mu.Unlock()   // ② 函数结束时自动解锁（不管怎么返回的）
    // ...
}
```

**为什么要锁？** Go 的 HTTP 天然支持并发，多个请求同时修改 `requests` map 会导致数据错乱或崩溃。锁保证串行执行：

```
请求A → Lock → 执行... → Unlock
请求B → 等待... → 拿到锁 → 执行... → Unlock
请求C → 等待... → 等待... → 拿到锁 → 执行...
```

### 4.2 清理过期记录（滑动窗口核心）

```go
now := time.Now()

// 这个 IP 之前来过
if requests, exists := rl.requests[key]; exists {
    var valid []time.Time
    for _, t := range requests {
        if now.Sub(t) < rl.window {   // 这个时间还在 1 分钟窗口内
            valid = append(valid, t)   // → 保留
        }
        // 超时的自动丢弃（不放进 valid）
    }
    rl.requests[key] = valid            // 只保留有效的
}
```

**图示**：

```
这个 IP 之前的请求时间：[10:00:05, 10:00:30, 10:00:50]
当前时间：10:01:10
窗口：1 分钟

逐个检查：
  10:00:05 → 距今 65 秒 > 1 分钟 → ❌ 扔掉
  10:00:30 → 距今 40 秒 < 1 分钟 → ✅ 保留
  10:00:50 → 距今 20 秒 < 1 分钟 → ✅ 保留

valid = [10:00:30, 10:00:50]  ← 只保留最近 1 分钟内的
```

### 4.3 判断是否超限 + 记录本次

```go
    // 超了？
    if len(rl.requests[key]) >= rl.limit {
        return false         // 拒绝
    }

    // 没超，记录本次请求时间
    rl.requests[key] = append(rl.requests[key], now)
    return true              // 放行
}
```

## 5. 中间件 `RateLimitMiddleware`

```go
func RateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        key := c.ClientIP()          // ① 取客户端 IP 作为限流标识

        if !limiter.Allow(key) {     // ② 检查是否超限
            c.JSON(429, gin.H{       // ③ 超了 → HTTP 429
                "error": "请求过于频繁，请稍后再试",
            })
            c.Abort()                //     → 拦截，不再执行后续
            return
        }

        c.Next()                     // ④ 没超 → 放行到处理函数
    }
}
```

> 复习：[Gin中间件详解.md](../Gin中间件详解.md) — `c.Abort()` 拦截 vs `c.Next()` 放行

## 6. 完整请求流程模拟

设置 `limit=2, window=1分钟`，同一 IP 连续 4 次请求：

```
第 1 次（10:00:00）
─────────────────
requests["127.0.0.1"] = []
  → 清理：空的，不用清
  → 检查：len=0 < 2 → 没超
  → 记录：[10:00:00]
  → 返回 true ✅  →  {"data": "success"}

第 2 次（10:00:30）
─────────────────
requests["127.0.0.1"] = [10:00:00]
  → 清理：10:00:00 距今 30 秒 < 1 分钟 → 保留
  → 检查：len=1 < 2 → 没超
  → 记录：[10:00:00, 10:00:30]
  → 返回 true ✅  →  {"data": "success"}

第 3 次（10:00:45）
─────────────────
requests["127.0.0.1"] = [10:00:00, 10:00:30]
  → 清理：两个都在 1 分钟窗口内 → 都保留
  → 检查：len=2 >= 2 → 超了！
  → 返回 false ❌  →  {"error": "请求过于频繁，请稍后再试"}（429）

第 4 次（10:01:10）— 等了一会儿，窗口滑动了
─────────────────
requests["127.0.0.1"] = [10:00:00, 10:00:30]
  → 清理：
    10:00:00 → 距今 70 秒 > 1 分钟 → 扔掉 ❌
    10:00:30 → 距今 40 秒 < 1 分钟 → 保留 ✅
    valid = [10:00:30]
  → 检查：len=1 < 2 → 没超！
  → 记录：[10:00:30, 10:01:10]
  → 返回 true ✅  →  {"data": "success"}
```

## 7. 滑动窗口 vs 固定窗口

| | 滑动窗口（本例） | 固定窗口 |
|------|------|------|
| 原理 | 动态看"最近 1 分钟" | 死看"每分钟整点重置" |
| 临界问题 | ✅ 没有 | ❌ 00:59 来 2 个，01:00 马上能来 2 个——实际 1 秒内来了 4 个 |
| 实现 | 略复杂 | 简单 |

```
固定窗口的漏洞：
  12:00:59 → 请求 x2 ✅（第 1 分钟的额度）
  12:01:00 → 请求 x2 ✅（第 2 分钟的额度）
  实际：1 秒内来了 4 个请求！

滑动窗口没有这个问题：
  12:00:59 → 请求 x2 ✅
  12:01:00 → 检查最近 1 分钟 → 还有 2 条记录 → 拒绝！
```

## 8. 一句话总结

> 每个请求过来 → 取 IP → 清理 1 分钟前的旧记录 → 数一下还剩几次 → 超过限制就返回 429 + `c.Abort()`，没超就记录时间并 `c.Next()` 放行。锁保证并发安全。这就是**滑动窗口限流算法**。
