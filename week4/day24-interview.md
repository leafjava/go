# 第24课：Go 面试高频题

> 精选 Web3 后端开发最常见的 Go 面试题 | 理论 + 代码实战

## 📋 本课目标

- 掌握 Go 面试中的高频考点
- 理解底层原理，不只是背答案
- 能用代码演示关键概念
- 为 Web3 后端面试做准备

---

## 1️⃣ 基础语法类

### Q1: Go 的三种变量声明方式有什么区别？

```go
// 方式1：完整声明（可用于全局变量）
var name string = "Alice"

// 方式2：类型推断（可用于全局变量）
var age = 25

// 方式3：短声明（只能在函数内使用）
func main() {
    city := "Beijing"  // 最常用
}
```

**面试要点**：
- `:=` 只能在函数内使用
- `var` 可以声明全局变量
- 短声明更简洁，是 Go 的惯用写法

---

### Q2: `new` 和 `make` 的区别？

```go
// new：分配内存并返回指针，值为零值
p := new(int)        // *int 类型，值为 0
fmt.Println(*p)      // 0

// make：只用于 slice、map、channel，返回初始化后的值
s := make([]int, 5)  // []int 类型，长度为5
m := make(map[string]int)
ch := make(chan int)
```

**记忆口诀**：
- `new` → 任何类型，返回指针
- `make` → 只能 slice/map/channel，返回值本身

---

### Q3: 值传递 vs 引用传递？

```go
// Go 只有值传递！但有些类型"看起来"像引用传递

// 1. 基本类型：真正的值传递
func changeInt(x int) {
    x = 100  // 不影响外部
}

// 2. 指针：传递的是指针的副本，但指向同一地址
func changePointer(p *int) {
    *p = 100  // 会影响外部
}

// 3. slice/map/channel：传递的是底层数据结构的引用
func changeSlice(s []int) {
    s[0] = 100  // 会影响外部
}
```

**面试要点**：Go 没有引用传递，但 slice/map/channel 内部包含指针。

---

## 2️⃣ 并发编程类（高频！）

### Q4: Goroutine 和线程的区别？

| 特性 | Goroutine | 线程 |
|------|-----------|------|
| 内存占用 | 2KB 起步 | 1-2MB |
| 调度方式 | Go 运行时调度（M:N） | 操作系统调度 |
| 切换成本 | 低（用户态） | 高（内核态） |
| 数量 | 可创建百万级 | 通常几千个 |

```go
// 创建 10000 个 goroutine 很轻松
for i := 0; i < 10000; i++ {
    go func(id int) {
        fmt.Println(id)
    }(i)
}
```

---

### Q5: Channel 的三种操作会阻塞吗？

```go
ch := make(chan int)

// 1. 发送：无缓冲 channel 会阻塞，直到有接收者
ch <- 42  // 阻塞

// 2. 接收：会阻塞，直到有数据
val := <-ch  // 阻塞

// 3. 关闭：不会阻塞
close(ch)

// 带缓冲的 channel
ch2 := make(chan int, 3)
ch2 <- 1  // 不阻塞（缓冲未满）
ch2 <- 2
ch2 <- 3
ch2 <- 4  // 阻塞（缓冲已满）
```

**面试要点**：
- 无缓冲 channel：发送和接收都会阻塞
- 有缓冲 channel：缓冲满时发送阻塞，缓冲空时接收阻塞

---

### Q6: 如何避免 Goroutine 泄漏？

```go
// ❌ 错误：goroutine 永远不会退出
func leak() {
    ch := make(chan int)
    go func() {
        val := <-ch  // 永远阻塞，因为没有发送者
        fmt.Println(val)
    }()
}

// ✅ 正确：使用 context 控制生命周期
func noLeak(ctx context.Context) {
    ch := make(chan int)
    go func() {
        select {
        case val := <-ch:
            fmt.Println(val)
        case <-ctx.Done():
            return  // 超时或取消时退出
        }
    }()
}
```

---

### Q7: `sync.Mutex` vs `sync.RWMutex`？

```go
// Mutex：读写都互斥
var mu sync.Mutex
mu.Lock()
// 临界区
mu.Unlock()

// RWMutex：读读不互斥，读写互斥
var rwMu sync.RWMutex

// 读锁（多个 goroutine 可同时持有）
rwMu.RLock()
// 读操作
rwMu.RUnlock()

// 写锁（独占）
rwMu.Lock()
// 写操作
rwMu.Unlock()
```

**使用场景**：
- 读多写少 → `RWMutex`
- 读写均衡 → `Mutex`

---

## 3️⃣ 数据结构类

### Q8: Slice 的底层结构？

```go
type slice struct {
    array unsafe.Pointer  // 指向底层数组
    len   int             // 当前长度
    cap   int             // 容量
}

// 扩容机制
s := make([]int, 0, 4)
fmt.Println(len(s), cap(s))  // 0 4

s = append(s, 1, 2, 3, 4)
fmt.Println(len(s), cap(s))  // 4 4

s = append(s, 5)  // 触发扩容
fmt.Println(len(s), cap(s))  // 5 8（容量翻倍）
```

**面试要点**：
- Slice 不是数组，是对数组的引用
- 扩容规则：容量 < 1024 时翻倍，>= 1024 时增长 25%

---

### Q9: Map 是线程安全的吗？

```go
// ❌ 普通 map 不是线程安全的
m := make(map[string]int)
go func() { m["key"] = 1 }()
go func() { m["key"] = 2 }()  // 可能 panic

// ✅ 方案1：使用 sync.Map
var sm sync.Map
sm.Store("key", 1)
val, ok := sm.Load("key")

// ✅ 方案2：使用 Mutex 保护
var mu sync.Mutex
mu.Lock()
m["key"] = 1
mu.Unlock()
```

