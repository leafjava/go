# Goroutine 泄露：原理、场景与排查

> 面试场景：HashKey 一面（2026.05.25）实际考题之一
> 适用岗位：Go 后端 / Web3 全栈

---

## 一、什么是 Goroutine 泄露

### 1.1 一句话定义

**Goroutine 泄露 = 启动了 Goroutine，但它永远无法退出，持续占用内存和栈空间，直到进程结束。**

### 1.2 为什么是个大问题

Goroutine 虽然轻量（初始栈 2KB），但：
- 每个 Goroutine 都占内存（栈最小 2KB，可增长到 1GB）
- Goroutine 还持有它引用的所有变量，这些变量也无法被 GC 回收
- Goroutine 数量持续增长 → 内存膨胀 → OOM Killed

### 1.3 交易所场景的真实危害

```
场景：用户充值 10 个 ETH
  ↓
后端启动一个 Goroutine 监听这笔交易的链上确认数
  ↓
某个 RPC 节点返回慢导致 Goroutine 永远卡在 client.Call()
  ↓
用户看到链上 12 个确认了，平台余额还是 0
  ↓
用户慌了：交易所是不是要跑路？
  ↓
客服群炸锅 + 监管投诉 + 平台信任危机
```

**核心结论：在交易所这不是技术问题，是产品事故。**

---

## 二、Goroutine 泄露的本质

### 2.1 Goroutine 的两种状态

```
正常 Goroutine：
  启动 → 执行任务 → 退出 → 资源回收

泄露 Goroutine：
  启动 → 执行到某一步阻塞了 → 永远等待 → 永远不退出
```

### 2.2 阻塞在哪里？

只要 Goroutine 处于以下状态且无法被唤醒，就泄露了：

| 阻塞类型 | 唤醒条件 | 常见泄露原因 |
|---------|---------|-------------|
| Channel 发送 `ch <- v` | 有人接收 | 接收方没了 |
| Channel 接收 `<-ch` | 有人发送 | 发送方没了 |
| `select` 所有 case | 至少一个 case 就绪 | 所有 case 都阻塞 |
| `Mutex.Lock()` | 持锁者 Unlock | 锁忘了释放 |
| `WaitGroup.Wait()` | 计数器归零 | Add/Done 不平衡 |
| 网络 I/O（HTTP/DB） | 数据返回或超时 | 没设超时，对端不响应 |

---

## 三、五大泄露场景（含代码 + 修复）

### 3.1 场景一：Channel 发送端没有接收者

#### ❌ 错误代码

```go
package main

import (
    "fmt"
    "runtime"
    "time"
)

func leak1() {
    ch := make(chan int) // 无缓冲 channel
    go func() {
        // 这个 goroutine 想往 ch 里塞 42
        // 但无缓冲 channel 必须有人同时接收才能成功
        // 没人接收，永远卡在这里
        ch <- 42
        fmt.Println("永远到不了这里")
    }()
    // 函数返回，没人接收 ch 了
    // 上面那个 goroutine 永远阻塞在 ch <- 42
}

func main() {
    fmt.Println("启动前 goroutine 数:", runtime.NumGoroutine()) // 1
    for i := 0; i < 100; i++ {
        leak1()
    }
    time.Sleep(time.Second)
    fmt.Println("启动后 goroutine 数:", runtime.NumGoroutine()) // 101！全泄露了
}
```

**输出：**
```
启动前 goroutine 数: 1
启动后 goroutine 数: 101
```

#### ✅ 修复方案 1：buffered channel

```go
func fix1a() {
    ch := make(chan int, 1) // 缓冲 1，发送不阻塞
    go func() {
        ch <- 42 // 立即成功，goroutine 正常退出
    }()
}
```

#### ✅ 修复方案 2：select + 超时

```go
func fix1b() {
    ch := make(chan int)
    go func() {
        select {
        case ch <- 42:
            // 发送成功
        case <-time.After(3 * time.Second):
            // 3 秒还没人收，goroutine 主动退出
            return
        }
    }()
}
```

#### ✅ 修复方案 3：Context 取消

```go
func fix1c(ctx context.Context) {
    ch := make(chan int)
    go func() {
        select {
        case ch <- 42:
        case <-ctx.Done(): // 调用方取消，goroutine 退出
            return
        }
    }()
}
```

---

### 3.2 场景二：Channel 接收端没有发送者

#### ❌ 错误代码 — 真实交易所场景

