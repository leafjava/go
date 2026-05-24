# Redis 分布式锁详解（从零开始）

---

## 第一章：Redis 是什么？（给忘了的人）

### 一句话

Redis 是一个**内存数据库**——数据存在内存里而不是硬盘上，所以**快到离谱**（微秒级响应）。

### 它和 MySQL 的区别

```
MySQL（硬盘数据库）：
  存盘 → 慢（毫秒级） → 但数据不会丢
  适合：用户信息、订单、交易记录等需要持久化的数据

Redis（内存数据库）：
  存内存 → 极快（微秒级） → 但断电就丢
  适合：缓存、计数器、排行榜、分布式锁
```

### 基本操作（就像操作一个超大 JSON）

```bash
# 存字符串
SET name "zhangsan"

# 取字符串
GET name          # 返回 "zhangsan"

# 存 + 设过期时间（秒），时间到自动删除
SET token "abc123" EX 60

# 数字自增
SET count 0
INCR count        # 返回 1
INCR count        # 返回 2

# 删除
DEL name
```

### 典型的缓存场景

```go
// ❌ 不用 Redis：每次请求都查 MySQL
func GetUser(id string) *User {
    return db.Query("SELECT * FROM users WHERE id = ?", id)  // 慢
}

// ✅ 用 Redis 缓存：
func GetUser(id string) *User {
    // 先看 Redis 里有没有
    if cached := redis.Get("user:" + id); cached != nil {
        return cached  // 命中缓存，极快
    }
    // 没有再去 MySQL 查，查完存 Redis
    user := db.Query("SELECT * FROM users WHERE id = ?", id)
    redis.Set("user:"+id, user, 60)  // 缓存 60 秒
    return user
}
```

---

## 第二章：什么是分布式锁？

### 先理解什么是"锁"

锁就是**排队**——一个厕所坑位，进去一个人锁门，外面的人排队等。

```
单机程序：
  Goroutine A 拿到锁 → 执行 → 释放锁
  Goroutine B 排队等 → 锁释放了 → 拿到锁 → 执行
```

### 为什么需要"分布式"锁？

单机程序里 Goroutine 都在同一个进程里，用 `sync.Mutex` 就够了。但分布式系统里，**多个服务实例跑在不同的机器上**，`sync.Mutex` 管不到别人：

```
服务器 1                    服务器 2
  │                           │
  │ 定时任务："发奖励"         │ 定时任务："发奖励"
  │ sync.Mutex 锁住自己       │ sync.Mutex 锁住自己
  │                           │
  └────────┬──────────────────┘
           │
           ▼
    同一个数据库，同一个用户
    → 两个实例都以为"只有我在发"，用户收到两份奖励！

✅ 分布式锁：
  服务器 1 先到 Redis："key 是我的了！"
  服务器 2 再来 Redis："key 被人占了啊，我等..."
  → 全系统只有一个人在发奖励
```

### 分布式锁的核心要求

```
1. 互斥：同一时刻只有一个客户端能拿到锁
2. 不死锁：持有锁的实例崩溃了，锁能自动释放
3. 解铃还须系铃人：只有加锁的人才能解锁
```

---

## 第三章：Redis 实现分布式锁

### 3.1 加锁：SETNX

`SETNX` = **SET** if **N**ot e**X**ists（如果 key 不存在才设置，存在就失败）

```bash
# 加锁
SET lock:send_reward "my-random-value" NX EX 30

# 拆解这条命令：
#   lock:send_reward  → 锁的名字（key）
#   "my-random-value" → 锁的值（用来验证"是不是我加的锁"）
#   NX                → Not eXists，key 不存在才能设置成功
#   EX 30             → 30 秒后自动过期删除
#
# 返回 OK  → 加锁成功
# 返回 nil → 锁已被别人拿着
```

**为什么一条命令**：

```
❌ 分开两步（不安全）：
  SET lock:xxx "value" NX     # 第 1 步：试着加锁
  EXPIRE lock:xxx 30           # 第 2 步：设过期时间
  如果第 1 步成功后，程序崩溃了，第 2 步没执行 → 死锁！

✅ 一条命令（安全）：
  SET lock:xxx "value" NX EX 30
  加锁和设过期是原子操作，要么都成功要么都不成功
```

