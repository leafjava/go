# Go 无锁并发安全数据结构详解

---

## 前置知识：Go 的并发哲学

Go 的官方名言：

> **"Don't communicate by sharing memory; share memory by communicating."**
> 不要通过共享内存来通信，而要通过通信来共享内存。

传统多线程编程（Java/C++）的思路：多个线程访问同一块内存，用锁来保护。
Go 的思路反过来：让**一个** Goroutine 独享数据，别人通过 Channel 发消息"请它"操作数据。

---

## 一、核心思路：Channel 通信模型

### 传统方式（共享内存 + 锁）

```
多个 Goroutine 直接读写同一块内存
        │
   ┌────┴────┐
   ▼         ▼
 锁 保护   锁 保护
   ▼         ▼
 共享 map / slice / struct
```

### Go 方式（Channel 通信）

```
一个后台 Goroutine 独享数据
         │
    ┌────┴────┐
    ▼         ▼
  请求       请求
(Channel)   (Channel)

其他 Goroutine 通过 Channel 发"请求"
后台 Goroutine 收到后操作数据，结果通过 Channel 返回
```

**关键区别**：不管多少个 Goroutine 发请求，**始终只有一个人碰数据** → 天然不需要锁。

---

## 二、招式一：后台 Goroutine 独享 + Channel 通信

### 完整例子：并发安全的缓存

```go
package main

import (
    "fmt"
    "sync"
)

// ============ 消息类型定义 ============

// 请求结构：想做什么 + 结果返回到哪里
type getRequest struct {
    key    string
    result chan string // 通过这个 Channel 把结果返回给调用方
}

type setRequest struct {
    key   string
    value string
}

// ============ 缓存 ============

type Cache struct {
    data  map[string]string // 数据本体，只有后台 Goroutine 碰它
    getCh chan getRequest   // 接收"读"请求
    setCh chan setRequest   // 接收"写"请求
    stop  chan struct{}     // 停止信号
}

func NewCache() *Cache {
    c := &Cache{
        data:  make(map[string]string),
        getCh: make(chan getRequest),
        setCh: make(chan setRequest),
        stop:  make(chan struct{}),
    }
    go c.loop() // 启动后台 Goroutine
    return c
}

// ============ 后台 Goroutine：唯一碰数据的人 ============

func (c *Cache) loop() {
    for {
        select {
        case req := <-c.getCh:
            // 查 map，塞回 result channel
            req.result <- c.data[req.key]

        case req := <-c.setCh:
            // 写 map
            c.data[req.key] = req.value

        case <-c.stop:
            return // 优雅退出
        }
    }
}

// ============ 公开方法：通过 Channel 发请求 ============

func (c *Cache) Get(key string) string {
    req := getRequest{
        key:    key,
        result: make(chan string), // 创建一个"回信"的 Channel
    }
    c.getCh <- req      // 发请求
    return <-req.result // 阻塞等待结果
}

func (c *Cache) Set(key, value string) {
    c.setCh <- setRequest{key: key, value: value}
}

func (c *Cache) Close() {
    close(c.stop)
}

// ============ 使用 ============

func main() {
    cache := NewCache()
    defer cache.Close()

    var wg sync.WaitGroup

    // 100 个 Goroutine 同时读写
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            key := fmt.Sprintf("user-%d", n%10)
            cache.Set(key, fmt.Sprintf("value-%d", n))
            _ = cache.Get(key)
        }(i)
    }

    wg.Wait()
    fmt.Println(cache.Get("user-1")) // 正常输出，无数据竞争
}
```

### 图解数据流

```
调用方 Goroutine A          调用方 Goroutine B
      │                           │
      │ cache.Get("foo")          │ cache.Set("bar", "v")
      ▼                           ▼
   getCh ──────────────────→  setCh
      │                           │
      ▼                           ▼
┌─────────────────────────────────────────┐
│          后台 loop() Goroutine          │
│                                          │
│   select {                              │
│   case req := <-getCh:                  │
│       req.result <- data[req.key]  ◀── 只有它碰 data！
│   case req := <-setCh:                  │
│       data[req.key] = req.value          │
│   }                                     │
└─────────────────────────────────────────┘
```

