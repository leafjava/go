# 第19课：Redis 缓存 + 消息队列

> 学习时间：3-4小时 | 难度：⭐⭐⭐

## 📋 本课目标

- 掌握 go-redis 库的使用
- 实现常见的缓存策略
- 理解 Redis Pub/Sub 和 Stream 消息队列
- 实现异步任务处理
- 集成 Redis 到 Web3 服务

## 1. 安装 Redis 和 go-redis

### 安装 Redis

```bash
# Windows（使用 Docker）
docker run -d --name redis -p 6379:6379 redis:7-alpine

# 或使用 Windows 版 Redis
# 下载地址: https://github.com/tporadowski/redis/releases

# macOS
brew install redis && brew services start redis

# Linux
sudo apt install redis-server && sudo systemctl start redis
```

### 安装 Go 库

```bash
go get github.com/redis/go-redis/v9
```

## 2. Redis 基本操作

### 连接和基础命令

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/redis/go-redis/v9"
)

func main() {
    // 创建 Redis 客户端
    rdb := redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",  // Redis 地址
        Password: "",                // 密码（无密码留空）
        DB:       0,                 // 默认数据库
    })

    // 测试连接
    ctx := context.Background()
    pong, err := rdb.Ping(ctx).Result()
    if err != nil {
        log.Fatal("Redis 连接失败:", err)
    }
    fmt.Println("Redis 连接成功:", pong)

    // === 字符串操作 ===
    // SET：设置键值
    rdb.Set(ctx, "key", "value", 10*time.Minute)
    
    // GET：获取键值
    val, _ := rdb.Get(ctx, "key").Result()
    fmt.Println("key:", val)

    // SETNX：仅当键不存在时设置（可用于分布式锁）
    ok, _ := rdb.SetNX(ctx, "lock:task1", "locked", 30*time.Second).Result()
    fmt.Println("获取锁:", ok)

    // INCR：自增（原子操作，适合计数器）
    rdb.Incr(ctx, "counter")
    count, _ := rdb.Get(ctx, "counter").Int64()
    fmt.Println("计数器:", count)

    // === Hash 操作 ===
    // HSET：设置哈希字段
    rdb.HSet(ctx, "user:1", map[string]interface{}{
        "name":  "张三",
        "email": "zhangsan@example.com",
        "age":   25,
    })

    // HGET：获取单个字段
    name, _ := rdb.HGet(ctx, "user:1", "name").Result()
    fmt.Println("用户名:", name)

    // HGETALL：获取所有字段
    user, _ := rdb.HGetAll(ctx, "user:1").Result()
    fmt.Println("用户信息:", user)

    // === List 操作 ===
    // LPUSH / RPUSH：推入列表
    rdb.LPush(ctx, "queue:tasks", "task1", "task2", "task3")
    
    // BRPOP：阻塞弹出（消费者模式）
    task, _ := rdb.BRPop(ctx, 2*time.Second, "queue:tasks").Result()
    fmt.Println("获取任务:", task)

    // === Set 操作 ===
    // SADD：添加集合成员
    rdb.SAdd(ctx, "whitelist:addresses", "0x123...", "0x456...", "0x789...")
    
    // SISMEMBER：判断是否成员
    isMember, _ := rdb.SIsMember(ctx, "whitelist:addresses", "0x123...").Result()
    fmt.Println("在白名单:", isMember)

    // === Sorted Set 操作 ===
    // ZADD：添加有序集合（常用于排行榜）
    rdb.ZAdd(ctx, "leaderboard:gas", redis.Z{Score: 0.05, Member: "0x123..."})
    rdb.ZAdd(ctx, "leaderboard:gas", redis.Z{Score: 0.12, Member: "0x456..."})
    
    // ZRANGE：按排名查询
    top, _ := rdb.ZRevRangeWithScores(ctx, "leaderboard:gas", 0, 9).Result()
    fmt.Println("排行榜:", top)

    // === 过期时间 ===
    // EXPIRE：设置过期时间
    rdb.Expire(ctx, "session:abc123", 30*time.Minute)
    
    // TTL：查看剩余时间
    ttl, _ := rdb.TTL(ctx, "session:abc123").Result()
    fmt.Println("剩余时间:", ttl)
}
```

## 3. Redis 缓存策略

### Cache-Aside 模式（最常用）

```go
package services

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/redis/go-redis/v9"
    "gorm.io/gorm"
)