### 3.2 解锁：Lua 脚本

解锁不能简单 `DEL`：

```
危险场景：
  1. 客户端 A 加锁，设 TTL 30 秒
  2. 客户端 A 的业务逻辑跑了 35 秒（超过了 30 秒）
  3. Redis 自动释放了 A 的锁（30 秒到期）
  4. 客户端 B 加锁成功
  5. 客户端 A 业务结束，执行 DEL → 把 B 的锁删了！
  6. 客户端 C 加锁成功 → A 和 C 同时认为自己持有锁
```

**解决方案**：加锁时设一个随机 value，解锁时先验证 value 是不是自己的：

```lua
-- Lua 脚本：原子地"验证 + 删除"
-- Redis 执行 Lua 脚本时，整个脚本是原子的（中间不会被其他命令插入）

if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
else
    return 0
end

-- 逻辑：
--   KEYS[1] = 锁的 key，如 "lock:send_reward"
--   ARGV[1] = 我当初加锁时设置的随机 value
--   如果 value 对得上 → 是我的锁，删除
--   如果 value 对不上 → 不是我的锁（可能被 TTL 释放了被别人拿了），别动
```

### 3.3 Go 代码完整实现

```go
package main

import (
    "context"
    "crypto/rand"
    "encoding/hex"
    "errors"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
)

// ============ 分布式锁 ============

type RedisLock struct {
    client *redis.Client
    key    string   // 锁的 key
    value  string   // 随机 value，用来验证"是不是我的锁"
    ttl    time.Duration
}

// NewLock 创建一个锁实例（还没加锁）
func NewLock(client *redis.Client, key string, ttl time.Duration) *RedisLock {
    return &RedisLock{
        client: client,
        key:    key,
        value:  randomValue(), // 每次创建都生成随机 value
        ttl:    ttl,
    }
}

// Lock 加锁，重试直到成功或超时
func (l *RedisLock) Lock(ctx context.Context, retryTimeout time.Duration) error {
    deadline := time.Now().Add(retryTimeout)

    for {
        // SET key value NX EX ttl
        ok, err := l.client.SetNX(ctx, l.key, l.value, l.ttl).Result()
        if err != nil {
            return fmt.Errorf("加锁失败: %w", err)
        }
        if ok {
            return nil // 加锁成功！
        }

        // 超时检查
        if time.Now().After(deadline) {
            return errors.New("加锁超时：锁被其他人持有")
        }

        // 等 100 毫秒再重试
        time.Sleep(100 * time.Millisecond)
    }
}

// Unlock 解锁（Lua 脚本保证安全）
func (l *RedisLock) Unlock(ctx context.Context) error {
    // Lua 脚本：验证 value 是自己的才删
    script := `
        if redis.call("GET", KEYS[1]) == ARGV[1] then
            return redis.call("DEL", KEYS[1])
        else
            return 0
        end
    `
    result, err := l.client.Eval(ctx, script, []string{l.key}, l.value).Result()
    if err != nil {
        return fmt.Errorf("解锁失败: %w", err)
    }
    if result.(int64) == 0 {
        return errors.New("解锁失败：锁不属于自己（可能已过期）")
    }
    return nil
}

// randomValue 生成随机字符串，保证每个客户端加锁的 value 不同
func randomValue() string {
    b := make([]byte, 16)
    rand.Read(b)
    return hex.EncodeToString(b)
}

// ============ 使用示例 ============

func main() {
    client := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    ctx := context.Background()

    // 创建锁
    lock := NewLock(client, "lock:send_reward", 30*time.Second)

    // 加锁（最多等 10 秒）
    if err := lock.Lock(ctx, 10*time.Second); err != nil {
        fmt.Println("加锁失败:", err)
        return
    }
    defer lock.Unlock(ctx) // 函数退出时解锁

    // ========== 受保护的业务逻辑 ==========
    fmt.Println("拿到锁了，执行定时任务...")
    sendRewards() // 发奖励
    fmt.Println("定时任务完成")
    // =====================================
}

func sendRewards() {
    time.Sleep(2 * time.Second) // 模拟业务逻辑
    fmt.Println("奖励已发放！")
}
```