---

## 三、招式二：请求带 result channel（阻塞获取结果）

上一个例子已经包含了这个模式，这里单独拆解：

```go
// 请求结构里嵌入一个"回信地址"
type getRequest struct {
    key    string
    result chan string // ← 这就是 result channel，相当于"请把结果放到这里"
}

// 调用方：创建 result channel，等待回信
func (c *Cache) Get(key string) string {
    req := getRequest{
        key:    key,
        result: make(chan string), // 创建一个专属的"回信通道"
    }
    c.getCh <- req      // 发送请求（请求里带有回信地址）
    return <-req.result // 阻塞，等后台 Goroutine 把结果塞回来
}

// 后台：收到请求，处理完再写回 result channel
case req := <-c.getCh:
    req.result <- c.data[req.key] // 把结果返回给调用方
```

**本质**：用无缓冲 Channel 实现"同步等待"——调用方发出请求后阻塞，直到后台写完结果才继续。

---

## 四、招式三：简单计数用 atomic

对于单个整数的并发操作，`sync/atomic` 包提供了硬件级别的原子操作，**比 Mutex 更快**。

```go
package main

import (
    "fmt"
    "sync"
    "sync/atomic"
)

// ✅ atomic 方式：无锁，性能最好
var counter int64

func increment() {
    atomic.AddInt64(&counter, 1) // 原子操作，硬件保证安全
}

func getCounter() int64 {
    return atomic.LoadInt64(&counter) // 原子读取
}

func main() {
    var wg sync.WaitGroup
    for i := 0; i < 10000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            increment()
        }()
    }
    wg.Wait()
    fmt.Println(getCounter()) // 10000，正确且无锁
}
```

### atomic 能做什么 / 不能做什么

```go
// ✅ 适合：单一整数的加减读写
var count int64
atomic.AddInt64(&count, 1)
atomic.LoadInt64(&count)
atomic.StoreInt64(&count, 100)
atomic.CompareAndSwapInt64(&count, 100, 200) // CAS

// ✅ 适合：布尔标记
var ready int32
atomic.StoreInt32(&ready, 1)
if atomic.LoadInt32(&ready) == 1 {
    // 已就绪
}

// ✅ 适合：指针类型（atomic.Value）
var config atomic.Value
config.Store(&Config{Host: "localhost"})
cfg := config.Load().(*Config)

// ❌ 不适合：操作复杂数据结构（map、slice、struct 多个字段）
//     这种场景还是用 Channel 或 Mutex
```

---

## 五、三招完整对比 + 可运行的完整例子

下面是一个把所有模式整合在一起的例子：