// CacheService 缓存服务
type CacheService struct {
    rdb *redis.Client
    db  *gorm.DB
}

// NewCacheService 创建缓存服务
func NewCacheService(rdb *redis.Client, db *gorm.DB) *CacheService {
    return &CacheService{rdb: rdb, db: db}
}

// GetUserWithCache Cache-Aside 模式获取用户
func (s *CacheService) GetUserWithCache(ctx context.Context, userID uint) (*User, error) {
    cacheKey := fmt.Sprintf("user:%d", userID)

    // 1. 先从缓存读取
    cached, err := s.rdb.Get(ctx, cacheKey).Result()
    if err == nil {
        var user User
        if err := json.Unmarshal([]byte(cached), &user); err == nil {
            return &user, nil
        }
    }

    // 2. 缓存未命中，从数据库读取
    var user User
    if err := s.db.First(&user, userID).Error; err != nil {
        return nil, err
    }

    // 3. 写入缓存
    data, _ := json.Marshal(user)
    s.rdb.Set(ctx, cacheKey, data, 10*time.Minute)

    return &user, nil
}

// InvalidateUserCache 淘汰用户缓存（更新/删除用户时调用）
func (s *CacheService) InvalidateUserCache(ctx context.Context, userID uint) error {
    cacheKey := fmt.Sprintf("user:%d", userID)
    return s.rdb.Del(ctx, cacheKey).Err()
}

// CacheAside 缓存策略说明
//
// 读流程:
//   1. 查缓存 → 命中：直接返回
//   2. 未命中 → 查数据库 → 写入缓存 → 返回
//
// 写流程:
//   1. 更新数据库
//   2. 删除缓存（不是更新缓存）
//   3. 下次读取时自动重建缓存
```

### 缓存穿透防护（布隆过滤器模拟）

```go
// BloomFilter 简单布隆过滤器（使用 Redis Bitmap）
type BloomFilter struct {
    rdb      *redis.Client
    key      string
    hashFunc int // 哈希函数数量
}

func NewBloomFilter(rdb *redis.Client, key string) *BloomFilter {
    return &BloomFilter{
        rdb:      rdb,
        key:      key,
        hashFunc: 3,
    }
}

// 缓存穿透防护：空值缓存
func (s *CacheService) GetWalletBalanceWithProtection(ctx context.Context, address string) (string, error) {
    cacheKey := fmt.Sprintf("wallet:balance:%s", address)
    nullMarker := "__NULL__"

    // 1. 查缓存
    cached, err := s.rdb.Get(ctx, cacheKey).Result()
    if err == nil {
        if cached == nullMarker {
            return "", fmt.Errorf("无效地址（已缓存）")
        }
        return cached, nil
    }

    // 2. 调用区块链 RPC（模拟）
    balance, err := s.fetchBalanceFromChain(ctx, address)
    if err != nil {
        // 3. 缓存空值，防止缓存穿透（短过期时间）
        s.rdb.Set(ctx, cacheKey, nullMarker, 1*time.Minute)
        return "", err
    }

    // 4. 缓存实际值
    s.rdb.Set(ctx, cacheKey, balance, 5*time.Minute)
    return balance, nil
}

func (s *CacheService) fetchBalanceFromChain(ctx context.Context, address string) (string, error) {
    // 模拟 RPC 调用
    time.Sleep(500 * time.Millisecond)
    return "1.5", nil
}
```

### 缓存热 Key 防护

```go
// GetBlockNumberWithLocalCache 带本地缓存的区块号查询
type LocalCache struct {
    data      sync.Map
    ttl       time.Duration
}

func NewLocalCache(ttl time.Duration) *LocalCache {
    return &LocalCache{ttl: ttl}
}

