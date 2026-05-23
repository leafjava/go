# Go 并发安全：sync.Mutex / sync.RWMutex / sync.Map 详解

---

## 前置知识：为什么需要锁？

Go 中多个 Goroutine 同时读写同一块内存时，会发生**数据竞争（data race）**，导致结果不可预期。

```go
// ❌ 有数据竞争的代码
var counter int

for i := 0; i < 1000; i++ {
    go func() {
        counter++ // 多个 Goroutine 同时写，结果不确定
    }()
}
fmt.Println(counter) // 每次运行结果都不一样
```

三种解决方案，对应本文三个主题：

| 方案 | 适用场景 |
|------|---------|
| `sync.Mutex` | 读写均衡，通用 |
| `sync.RWMutex` | 读多写少 |
| `sync.Map` | 特定场景（见下文） |

---

## 一、sync.Mutex（互斥锁）

### 是什么？

一把锁，同一时刻**只有一个** Goroutine 能持有。持有者可以读/写共享数据，其他人排队等。

**比喻**：厕所门锁，进去一个人锁门，外面的人排队等。

### 基本用法

```go
var mu sync.Mutex
var counter int

// 写操作：加锁
func increment() {
    mu.Lock()         // 拿锁，拿不到就阻塞等待
    counter++         // 安全地修改
    mu.Unlock()       // 释放锁
}

// 读操作：也要加锁
func getCounter() int {
    mu.Lock()
    defer mu.Unlock() // defer 保证即使 panic 也解锁
    return counter
}
```

### 关键细节：defer 解锁

```go
// ✅ 推荐写法：用 defer 保证解锁
func safeUpdate() {
    mu.Lock()
    defer mu.Unlock() // 函数返回前一定解锁，包括 panic 的情况

    // 复杂逻辑...
    if someCondition {
        return // 这里 return 前 defer 会自动执行 Unlock
    }
    // 更多操作...
}

// ❌ 危险写法：手动解锁，容易漏
func dangerousUpdate() {
    mu.Lock()
    if someCondition {
        return // 忘了解锁！死锁！
    }
    mu.Unlock()
}
```

### 完整例子：银行转账

```go
package main

import (
    "fmt"
    "sync"
)

type BankAccount struct {
    mu      sync.Mutex
    balance int
}

// 存钱
func (a *BankAccount) Deposit(amount int) {
    a.mu.Lock()
    defer a.mu.Unlock()
    a.balance += amount
}

// 取钱
func (a *BankAccount) Withdraw(amount int) bool {
    a.mu.Lock()
    defer a.mu.Unlock()
    if a.balance >= amount {
        a.balance -= amount
        return true
    }
    return false
}

// 查余额
func (a *BankAccount) Balance() int {
    a.mu.Lock()
    defer a.mu.Unlock()
    return a.balance
}

func main() {
    account := &BankAccount{balance: 1000}

    var wg sync.WaitGroup

    // 100 个 Goroutine 同时操作
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            account.Deposit(1) // 安全，Mutex 保护
        }()
    }

    wg.Wait()
    fmt.Println(account.Balance()) // 1100，正确！
}
```

---

## 二、sync.RWMutex（读写锁）

### 是什么？

Mutex 的升级版，区分"读"和"写"：

- **读锁（RLock）**：多个 Goroutine 可以**同时**持有读锁
- **写锁（Lock）**：同一时刻只有**一个** Goroutine 持有，且**不能和读锁共存**

**比喻**：图书馆的书。多个人可以同时看书（读锁共享），但有人要修改书时（写锁），其他人不能看也不能改。

### 与 Mutex 的核心区别

```
Mutex（普通锁）：
  读操作 ──┬── 读操作   ❌ 互斥，等
           ├── 写操作   ❌ 互斥，等

RWMutex（读写锁）：
  读操作 ──┬── 读操作   ✅ 可以同时进行！
           ├── 写操作   ❌ 互斥，等
```

### 基本用法

```go
var rwmu sync.RWMutex
var config map[string]string

// 读操作：用 RLock
func getConfig(key string) string {
    rwmu.RLock()              // 读锁，其他人也能同时读
    defer rwmu.RUnlock()
    return config[key]
}

// 写操作：用 Lock（和 Mutex 一样）
func setConfig(key, value string) {
    rwmu.Lock()               // 写锁，独占
    defer rwmu.Unlock()
    config[key] = value
}
```

### 完整例子：系统配置管理器

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

type ConfigManager struct {
    mu     sync.RWMutex
    config map[string]string
}

func NewConfigManager() *ConfigManager {
    return &ConfigManager{
        config: map[string]string{
            "host": "localhost",
            "port": "8080",
        },
    }
}

// 读取配置：使用读锁
func (c *ConfigManager) Get(key string) string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    return c.config[key]
}

// 更新配置：使用写锁
func (c *ConfigManager) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.config[key] = value
}

// 打印全部配置：使用读锁
func (c *ConfigManager) GetAll() map[string]string {
    c.mu.RLock()
    defer c.mu.RUnlock()
    // 注意：返回的是原始 map 的引用，调用方不应该修改它
    // 实际项目中最好返回一个拷贝
    result := make(map[string]string)
    for k, v := range c.config {
        result[k] = v
    }
    return result
}

func main() {
    cm := NewConfigManager()

    // 100 个 Goroutine 同时读，不会互相阻塞
    for i := 0; i < 100; i++ {
        go func() {
            host := cm.Get("host")
            _ = host
        }()
    }

    // 写操作会等所有读操作完成后才执行
    go func() {
        time.Sleep(10 * time.Millisecond)
        cm.Set("port", "9090")
        fmt.Println("配置已更新")
    }()

    time.Sleep(100 * time.Millisecond)
    fmt.Println(cm.GetAll())
}
```

### 何时用 RWMutex 而非 Mutex？

```
✅ 用 RWMutex：读多写少
   例：系统配置、白名单、缓存、路由表

