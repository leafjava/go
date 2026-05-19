# 第7课：并发编程：Goroutine + Channel

> 学习时间：3-4小时 | 难度：⭐⭐⭐⭐

## 📋 本课目标

- 理解 Goroutine 的概念
- 掌握 Channel 的使用
- 学会使用 select 处理多个 Channel
- 理解并发安全和同步

## 1. Goroutine（协程）

### 基本使用

```go
package main

import (
    "fmt"
    "time"
)

func sayHello(name string) {
    for i := 0; i < 3; i++ {
        fmt.Printf("Hello, %s! (%d)\n", name, i)
        time.Sleep(100 * time.Millisecond)
    }
}

func main() {
    // 普通调用（同步）
    sayHello("Alice")
    
    // Goroutine 调用（异步）
    go sayHello("Bob")
    go sayHello("Charlie")
    
    // 等待 goroutine 完成
    time.Sleep(500 * time.Millisecond)
    fmt.Println("主函数结束")
}
```

### 匿名 Goroutine

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    // 匿名函数 + goroutine
    go func() {
        fmt.Println("匿名 goroutine 执行")
    }()
    
    // 带参数的匿名 goroutine
    name := "林燊"
    go func(n string) {
        fmt.Println("Hello,", n)
    }(name)
    
    time.Sleep(100 * time.Millisecond)
}
```

## 2. Channel（通道）⭐⭐⭐

### 基本操作

```go
package main

import "fmt"

func main() {
    // 创建 channel
    ch := make(chan int)
    
    // 发送数据（在 goroutine 中）
    go func() {
        ch <- 42  // 发送
    }()
    
    // 接收数据
    value := <-ch  // 接收
    fmt.Println("接收到:", value)
    
    // 带缓冲的 channel
    bufferedCh := make(chan string, 3)
    bufferedCh <- "A"
    bufferedCh <- "B"
    bufferedCh <- "C"
    
    fmt.Println(<-bufferedCh)  // A
    fmt.Println(<-bufferedCh)  // B
    fmt.Println(<-bufferedCh)  // C
}
```

### Channel 方向

```go
package main

import "fmt"

// 只发送 channel
func sendOnly(ch chan<- int) {
    ch <- 100
}

// 只接收 channel
func receiveOnly(ch <-chan int) {
    value := <-ch
    fmt.Println("接收:", value)
}

func main() {
    ch := make(chan int)
    
    go sendOnly(ch)
    receiveOnly(ch)
}
```

### 关闭 Channel

```go
package main

import "fmt"

func main() {
    ch := make(chan int, 3)
    
    // 发送数据
    ch <- 1
    ch <- 2
    ch <- 3
    close(ch)  // 关闭 channel
    
    // 接收所有数据
    for value := range ch {
        fmt.Println(value)
    }
    
    // 检查 channel 是否关闭
    value, ok := <-ch
    if !ok {
        fmt.Println("Channel 已关闭")
    }
}
```

## 3. Web3 实战：并发查询区块链

### 并发查询多个地址余额

```go
package main

import (
    "fmt"
    "math/rand"
    "time"
)

type Balance struct {
    Address string
    Amount  float64
    Error   error
}

// 模拟查询余额（耗时操作）
func queryBalance(address string) (float64, error) {
    time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)
    return rand.Float64() * 10, nil
}

// 并发查询
func queryBalancesConcurrent(addresses []string) []Balance {
    results := make(chan Balance, len(addresses))
    
    // 启动多个 goroutine
    for _, addr := range addresses {
        go func(address string) {
            amount, err := queryBalance(address)
            results <- Balance{
                Address: address,
                Amount:  amount,
                Error:   err,
            }
        }(addr)
    }
    
    // 收集结果
    balances := make([]Balance, 0, len(addresses))
    for i := 0; i < len(addresses); i++ {
        balances = append(balances, <-results)
    }
    
    return balances
}

func main() {
    addresses := []string{
        "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
        "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
        "0x1234567890123456789012345678901234567890",
    }
    
    start := time.Now()
    balances := queryBalancesConcurrent(addresses)
    fmt.Printf("查询完成，耗时: %v\n", time.Since(start))
    
    for _, b := range balances {
        fmt.Printf("%s: %.2f ETH\n", b.Address, b.Amount)
    }
}
```

### Worker Pool 模式

```go
package main

import (
    "fmt"
    "time"
)

type Job struct {
    ID   int
    Data string
}

type Result struct {
    Job    Job
    Result string
}

// Worker 函数
func worker(id int, jobs <-chan Job, results chan<- Result) {
    for job := range jobs {
        fmt.Printf("Worker %d 处理任务 %d\n", id, job.ID)
        time.Sleep(100 * time.Millisecond)
        
        results <- Result{
            Job:    job,
            Result: fmt.Sprintf("处理完成: %s", job.Data),
        }
    }
}