type cacheItem struct {
    value      interface{}
    expireTime time.Time
}

func (lc *LocalCache) Get(key string) (interface{}, bool) {
    if val, ok := lc.data.Load(key); ok {
        item := val.(cacheItem)
        if time.Now().Before(item.expireTime) {
            return item.value, true
        }
        lc.data.Delete(key)
    }
    return nil, false
}

func (lc *LocalCache) Set(key string, value interface{}) {
    lc.data.Store(key, cacheItem{
        value:      value,
        expireTime: time.Now().Add(lc.ttl),
    })
}

// GetGasPriceWithMultiLevelCache 多级缓存：本地 → Redis → API
func (s *CacheService) GetGasPriceWithMultiLevelCache(ctx context.Context) (string, error) {
    // 第1级：本地内存缓存（极快，进程内）
    if val, ok := s.localCache.Get("gas_price"); ok {
        return val.(string), nil
    }

    // 第2级：Redis 缓存（较快，网络 IO）
    cached, err := s.rdb.Get(ctx, "gas:price:current").Result()
    if err == nil {
        s.localCache.Set("gas_price", cached)
        return cached, nil
    }

    // 第3级：RPC API（慢）
    gasPrice, err := s.fetchGasPriceFromRPC(ctx)
    if err != nil {
        return "", err
    }

    // 回写缓存
    s.rdb.Set(ctx, "gas:price:current", gasPrice, 12*time.Second) // 12秒过期（新区块约12秒）
    s.localCache.Set("gas_price", gasPrice)

    return gasPrice, nil
}
```

## 4. Redis 消息队列

### Pub/Sub 发布订阅

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/redis/go-redis/v9"
)

func main() {
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    ctx := context.Background()

    // === 发布者 ===
    go func() {
        ticker := time.NewTicker(3 * time.Second)
        for range ticker.C {
            msg := fmt.Sprintf("新区块: %d", time.Now().Unix())
            err := rdb.Publish(ctx, "channel:blocks", msg).Err()
            if err != nil {
                log.Printf("发布失败: %v", err)
            }
            fmt.Printf("📤 发布消息: %s\n", msg)
        }
    }()

    // === 订阅者1 ===
    go func() {
        sub := rdb.Subscribe(ctx, "channel:blocks")
        defer sub.Close()

        ch := sub.Channel()
        for msg := range ch {
            fmt.Printf("📥 [订阅者1] 收到: %s\n", msg.Payload)
        }
    }()

    // === 订阅者2（使用模式订阅） ===
    go func() {
        sub := rdb.PSubscribe(ctx, "channel:*") // 匹配所有 channel: 开头的频道
        defer sub.Close()

        ch := sub.Channel()
        for msg := range ch {
            fmt.Printf("📥 [订阅者2（模式）] 频道=%s, 消息=%s\n", msg.Channel, msg.Payload)
        }
    }()

    // 等待
    time.Sleep(15 * time.Second)
}

// Pub/Sub 适用场景：
// ✅ 实时通知（区块链事件推送、用户消息）
// ✅ 配置热更新广播
// ✅ 缓存失效通知（一个实例更新数据后通知其他实例删除缓存）
//
// ❌ 不适用场景：
// ❌ 需要消息持久化（消费者离线时消息丢失）
// ❌ 需要消息确认和重试
```

### Redis Stream 消息队列（可靠消息）