```go
// 监听链上交易确认 — 但请求出错了
func watchTransaction(txHash string) {
    confirmCh := make(chan int)

    go func() {
        // 想从某个 RPC 调用拿确认数
        confirmations, err := callRPC(txHash)
        if err != nil {
            return // ⚠️ 这里直接 return 了，confirmCh 永远不会有数据
        }
        confirmCh <- confirmations
    }()

    // 主流程在等 confirmCh
    confirms := <-confirmCh // 上面 return 了，这里永远阻塞
    fmt.Println("确认数:", confirms)
}
```

#### ✅ 修复方案：errCh 双通道

```go
func watchTransactionFixed(ctx context.Context, txHash string) (int, error) {
    confirmCh := make(chan int, 1)
    errCh := make(chan error, 1)

    go func() {
        confirmations, err := callRPC(txHash)
        if err != nil {
            errCh <- err
            return
        }
        confirmCh <- confirmations
    }()

    select {
    case confirms := <-confirmCh:
        return confirms, nil
    case err := <-errCh:
        return 0, err
    case <-ctx.Done():
        return 0, ctx.Err() // 超时也能退出
    }
}
```

---

### 3.3 场景三：for range Channel 没人 close

#### ❌ 错误代码

```go
// Worker 模式 — 但生产者忘了 close
func workerLeak() {
    jobs := make(chan int, 5)

    // 生产者
    go func() {
        for i := 0; i < 5; i++ {
            jobs <- i
        }
        // ⚠️ 忘了 close(jobs)
    }()

    // 消费者
    go func() {
        for job := range jobs {
            // for range 在 channel close 之前永远不退出
            fmt.Println("处理任务:", job)
        }
        // 永远到不了这里
    }()

    time.Sleep(time.Second)
}
```

#### ✅ 修复：发送方负责 close

```go
func workerFixed() {
    jobs := make(chan int, 5)

    go func() {
        defer close(jobs) // ← 发送完毕一定要 close
        for i := 0; i < 5; i++ {
            jobs <- i
        }
    }()

    go func() {
        for job := range jobs {
            fmt.Println("处理任务:", job)
        }
        // close 之后 for range 自动退出
    }()
}
```

**黄金法则**：**永远是发送方 close Channel，接收方不能 close**

为什么？
- 接收方 close 了 → 发送方继续发会 panic
- 多个发送方时，谁来 close？协调成本高

---

### 3.4 场景四：select 缺少退出分支

#### ❌ 错误代码

```go
// 后台事件处理器 — 永远不退出
func eventLoopLeak(eventCh chan Event) {
    go func() {
        for {
            select {
            case event := <-eventCh:
                processEvent(event)
            // ⚠️ 没有 ctx.Done() 退出分支
            // 服务关闭时这个 goroutine 永远卡在这里
            }
        }
    }()
}
```

#### ✅ 修复：必加 ctx.Done()

```go
func eventLoopFixed(ctx context.Context, eventCh chan Event) {
    go func() {
        for {
            select {
            case event := <-eventCh:
                processEvent(event)
            case <-ctx.Done():
                // 服务关闭时优雅退出
                fmt.Println("event loop 退出")
                return
            }
        }
    }()
}
```

#### 应用：服务优雅关闭

```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    eventCh := make(chan Event)
    eventLoopFixed(ctx, eventCh)

    // 监听系统信号
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh

    cancel() // 通知所有子 goroutine 退出
    time.Sleep(time.Second) // 给它们时间清理
}
```

---

### 3.5 场景五：HTTP Body / DB 连接没关闭

#### ❌ 错误代码

```go
func fetchPriceLeak(url string) ([]byte, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    // ⚠️ 忘了 resp.Body.Close()
    // 底层 HTTP 连接无法回到连接池复用
    // 时间一长，连接池耗尽 + 后台 goroutine 等连接释放堆积

    return io.ReadAll(resp.Body)
}
```

#### ✅ 修复：永远 defer Close

```go
func fetchPriceFixed(url string) ([]byte, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close() // ← 关键！

    return io.ReadAll(resp.Body)
}
```

#### 还有这些资源也要 defer Close

```go
// 数据库 Rows
rows, err := db.Query("SELECT ...")
if err != nil { return err }
defer rows.Close() // ← 必须

// 文件
f, err := os.Open("foo.txt")
if err != nil { return err }
defer f.Close()

// gRPC 流
stream, err := client.Subscribe(ctx, &req)
if err != nil { return err }
defer stream.CloseSend()
```

---

### 3.6 [加分场景] time.After 在 for 循环里

#### ❌ 微妙的泄露

```go
// 看起来没问题的代码 — 实际有内存泄露
func tickerLeak() {
    for {
        select {
        case <-time.After(1 * time.Second):
            doWork()
        }
    }
    // 每次循环都 new 一个 Timer
    // 旧的 Timer 还没触发就进入下一轮
    // GC 无法回收（select case 还在引用）
    // 长时间运行内存持续增长
}
```