func main() {
    jobs := make(chan Job, 10)
    results := make(chan Result, 10)
    
    // 启动 3 个 worker
    for w := 1; w <= 3; w++ {
        go worker(w, jobs, results)
    }
    
    // 发送 9 个任务
    for j := 1; j <= 9; j++ {
        jobs <- Job{
            ID:   j,
            Data: fmt.Sprintf("交易 #%d", j),
        }
    }
    close(jobs)
    
    // 收集结果
    for a := 1; a <= 9; a++ {
        result := <-results
        fmt.Println(result.Result)
    }
}
```

## 4. Select 语句

### 基本用法

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    ch1 := make(chan string)
    ch2 := make(chan string)
    
    go func() {
        time.Sleep(100 * time.Millisecond)
        ch1 <- "来自 ch1"
    }()
    
    go func() {
        time.Sleep(200 * time.Millisecond)
        ch2 <- "来自 ch2"
    }()
    
    // select 等待多个 channel
    for i := 0; i < 2; i++ {
        select {
        case msg1 := <-ch1:
            fmt.Println("接收:", msg1)
        case msg2 := <-ch2:
            fmt.Println("接收:", msg2)
        }
    }
}
```

### 超时控制

```go
package main

import (
    "fmt"
    "time"
)

func queryWithTimeout(address string, timeout time.Duration) (float64, error) {
    result := make(chan float64)
    
    go func() {
        time.Sleep(500 * time.Millisecond)  // 模拟耗时操作
        result <- 10.5
    }()
    
    select {
    case balance := <-result:
        return balance, nil
    case <-time.After(timeout):
        return 0, fmt.Errorf("查询超时")
    }
}

func main() {
    // 超时时间 1 秒
    balance, err := queryWithTimeout("0x742d35Cc...", 1*time.Second)
    if err != nil {
        fmt.Println("错误:", err)
    } else {
        fmt.Printf("余额: %.2f ETH\n", balance)
    }
    
    // 超时时间 100 毫秒
    balance, err = queryWithTimeout("0x742d35Cc...", 100*time.Millisecond)
    if err != nil {
        fmt.Println("错误:", err)
    } else {
        fmt.Printf("余额: %.2f ETH\n", balance)
    }
}
```

## 5. sync.WaitGroup

```go
package main

import (
    "fmt"
    "sync"
    "time"
)

func processTransaction(id int, wg *sync.WaitGroup) {
    defer wg.Done()  // 完成时调用
    
    fmt.Printf("处理交易 %d\n", id)
    time.Sleep(100 * time.Millisecond)
    fmt.Printf("交易 %d 完成\n", id)
}

func main() {
    var wg sync.WaitGroup
    
    // 启动 5 个 goroutine
    for i := 1; i <= 5; i++ {
        wg.Add(1)  // 计数器 +1
        go processTransaction(i, &wg)
    }
    
    wg.Wait()  // 等待所有 goroutine 完成
    fmt.Println("所有交易处理完成")
}
```

## 6. sync.Mutex（互斥锁）

```go
package main

import (
    "fmt"
    "sync"
)

type SafeCounter struct {
    mu    sync.Mutex
    count int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.count++
}

func (c *SafeCounter) GetCount() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.count
}

func main() {
    counter := &SafeCounter{}
    var wg sync.WaitGroup
    
    // 启动 1000 个 goroutine 同时增加计数
    for i := 0; i < 1000; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            counter.Increment()
        }()
    }
    
    wg.Wait()
    fmt.Println("最终计数:", counter.GetCount())  // 1000
}
```

## 📝 作业

### 作业1：并发交易处理器

创建 `homework/day7/concurrent_tx_processor.go`：

```go
package main

type Transaction struct {
    Hash   string
    Amount float64
}

// TODO: 实现并发处理交易
func ProcessTransactions(txs []Transaction, workers int) []string {
    // 使用 worker pool 模式处理交易
    // 返回处理结果（交易哈希列表）
    return nil
}

func main() {
    // 测试并发处理
}
```

### 作业2：区块链数据聚合器

创建 `homework/day7/blockchain_aggregator.go`：

```go
package main

// TODO: 并发查询多条链的数据
func AggregateBalances(addresses []string, chains []string) map[string]map[string]float64 {
    // 返回：地址 -> (链名 -> 余额)
    // 使用 goroutine + channel 并发查询
    return nil
}

func main() {
    // 测试聚合器
}
```

### 作业3：实时事件监听器

创建 `homework/day7/event_listener.go`：

```go
package main

// TODO: 实现事件监听器
type EventListener struct {
    // events chan Event
}

// TODO: 实现方法
// 1. Start() - 启动监听
// 2. Stop() - 停止监听
// 3. Subscribe(callback func(Event)) - 订阅事件

func main() {
    // 测试事件监听
}
```

## 🎯 检查点

- ✅ 理解 Goroutine 和 Channel
- ✅ 能够使用 select 处理多个 Channel
- ✅ 掌握 WaitGroup 和 Mutex
- ✅ 能够实现并发模式（Worker Pool 等）

## 💡 重点提示

1. **Goroutine 很轻量**，可以创建成千上万个
2. **Channel 是 Go 并发的核心**，用于 goroutine 间通信
3. **避免共享内存**，优先使用 Channel 通信
4. **使用 Mutex 保护共享数据**

## ⏭️ 下一课

[第8课：Gin 框架入门](../week2/day8-gin-intro.md)

---

**🎉 恭喜！你已经完成了第一周的学习！**