```go
// TaskQueue 基于 Redis Stream 的任务队列
type TaskQueue struct {
    rdb       *redis.Client
    streamKey string
    groupName string
}

// NewTaskQueue 创建任务队列
func NewTaskQueue(rdb *redis.Client, streamKey, groupName string) (*TaskQueue, error) {
    ctx := context.Background()

    // 创建消费者组（如果不存在）
    err := rdb.XGroupCreateMkStream(ctx, streamKey, groupName, "0").Err()
    if err != nil && !strings.Contains(err.Error(), "BUSYGROUP") {
        return nil, err
    }

    return &TaskQueue{
        rdb:       rdb,
        streamKey: streamKey,
        groupName: groupName,
    }, nil
}

// PublishTask 发布任务
func (q *TaskQueue) PublishTask(ctx context.Context, taskType string, data map[string]interface{}) error {
    return q.rdb.XAdd(ctx, &redis.XAddArgs{
        Stream: q.streamKey,
        Values: map[string]interface{}{
            "type":    taskType,
            "data":    data,
            "created": time.Now().Unix(),
        },
    }).Err()
}

// ConsumeTasks 消费任务
func (q *TaskQueue) ConsumeTasks(ctx context.Context, consumerName string, handler func(taskType string, data map[string]interface{}) error) {
    for {
        // 阻塞读取新消息
        result, err := q.rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
            Group:    q.groupName,
            Consumer: consumerName,
            Streams:  []string{q.streamKey, ">"}, // ">" 表示读取未投递的新消息
            Count:    5,                           // 每次最多读取5条
            Block:    2 * time.Second,             // 阻塞等待时间
        }).Result()

        if err != nil {
            if err == redis.Nil {
                continue // 超时，继续等待
            }
            log.Printf("读取消息出错: %v", err)
            time.Sleep(time.Second)
            continue
        }

        for _, stream := range result {
            for _, msg := range stream.Messages {
                taskType := msg.Values["type"].(string)
                
                // 处理任务
                err := handler(taskType, msg.Values)
                if err != nil {
                    log.Printf("处理任务失败: %v", err)
                    continue
                }

                // 确认消息
                q.rdb.XAck(ctx, q.streamKey, q.groupName, msg.ID)
            }
        }
    }
}

// PendingTaskHandler 处理挂起的未确认消息
func (q *TaskQueue) HandlePendingTasks(ctx context.Context, handler func(taskType string, data map[string]interface{}) error) {
    // 读取挂起超过 30 秒的消息
    pending, err := q.rdb.XPendingExt(ctx, &redis.XPendingExtArgs{
        Stream: q.streamKey,
        Group:  q.groupName,
        Start:  "-",
        End:    "+",
        Count:  100,
        Idle:   30 * time.Second, // 超过30秒未被确认
    }).Result()

    if err != nil {
        log.Printf("查询挂起消息出错: %v", err)
        return
    }

    for _, p := range pending {
        // 认领并重新处理
        claimed, _ := q.rdb.XClaim(ctx, &redis.XClaimArgs{
            Stream:   q.streamKey,
            Group:    q.groupName,
            Consumer: "recovery",
            MinIdle:  30 * time.Second,
            Messages: []string{p.ID},
        }).Result()

        for _, msg := range claimed {
            taskType := msg.Values["type"].(string)
            if err := handler(taskType, msg.Values); err == nil {
                q.rdb.XAck(ctx, q.streamKey, q.groupName, msg.ID)
            }
        }
    }
}
```

## 5. Web3 实战：交易处理队列

### 完整示例：异步交易处理器