#### ✅ 修复：复用 Timer 或用 Ticker

```go
// 方案 1：复用 Timer
func tickerFixed1() {
    timer := time.NewTimer(1 * time.Second)
    defer timer.Stop()

    for {
        select {
        case <-timer.C:
            doWork()
            timer.Reset(1 * time.Second) // 复用
        }
    }
}

// 方案 2：直接用 Ticker（更推荐）
func tickerFixed2() {
    ticker := time.NewTicker(1 * time.Second)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            doWork()
        }
    }
}
```

---

## 四、排查 Goroutine 泄露

### 4.1 单元测试中检测

```go
package main

import (
    "runtime"
    "testing"
    "time"
)

func TestNoGoroutineLeak(t *testing.T) {
    before := runtime.NumGoroutine()

    // 执行可能泄露的代码
    yourFunction()

    // 给被测代码一点时间执行完毕
    time.Sleep(100 * time.Millisecond)

    after := runtime.NumGoroutine()
    if after > before {
        t.Errorf("goroutine 泄露: before=%d after=%d", before, after)
    }
}
```

**进阶：用 goleak 库（uber 出品）**

```bash
go get go.uber.org/goleak
```

```go
import "go.uber.org/goleak"

func TestMain(m *testing.M) {
    goleak.VerifyTestMain(m) // 所有测试结束后检测泄露
}

// 或者单个测试
func TestSomething(t *testing.T) {
    defer goleak.VerifyNone(t)
    // ... 测试代码 ...
}
```

### 4.2 线上用 pprof

#### 启用 pprof

```go
import (
    _ "net/http/pprof" // 副作用导入，自动注册路由
    "net/http"
)

func main() {
    go func() {
        // 仅在内网监听！别暴露到公网！
        http.ListenAndServe("localhost:6060", nil)
    }()

    // 你的业务代码
}
```

#### 抓取 goroutine 栈

```bash
# 命令行直接看
curl http://localhost:6060/debug/pprof/goroutine?debug=1

# 用 pprof 工具分析
go tool pprof http://localhost:6060/debug/pprof/goroutine
> top         # 看哪些函数 goroutine 数量异常
> list 函数名  # 看具体阻塞在哪一行
> web         # 生成调用图（需要 graphviz）
```

#### 实战：发现泄露的典型模式

```
看到这种栈就要警惕了：
  goroutine 2543 [chan receive, 60 minutes]:
      main.workerLoop(0xc0000a4000)
          /app/worker.go:42 +0x85

→ 2543 个 goroutine 卡在 chan receive
→ 已经卡了 60 分钟
→ 都阻塞在 worker.go:42 这一行
→ 大概率是 Channel 没人发数据
```

### 4.3 生产环境监控告警

```go
// 定时上报 goroutine 数量到监控系统
func StartGoroutineMonitor(ctx context.Context) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    baseline := runtime.NumGoroutine()
    threshold := baseline * 3 // 翻 3 倍就告警

    for {
        select {
        case <-ticker.C:
            current := runtime.NumGoroutine()

            // 上报到 Prometheus / OpenTelemetry
            metrics.Gauge("go.goroutines.count").Set(float64(current))

            if current > threshold {
                // 触发告警 + 自动 dump 现场
                log.Errorf("⚠️ Goroutine 数量异常: baseline=%d current=%d", baseline, current)
                dumpGoroutineProfile()
            }
        case <-ctx.Done():
            return
        }
    }
}

func dumpGoroutineProfile() {
    f, _ := os.Create(fmt.Sprintf("goroutine-%d.prof", time.Now().Unix()))
    defer f.Close()
    pprof.Lookup("goroutine").WriteTo(f, 1)
}
```

---

## 五、Web3/交易所场景的实战案例

### 5.1 案例：RPC 调用没超时导致 Goroutine 雪崩

#### 真实事故还原

```go
// ❌ 高峰期某个 RPC 节点挂了
func queryBalanceBad(client *rpc.Client, addr string) (string, error) {
    var result string
    // 没设超时！节点挂了 client.Call 永远不返回
    err := client.Call(&result, "eth_getBalance", addr, "latest")
    return result, err
}

// 调用方：每个用户请求开一个 goroutine
func handleUserRequest(addr string) {
    go func() {
        balance, _ := queryBalanceBad(client, addr) // 永远卡住
        fmt.Println(balance)
    }()
}

// 高峰期 2000 个用户请求 → 2000 个 goroutine 全卡死
// 内存涨 2GB → OOM Killed
```

#### 修复