```go
package main

import (
    "fmt"
    "sync"
    "sync/atomic"
    "time"
)

// ========== 招式1+2：Channel 通信（访问 map）==========

type Cache struct {
    data  map[string]string
    getCh chan getReq
    setCh chan setReq
    stop  chan struct{}
}

type getReq struct {
    key    string
    result chan string
}

type setReq struct {
    key   string
    value string
}

func NewCache() *Cache {
    c := &Cache{
        data:  make(map[string]string),
        getCh: make(chan getReq),
        setCh: make(chan setReq),
        stop:  make(chan struct{}),
    }
    go c.loop()
    return c
}

func (c *Cache) loop() {
    for {
        select {
        case req := <-c.getCh:
            req.result <- c.data[req.key]
        case req := <-c.setCh:
            c.data[req.key] = req.value
        case <-c.stop:
            return
        }
    }
}

func (c *Cache) Get(key string) string {
    req := getReq{key: key, result: make(chan string)}
    c.getCh <- req
    return <-req.result
}

func (c *Cache) Set(key, value string) {
    c.setCh <- setReq{key: key, value: value}
}

func (c *Cache) Close() {
    close(c.stop)
}

// ========== 招式3：atomic 计数 ==========

type Counter struct {
    value int64
}

func (c *Counter) Inc() {
    atomic.AddInt64(&c.value, 1)
}

func (c *Counter) Get() int64 {
    return atomic.LoadInt64(&c.value)
}

// ========== 对比：Mutex 方式 ==========

type MutexCache struct {
    mu   sync.Mutex
    data map[string]string
}

func NewMutexCache() *MutexCache {
    return &MutexCache{data: make(map[string]string)}
}

func (m *MutexCache) Get(key string) string {
    m.mu.Lock()
    defer m.mu.Unlock()
    return m.data[key]
}

func (m *MutexCache) Set(key, value string) {
    m.mu.Lock()
    defer m.mu.Unlock()
    m.data[key] = value
}

// ========== 压测对比 ==========

func main() {
    // Channel 方式
    cache := NewCache()
    var wg sync.WaitGroup
    start := time.Now()
    for i := 0; i < 10000; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            cache.Set("key", fmt.Sprintf("val-%d", n))
            cache.Get("key")
        }(i)
    }
    wg.Wait()
    fmt.Println("Channel 缓存耗时:", time.Since(start))
    cache.Close()

    // Mutex 方式
    mCache := NewMutexCache()
    start = time.Now()
    for i := 0; i < 10000; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            mCache.Set("key", fmt.Sprintf("val-%d", n))
            mCache.Get("key")
        }(i)
    }
    wg.Wait()
    fmt.Println("Mutex  缓存耗时:", time.Since(start))

    // atomic 方式
    var counter Counter
    start = time.Now()
    for i := 0; i < 10000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter.Inc()
        }()
    }
    wg.Wait()
    fmt.Println("atomic 计数耗时:", time.Since(start))
    fmt.Println("counter =", counter.Get())
}
```

---

## 六、务实原则：什么时候才用无锁？

```
┌─────────────────────────────────────────────────────┐
│                                                     │
│  默认：Mutex + map                                  │
│         │                                           │
│         ├── 简单、易懂、调试方便                     │
│         └── 99% 场景下性能足够                      │
│                                                     │
│  仅当 Mutex 成为性能瓶颈时：                         │
│         │                                           │
│         ├── 读写单一计数器 → atomic                  │
│         │                                           │
│         └── 复杂数据结构，锁竞争严重                  │
│             └── Channel 通信模型                    │
│                                                     │
└─────────────────────────────────────────────────────┘
```

**三个不要**：
- 不要为了"炫技"用无锁——Mutex 简单可靠，是务实的默认选择
- 不要在没 profiling 的情况下优化——先用 pprof 确认锁是瓶颈
- 不要过早优化——写出正确的代码比写出"酷"的代码重要得多

---

## 七、面试口述话术（30 秒版）

> "Go 的哲学是——不通过共享内存来通信，而通过通信来共享内存。实现无锁结构就是让一个后台 Goroutine 独享数据，外部通过 Channel 发消息来读写。比如一个缓存，外面发 getCh 请求带一个 result channel，后台 Goroutine select 收到后查 map 把结果塞回 result channel。这样整个系统只有一个 Goroutine 碰数据，自然不需要锁。对于简单计数用 atomic 就够了。但说实话，除非锁真的成了性能瓶颈，否则 Mutex 加 map 是最务实的方案。"

---

## 八、要点速记

**无锁并发三招**：
- Channel 通信模型——后台 Goroutine 独享数据，外部通过 Channel 发消息，只有一人碰数据，天然无锁
- 请求带 result channel——发 `getCh` 时附带 `chan result`，后台查完塞回去，调用方阻塞等结果
- 简单计数用 `atomic`——`atomic.AddInt64` / `atomic.LoadInt64` 足够，不用上锁

**务实原则**：
- 除非锁真的成了性能瓶颈，否则 `Mutex` + map 就是最好的方案