```go
package services

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "github.com/redis/go-redis/v9"
)

// TransactionTask 交易任务
type TransactionTask struct {
    ID          string `json:"id"`
    FromAddress string `json:"from_address"`
    ToAddress   string `json:"to_address"`
    Amount      string `json:"amount"`
    ChainID     int64  `json:"chain_id"`
    Status      string `json:"status"` // pending, processing, completed, failed
    TxHash      string `json:"tx_hash,omitempty"`
    Error       string `json:"error,omitempty"`
    CreatedAt   int64  `json:"created_at"`
}

// TransactionQueue 交易处理队列
type TransactionQueue struct {
    rdb         *redis.Client
    taskQueue   *TaskQueue
    resultQueue *TaskQueue
}

// NewTransactionQueue 创建交易队列
func NewTransactionQueue(rdb *redis.Client) (*TransactionQueue, error) {
    taskQueue, err := NewTaskQueue(rdb, "stream:transactions", "tx-processors")
    if err != nil {
        return nil, err
    }

    resultQueue, err := NewTaskQueue(rdb, "stream:tx-results", "tx-results-group")
    if err != nil {
        return nil, err
    }

    return &TransactionQueue{
        rdb:         rdb,
        taskQueue:   taskQueue,
        resultQueue: resultQueue,
    }, nil
}

// SubmitTransaction 提交交易任务
func (tq *TransactionQueue) SubmitTransaction(ctx context.Context, task TransactionTask) error {
    task.Status = "pending"
    task.CreatedAt = time.Now().Unix()
    task.ID = fmt.Sprintf("tx_%d_%s", task.CreatedAt, task.FromAddress[:10])

    // 保存任务状态
    data, _ := json.Marshal(task)
    tq.rdb.Set(ctx, "tx:task:"+task.ID, data, 24*time.Hour)

    // 发布到队列
    return tq.taskQueue.PublishTask(ctx, "send_transaction", map[string]interface{}{
        "task_id": task.ID,
        "task":    string(data),
    })
}

// ProcessTransactions 处理交易（消费者）
func (tq *TransactionQueue) ProcessTransactions(ctx context.Context) {
    tq.taskQueue.ConsumeTasks(ctx, "processor-1", func(taskType string, data map[string]interface{}) error {
        taskID := data["task_id"].(string)

        // 1. 更新状态为处理中
        tq.updateTaskStatus(ctx, taskID, "processing")

        // 2. 执行交易（模拟）
        log.Printf("处理交易: %s", taskID)
        time.Sleep(1 * time.Second) // 模拟区块链确认时间

        // 3. 模拟成功
        txHash := fmt.Sprintf("0x%x", time.Now().UnixNano())
        tq.completeTask(ctx, taskID, txHash, "")

        return nil
    })
}

// GetTaskStatus 查询任务状态
func (tq *TransactionQueue) GetTaskStatus(ctx context.Context, taskID string) (*TransactionTask, error) {
    data, err := tq.rdb.Get(ctx, "tx:task:"+taskID).Result()
    if err != nil {
        return nil, fmt.Errorf("任务不存在: %s", taskID)
    }

    var task TransactionTask
    if err := json.Unmarshal([]byte(data), &task); err != nil {
        return nil, err
    }

    return &task, nil
}

func (tq *TransactionQueue) updateTaskStatus(ctx context.Context, taskID, status string) {
    data, _ := tq.rdb.Get(ctx, "tx:task:"+taskID).Result()
    var task TransactionTask
    json.Unmarshal([]byte(data), &task)
    task.Status = status
    newData, _ := json.Marshal(task)
    tq.rdb.Set(ctx, "tx:task:"+taskID, newData, 24*time.Hour)
}

func (tq *TransactionQueue) completeTask(ctx context.Context, taskID, txHash, errMsg string) {
    data, _ := tq.rdb.Get(ctx, "tx:task:"+taskID).Result()
    var task TransactionTask
    json.Unmarshal([]byte(data), &task)
    
    if errMsg != "" {
        task.Status = "failed"
        task.Error = errMsg
    } else {
        task.Status = "completed"
        task.TxHash = txHash
    }

    newData, _ := json.Marshal(task)
    tq.rdb.Set(ctx, "tx:task:"+taskID, newData, 24*time.Hour)
}
```

## 6. Redis 与 Gin 集成