✅ 用 Mutex：读写差不多
   例：计数器、队列、状态机
```

---

## 三、sync.Map（并发安全 Map）

### 是什么？

Go 官方提供的**开箱即用**的并发安全 Map，不需要自己加锁。

**关键理解**：它不是 "Mutex + map" 的简单包装，内部做了优化。大多数情况下，**普通 map + Mutex 更好**，sync.Map 只适用于两种特定场景。

### 基本用法

```go
var sm sync.Map

// 存
sm.Store("key1", "value1")
sm.Store("key2", 123) // 注意：key 和 value 都是 interface{}

// 取
v, ok := sm.Load("key1")
if ok {
    fmt.Println(v.(string)) // 需要类型断言
}

// 存或读：如果有就返回，没有就存
actual, loaded := sm.LoadOrStore("key1", "newValue")
fmt.Println(actual, loaded) // value1 true（已存在，返回原值）

// 删除
sm.Delete("key1")

// 遍历（用于场景2：多 Goroutine 操作不同 key）
sm.Range(func(key, value interface{}) bool {
    fmt.Println(key, value)
    return true // 返回 false 停止遍历
})
```

### 两个适用场景（官方文档说的）

**场景 1：key 只写一次，但读很多次**

```go
// ✅ sync.Map 擅长：缓存类的场景
var cache sync.Map

// 初始化阶段：写一次
func initCache() {
    cache.Store("usdt", "0xdAC17F958D2ee523a2206206994597C13D831ec7")
    cache.Store("wbtc", "0x2260FAC5E5542a773Aa44fBCfeDf7C193bc2C599")
    // ... 只写一次
}

// 运行阶段：大量读
func getTokenAddress(symbol string) string {
    v, ok := cache.Load(symbol)
    if !ok {
        return ""
    }
    return v.(string)
}
```

**场景 2：多个 Goroutine 操作各自不同的 key（无竞争）**

```go
// ✅ sync.Map 擅长：每个 Goroutine 操作自己的 key
var userSessions sync.Map

func handleSession(userID string) {
    // 每个 userID 对应不同 key，Goroutine 之间不冲突
    session := &Session{UserID: userID}
    userSessions.Store(userID, session)

    // 后续更新也是同一个 key
    // ...
}
```

### 什么时候别用 sync.Map？

```go
// ❌ sync.Map 不适合：频繁更新同一个 key
var sm sync.Map
sm.Store("counter", 0)
for i := 0; i < 1000; i++ {
    go func() {
        // 大家都在改同一个 key，sync.Map 效率不如 Mutex + map
        v, _ := sm.Load("counter")
        sm.Store("counter", v.(int)+1) // 还有数据竞争！
    }()
}

// ✅ 这种场景用 Mutex + map 更好
var mu sync.Mutex
var m = make(map[string]int)
m["counter"] = 0
for i := 0; i < 1000; i++ {
    go func() {
        mu.Lock()
        m["counter"]++
        mu.Unlock()
    }()
}
```

### sync.Map 的"坑"

```go
// 坑1：需要类型断言，失去类型安全
var sm sync.Map
sm.Store("count", 100)
// sm.Store("count", "一百") // 不会编译报错！运行时才出问题

// 坑2：Range 遍历期间可能看到脏数据
sm.Range(func(k, v interface{}) bool {
    // 遍历期间如果有其他 Goroutine 在写，结果是"尽力而为"的
    return true
})

// 坑3：没有 Len() 方法！想知道大小只能遍历数
count := 0
sm.Range(func(k, v interface{}) bool {
    count++
    return true
})
```

---

## 四、三者对比总结

| 特性 | Mutex | RWMutex | sync.Map |
|------|-------|---------|----------|
| 类型安全 | ✅ 和普通类型搭配 | ✅ 和普通类型搭配 | ❌ `interface{}`，需类型断言 |
| 读并发 | ❌ 读也互斥 | ✅ 多个读同时进行 | ✅ 无锁读（特定场景） |
| 使用复杂度 | 简单 | 简单（多两个方法） | 中等（API 不同） |
| 适用场景 | 通用 | 读多写少 | 写一次读多次 / 不同 key 不同 Goroutine |
| 内存开销 | 8 字节 | 24 字节 | 内部复杂，更大 |

---

## 五、面试口述话术（30 秒版）

> "三种锁各司其职。**Mutex** 最通用，读写都互斥，适合计数器、状态机这类读写均衡的场景。**RWMutex** 适合读多写少，比如系统配置、白名单，多个读可以同时进行不互斥。**sync.Map** 有两个特定场景——key 只写一次但读很多次，或者多个 Goroutine 各自操作不同的 key。大多数情况下 Mutex 加普通 map 就够用了，sync.Map 不是银弹，需要类型断言还丢掉了类型安全。"

---

## 六、快速决策流程图

```
你需要并发安全地访问共享数据吗？
  │
  ├── 否 → 不用锁，普通变量即可
  │
  └── 是 → 数据类型是什么？
            │
            ├── map，且 key 只写一次 / 各 Goroutine 操作不同 key
            │     → sync.Map
            │
            ├── map，频繁更新
            │     → Mutex + map 或 RWMutex + map
            │
            └── 普通变量
                  │
                  ├── 读多写少 → RWMutex
                  └── 读写均衡 → Mutex
```