```go
// ✅ 加超时 + 熔断
type Client struct {
    rpc      *rpc.Client
    breaker  *circuit.Breaker // 熔断器
}

func (c *Client) QueryBalance(ctx context.Context, addr string) (string, error) {
    // 1. 熔断检查
    if !c.breaker.Allow() {
        return "", errors.New("熔断中，跳过 RPC")
    }

    // 2. 设置超时
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    // 3. 异步调用 + select 超时
    type result struct {
        balance string
        err     error
    }
    resCh := make(chan result, 1)

    go func() {
        var balance string
        err := c.rpc.CallContext(ctx, &balance, "eth_getBalance", addr, "latest")
        resCh <- result{balance, err}
    }()

    select {
    case r := <-resCh:
        if r.err != nil {
            c.breaker.RecordFailure()
        } else {
            c.breaker.RecordSuccess()
        }
        return r.balance, r.err
    case <-ctx.Done():
        c.breaker.RecordFailure()
        return "", ctx.Err()
    }
}
```

### 5.2 案例：充值监听器漏掉退出信号

```go
// ❌ 服务关闭时这个 goroutine 不退出
type DepositMonitor struct {
    client *ethclient.Client
}

func (m *DepositMonitor) WatchBlocks() {
    headerCh := make(chan *types.Header)
    sub, _ := m.client.SubscribeNewHead(context.Background(), headerCh)

    go func() {
        for {
            select {
            case header := <-headerCh:
                m.processBlock(header)
            case err := <-sub.Err():
                log.Printf("订阅错误: %v", err)
                // ⚠️ 错误了但没 return，下次循环继续等
                // ⚠️ 也没有 ctx.Done() 退出路径
            }
        }
    }()
}

// ✅ 修复
func (m *DepositMonitor) WatchBlocksFixed(ctx context.Context) error {
    headerCh := make(chan *types.Header)
    sub, err := m.client.SubscribeNewHead(ctx, headerCh)
    if err != nil {
        return err
    }
    defer sub.Unsubscribe() // 退出时取消订阅

    go func() {
        for {
            select {
            case header := <-headerCh:
                m.processBlock(ctx, header)
            case err := <-sub.Err():
                log.Printf("订阅错误，重连: %v", err)
                return // 让上层重连
            case <-ctx.Done():
                log.Println("监听器优雅退出")
                return
            }
        }
    }()
    return nil
}
```

---

## 六、记忆口诀

### 6.1 五大泄露场景

```
发不出（无人收）
收不到（无人发）
range 不 close（永远循环）
select 没 ctx.Done（无退出路径）
defer 忘 Close（资源未释放）
```

### 6.2 三条铁律

```
1. 谁发数据，谁负责 close(ch)
2. 每个 goroutine 都必须有 ctx.Done() 退出路径
3. 资源用 defer Close（Body / Rows / File / Stream）
```

### 6.3 排查三板斧

```
1. 单测：runtime.NumGoroutine() 测试前后对比
2. 离线：go.uber.org/goleak 自动检测
3. 线上：pprof goroutine profile + 监控告警
```

---

## 七、面试话术速记

### 7.1 技术视角（标准版）

> "Goroutine 泄露的本质是 Goroutine 卡在某个阻塞操作上永远等不到退出信号。最常见五个场景——Channel 无人收、Channel 无人发、for range 没 close、select 缺 ctx.Done()、HTTP/DB 资源没 Close。排查用 pprof 的 goroutine profile，或测试时对比 runtime.NumGoroutine()。铁律三句话：发送方 close、每个 goroutine 必有 ctx.Done()、资源 defer Close。"

### 7.2 产品视角（HashKey 加分版）

> "Goroutine 泄露在交易所不是技术问题，是用户事故。比如用户充了 10 ETH，链上 12 确认了，但监听确认数的 Goroutine 卡死了，用户余额一直是 0——他第一反应是'交易所要跑路'，不是'程序有 bug'。所以这不是加分项，是底线。常见就五类场景，排查用 pprof 看阻塞栈。铁律是发送方 close、ctx.Done() 退出、资源 defer Close——确保用户的每一分钱入账都不会因为一个 Goroutine 卡死而延迟。"

---

## 八、扩展阅读

### 相关知识点
- [[Q4: 高并发 + GC]] — Goroutine 泄露常常伴随内存泄露
- [[pprof 性能分析]] — Goroutine profile 是 pprof 的一种
- [[Context 使用规范]] — 正确传递取消信号

### 推荐资料
- Go 官方 Blog: [Go Concurrency Patterns: Pipelines and cancellation](https://go.dev/blog/pipelines)
- Uber 工程: [go.uber.org/goleak](https://github.com/uber-go/goleak)
- Dave Cheney: [Never start a goroutine without knowing how it will stop](https://dave.cheney.net/2016/12/22/never-start-a-goroutine-without-knowing-how-it-will-stop)