```go
package main

import (
    "context"
    "net/http"
    "time"

    "your-project/services"

    "github.com/gin-gonic/gin"
    "github.com/redis/go-redis/v9"
)

var (
    rdb     *redis.Client
    txQueue *services.TransactionQueue
)

func main() {
    // 初始化 Redis
    rdb = redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })

    // 初始化交易队列
    var err error
    txQueue, err = services.NewTransactionQueue(rdb)
    if err != nil {
        panic(err)
    }

    // 启动交易处理器
    go txQueue.ProcessTransactions(context.Background())

    r := gin.Default()

    // 交易 API
    tx := r.Group("/api/v1/transactions")
    {
        tx.POST("/submit", submitTransaction)
        tx.GET("/status/:id", getTransactionStatus)
    }

    // 缓存 API
    cache := r.Group("/api/v1/cache")
    {
        cache.GET("/block-number", getCachedBlockNumber)
        cache.DELETE("/block-number", invalidateBlockNumberCache)
    }

    r.Run(":8080")
}

// POST /api/v1/transactions/submit
type SubmitTxRequest struct {
    FromAddress string `json:"from_address" binding:"required"`
    ToAddress   string `json:"to_address" binding:"required"`
    Amount      string `json:"amount" binding:"required"`
}

func submitTransaction(c *gin.Context) {
    var req SubmitTxRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    task := services.TransactionTask{
        FromAddress: req.FromAddress,
        ToAddress:   req.ToAddress,
        Amount:      req.Amount,
        ChainID:     1, // Ethereum Mainnet
    }

    if err := txQueue.SubmitTransaction(context.Background(), task); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "提交任务失败"})
        return
    }

    c.JSON(http.StatusAccepted, gin.H{
        "task_id": task.ID,
        "status":  "pending",
        "message": "交易已提交，正在处理",
    })
}

// GET /api/v1/transactions/status/:id
func getTransactionStatus(c *gin.Context) {
    taskID := c.Param("id")

    task, err := txQueue.GetTaskStatus(context.Background(), taskID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, task)
}

// GET /api/v1/cache/block-number
func getCachedBlockNumber(c *gin.Context) {
    ctx := context.Background()
    cached, err := rdb.Get(ctx, "blockchain:block-number").Result()
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "缓存未命中"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "block_number": cached,
        "source":       "redis_cache",
    })
}

// DELETE /api/v1/cache/block-number
func invalidateBlockNumberCache(c *gin.Context) {
    rdb.Del(context.Background(), "blockchain:block-number")
    c.JSON(http.StatusOK, gin.H{"message": "缓存已清除"})
}
```

## 📝 作业

### 作业1：余额缓存服务

```go
// TODO: 实现区块链余额缓存服务
// 1. 查询余额时优先读 Redis 缓存
// 2. 缓存过期后自动从链上刷新
// 3. 实现缓存预热（服务启动时加载热数据）
// 4. 使用 Pipeline 批量查询以提升性能
```

### 作业2：交易重试队列

```go
// TODO: 实现带重试机制的交易队列
// 1. 交易失败自动重试（最多3次）
// 2. 指数退避延迟（1s → 2s → 4s）
// 3. 失败后写入死信队列
// 4. 提供管理 API 查看和重放失败交易
```

### 作业3：分布式限流器

```go
// TODO: 使用 Redis 实现分布式限流器
// 1. 实现令牌桶算法（Token Bucket）
// 2. 每用户每分钟最多 100 次 API 请求
// 3. 在 Gin 中间件中使用
// 4. 返回限流信息（剩余请求数、重置时间）
func RateLimitMiddleware(rdb *redis.Client, limit int, window time.Duration) gin.HandlerFunc {
    // TODO: 实现
    return func(c *gin.Context) {}
}
```

## 🎯 检查点

- ✅ 掌握 go-redis 基本操作
- ✅ 理解缓存策略（Cache-Aside、多级缓存、穿透防护）
- ✅ 能够使用 Pub/Sub 实现实时通知
- ✅ 掌握 Redis Stream 可靠消息队列
- ✅ 实现异步交易处理
- ✅ Redis 与 Gin 框架集成

## 💡 Redis 使用建议

| 场景 | 推荐数据结构 | 过期时间建议 |
|------|-------------|-------------|
| 用户信息缓存 | String (JSON) | 10-30 分钟 |
| Gas 价格缓存 | String | 12 秒（一个区块时间） |
| 会话存储 | String (JSON) | 30 分钟 - 24 小时 |
| 排行榜 | Sorted Set | 根据需要 |
| 限流计数器 | String (INCR) | 按限流窗口 |
| 任务队列 | Stream + Consumer Group | 处理完成后 ACK |
| 分布式锁 | String (SETNX) | 根据任务时长 |
| 白名单/黑名单 | Set | 长期/永久 |

## ⏭️ 下一课

[第20-21课：实战 - 区块链交互服务](./day20-21-blockchain-service.md)