---

### Q10: 如何判断 Map 的 Key 是否存在？

```go
m := map[string]int{"age": 25}

// ❌ 错误：无法区分"不存在"和"值为0"
val := m["age"]

// ✅ 正确：使用 comma ok 模式
val, ok := m["age"]
if ok {
    fmt.Println("存在:", val)
} else {
    fmt.Println("不存在")
}
```

---

## 4️⃣ 接口与反射类

### Q11: 接口的底层结构？

```go
// 空接口（interface{}）
type eface struct {
    _type *_type          // 类型信息
    data  unsafe.Pointer  // 数据指针
}

// 非空接口
type iface struct {
    tab  *itab           // 类型 + 方法表
    data unsafe.Pointer  // 数据指针
}
```

**面试要点**：
- 接口变量包含：类型信息 + 数据指针
- `nil` 接口：类型和数据都为 `nil`

---

### Q12: 接口的 `nil` 判断陷阱？

```go
func returnsError() error {
    var p *MyError = nil
    return p  // 返回的不是 nil！
}

func main() {
    err := returnsError()
    if err != nil {  // true！
        fmt.Println("有错误")  // 会执行
    }
}
```

**原因**：接口包含类型信息，即使数据为 `nil`，类型不为 `nil`。

**解决方案**：
```go
func returnsError() error {
    var p *MyError = nil
    if p == nil {
        return nil  // 显式返回 nil
    }
    return p
}
```

---

## 5️⃣ 错误处理类

### Q13: `panic` 和 `recover` 的使用场景？

```go
func safeDivide(a, b int) (result int) {
    defer func() {
        if r := recover(); r != nil {
            fmt.Println("捕获 panic:", r)
            result = 0
        }
    }()
    
    return a / b  // b=0 时会 panic
}

func main() {
    fmt.Println(safeDivide(10, 0))  // 输出: 0
}
```

**面试要点**：
- `panic` 用于不可恢复的错误（如数组越界）
- `recover` 只能在 `defer` 中使用
- 生产代码应该用 `error` 而不是 `panic`

---

## 6️⃣ 性能优化类

### Q14: 如何避免内存逃逸？

```go
// ❌ 逃逸到堆：返回局部变量的指针
func escape() *int {
    x := 42
    return &x  // x 逃逸到堆
}

// ✅ 不逃逸：返回值
func noEscape() int {
    x := 42
    return x  // x 在栈上
}

// 查看逃逸分析
// go build -gcflags="-m" main.go
```

**面试要点**：
- 栈分配比堆分配快
- 返回指针、闭包、接口赋值可能导致逃逸

---

### Q15: `defer` 的性能开销？

```go
// defer 有一定开销，但 Go 1.14+ 已优化

// 高频路径避免 defer
func fastPath() {
    mu.Lock()
    // 快速操作
    mu.Unlock()
}

// 复杂逻辑使用 defer（可读性更重要）
func complexPath() {
    mu.Lock()
    defer mu.Unlock()
    // 多个 return 路径
}
```

---

## 7️⃣ Web3 相关（加分项）

### Q16: 如何安全处理私钥？

```go
// ❌ 错误：硬编码私钥
privateKey := "0x1234..."

// ✅ 正确：从环境变量读取
privateKey := os.Getenv("PRIVATE_KEY")
if privateKey == "" {
    log.Fatal("PRIVATE_KEY not set")
}

// ✅ 更好：使用 KMS 或 Vault
```

---

### Q17: 如何处理区块链的并发请求？

```go
// 使用 worker pool 限制并发
func processTransactions(txs []string) {
    const workers = 10
    sem := make(chan struct{}, workers)
    
    for _, tx := range txs {
        sem <- struct{}{}  // 获取令牌
        go func(txHash string) {
            defer func() { <-sem }()  // 释放令牌
            // 处理交易
        }(tx)
    }
}
```

---

## 🎯 面试准备建议

### 1. 必须掌握的知识点
- ✅ Goroutine 和 Channel
- ✅ Slice 和 Map 的底层原理
- ✅ 接口的使用和陷阱
- ✅ 错误处理最佳实践
- ✅ 并发安全（Mutex、RWMutex、sync.Map）

### 2. 加分项
- Context 的使用场景
- 性能优化技巧（逃逸分析、内存对齐）
- Go Modules 和依赖管理
- 单元测试和基准测试

### 3. 实战演示
面试时可能要求现场写代码，准备这些：
- 实现一个线程安全的缓存
- 用 Channel 实现生产者-消费者模式
- 处理超时和取消的 HTTP 请求

---

## 📝 今日作业

### 作业1：实现线程安全的计数器
```go
type SafeCounter struct {
    // TODO: 添加字段
}

func (c *SafeCounter) Inc() {
    // TODO: 实现
}

func (c *SafeCounter) Value() int {
    // TODO: 实现
}
```

### 作业2：用 Channel 实现超时控制
```go
func fetchWithTimeout(url string, timeout time.Duration) (string, error) {
    // TODO: 实现
    // 提示：使用 select + time.After
}
```

### 作业3：找出以下代码的问题
```go
func buggyCode() {
    var wg sync.WaitGroup
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            fmt.Println(i)  // 问题在哪？
        }()
    }
    wg.Wait()
}
```

---

## 🔗 扩展阅读

- [Go 面试题汇总](https://github.com/lifei6671/interview-go)
- [Go 语言高性能编程](https://geektutu.com/post/high-performance-go.html)
- [Effective Go](https://go.dev/doc/effective_go)

---

**下一课**：[第25-28课：毕业项目](./day25-28-final-project.md)