---

## 第四章：三个经典的坑

### 坑 1：不设 TTL → 死锁

```
场景：
  客户端 A 加锁成功
  客户端 A 的进程突然崩溃（OOM / 断电 / kill -9）
  → 锁永远不会释放
  → 其他所有客户端永远等

解决：SET key value NX EX 30  ← EX 就是 TTL
     即使进程崩溃，30 秒后 Redis 自动删除 key
```

**TTL 设多长**：比你预估的业务最长执行时间多 3~5 秒。比如任务通常 3 秒跑完，TTL 设 10 秒。

### 坑 2：解锁不验证 → 删了别人的锁

```
时间线：
  0s   客户端 A 加锁，TTL=10s
  12s  A 的业务还在跑（超了 TTL）
       但 Redis 已经在 10s 自动释放了 A 的锁
  13s  客户端 B 加锁成功
  15s  A 业务结束，执行 DEL → 把 B 的锁删了！
  16s  客户端 C 加锁成功
  → B 和 C 同时认为自己持有锁 ⚠️

解决：解锁前用 Lua 脚本验证 value
     只有 value 是自己的才能 DEL
     A 的 value 和 B 的 value 不同，所以 A 删不掉 B 的锁
```

### 坑 3：主从架构 → 脑裂（Split Brain）

Redis 主从架构：一个主节点（可读写）+ 多个从节点（只读）。

```
时间线：
  0s   客户端 A 向主节点加锁 → 成功
  0.1s 主节点还没来得及同步给从节点...
  0.2s 主节点宕机！
  0.3s 哨兵选举从节点 1 为新的主节点
  0.4s 客户端 B 向新的主节点加锁 → 成功！
       （因为 A 的锁还没同步过来，新主节点上没有这个 key）
  → A 和 B 同时拿到锁 ⚠️

解决：Redlock 算法
  不是向一个 Redis 加锁，而是向 5 个独立的 Redis 节点分别加锁
  拿到超过半数（≥3 个）才算加锁成功
  任何单节点宕机都不会导致脑裂
```

Redlock 的代价：需要维护多个 Redis 实例，性能更低。如果业务允许短暂的不一致，单实例 + 合理 TTL 通常够了。

---

## 第五章：总结与面试话术

### 核心要素速记

| 要素 | 怎么做 | 为什么 |
|------|--------|--------|
| 加锁 | `SET key value NX EX ttl` | 原子操作，一条命令完成加锁+过期 |
| 解锁 | Lua 脚本：先 GET 比对再 DEL | 保证只删自己的锁 |
| 防死锁 | 必须设 TTL | 持有者崩溃，超时自动释放 |
| 防误删 | 随机 value 标识身份 | 区分"我的锁"和"别人的锁" |
| 防脑裂 | Redlock（多独立节点） | 主从同步延迟导致锁丢失 |

### 30 秒口述话术

> "用 Redis 的 SETNX 做加锁——`SET key value NX EX ttl`，原子操作不分裂。解锁必须用 Lua 脚本先 GET 比较 value 再 DEL，保证只有加锁的人才能解锁。三个常见的坑：一是必须设 TTL，否则持有者崩溃就死锁了；二是解锁前必须验证 value 是不是自己设的，不然可能把别人的锁删了，这就是为什么用 Lua 脚本；三是主从架构下主节点宕机，锁可能还没同步到从节点，导致两个客户端同时拿到锁，解决方案是 Redlock 算法向多个独立节点加锁。"

### Redis 分布式锁清单

- `SET key value NX EX ttl`——一条命令完成加锁+设过期，原子不分步
- 必须设 TTL——持有者崩溃也不死锁，超时自动释放
- Lua 脚本解锁——先 `GET` 比对 value 是不是自己的，再 `DEL`，保证只删自己的锁
- 主从防脑裂——主节点宕机锁没同步到从节点 → 两个客户端同时拿锁，用 Redlock（多独立节点）解决
