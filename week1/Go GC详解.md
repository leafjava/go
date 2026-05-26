# Go GC（垃圾回收）详解：从零开始理解

> 面向读者：了解基本编程概念，但没接触过 Go 和 GC
> 目标：理解 GC 是什么、为什么要优化、怎么优化

---

## 一、GC 是什么？

### 1.1 生活中的类比

想象你在餐厅吃饭：

```
❌ 没有 GC（C/C++ 手动管理）：
  你吃完一盘菜 → 必须自己把盘子端回厨房 → 忘了端？桌子堆满盘子 → 没地方放新菜

✅ 有 GC（Go/Java/Python 自动管理）：
  你吃完一盘菜 → 服务员自动收走 → 你只管吃，不用管盘子去哪
```

**GC（Garbage Collection，垃圾回收）就是那个"服务员"——自动回收不再使用的内存，程序员不用手动释放。**

### 1.2 为什么需要 GC？

```go
// 你写代码时，一直在"创建"东西
func processOrder() {
    user := getUser()           // 创建 user 对象，占一块内存
    order := createOrder(user)  // 创建 order 对象，又占一块内存
    result := saveToDB(order)   // 创建 result 对象，再占一块内存

    fmt.Println(result)
    // 函数结束，user、order、result 都不再需要了
    // 但这些内存还占着呢！谁来清理？
    // → GC 来清理
}
```

如果没有 GC，每个变量用完后你都得手动写 `free(user)`、`free(order)`、`free(result)`。忘一个就泄露一块内存，程序越跑越慢，最后崩溃。

### 1.3 GC 的工作流程（简化版）

```
第一步：标记（Mark）
  扫描所有变量，标记哪些还在用（活的），哪些不用了（垃圾）

  内存：[user(在用)] [order(不用了)] [result(不用了)] [config(在用)]
               ✅              ❌              ❌              ✅

第二步：清理（Sweep）
  把标记为"不用了"的内存回收，腾出空间

  内存：[user(在用)] [  空了  ] [  空了  ] [config(在用)]

第三步：下次有新对象要创建，就用这些空位
```

这就是 Go 的**三色标记法**——但面试说"并发三色标记"就行，细节面试官很少追问。

---

## 二、GC 的问题在哪？

### 2.1 核心矛盾：GC 本身也要消耗 CPU

```
你的程序在做正事：处理订单、计算价格、推送行情...
        ↓
    GC 突然启动："等等！我先扫一遍内存！"
        ↓
    你的程序被暂停（STW - Stop The World）
        ↓
    GC 扫描完 → 程序继续
        ↓
    用户感觉：卡了一下
```

**Go 的 GC 已经很优秀了（STW 通常 < 1 毫秒），但如果 GC 太频繁，累计起来就很可观。**

### 2.2 GC 什么时候触发？

Go 默认：**堆内存增长 100% 就触发一次 GC**（这个比例叫 GOGC，默认值是 100）

```go
// 例子：假设当前堆内存是 10MB
// GOGC=100 意味着：堆增长到 20MB（100% 增长）就触发 GC

// 如果你的程序每秒分配 50MB 内存
// → 堆很快涨到 10MB → 20MB → GC
// → 清理后回到 10MB → 又涨到 20MB → 又 GC
// → 频繁 GC → CPU 被 GC 吃掉 → 你的业务逻辑变慢
```

### 2.3 代码感受一下

```go
package main

import (
    "fmt"
    "runtime"
    "time"
)

func main() {
    // 看 GC 默认会怎样

    fmt.Println("=== 开始疯狂分配内存 ===")

    for i := 0; i < 10; i++ {
        // 每次都创建大量垃圾数据
        waste := make([]byte, 10*1024*1024) // 一次分配 10MB
        _ = waste // 用完就扔掉

        // 打印当前 GC 状态
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        fmt.Printf("第 %d 次 | GC 累计执行了 %d 次 | 堆上有 %.2f MB\n",
            i+1, m.NumGC, float64(m.HeapAlloc)/1024/1024)

        time.Sleep(100 * time.Millisecond)
    }
}
```

运行这个程序你会看到 `NumGC` 不断增长——GC 在背后默默做了很多次回收。

---

## 三、为什么交易所场景对 GC 特别敏感？

### 3.1 场景：行情推送

```
正常时刻：每秒 100 条价格更新
  每次更新创建 PriceTick 对象 → 每秒 100 次分配
  GC 每分钟触发一两次 → 停顿 < 1ms → 用户无感

极端行情：BTC 一分钟涨 5%
  每秒 5000 条价格更新
  每次更新创建 PriceTick 对象 → 每秒 5000 次分配
  GC 每几秒就触发一次 → 停顿累积 → 用户感觉"怎么卡了？"
```

### 3.2 卡一下的代价

```
用户在 BTC=$70,000 时点市价买入
  ↓
GC 突然触发，停顿 200ms
  ↓
交易执行时 BTC 已经 $70,140
  ↓
用户多花了 $140
  ↓
用户投诉："你们平台怎么这么坑？"
```

**这就是 GC 优化的意义——不是为了 benchmark 数据好看，是为了用户的真金白银。**

---

## 四、GC 优化四招（面试速记 + 详细讲解）

### 第一招：对象池复用（sync.Pool）

#### 问题

```go
// ❌ 每秒创建 5000 个 PriceTick → 每秒产生 5000 个"垃圾"
// GC 得不停地扫、不停地清
func processPrice(price float64) {
    tick := &PriceTick{  // ← 每次 new 一个新对象
        Symbol: "BTC/USDT",
        Price:  price,
        Time:   time.Now(),
    }
    broadcast(tick)
    // tick 用完了，没人引用 → 变成垃圾 → 等 GC 来收
}
```

#### 解决：sync.Pool — 对象的"共享充电宝"

```go
// 创建一个 PriceTick 的对象池
// 就像共享充电宝：用完还回去，下次别人接着用，不用买新的
var tickPool = sync.Pool{
    New: func() any {
        return &PriceTick{} // 池子里没有可用的了，才 new 一个
    },
}

func processPrice(price float64) {
    tick := tickPool.Get().(*PriceTick) // 从池子里借一个
    tick.Symbol = "BTC/USDT"
    tick.Price = price
    tick.Time = time.Now()

    broadcast(tick)

    tickPool.Put(tick) // 用完还回去！
    // 下次有请求来，直接复用这个对象，不用创建新的
    // → 分配量减少 90% → GC 压力大降
}
```

**效果**：之前每秒创建 5000 个对象，现在只创建几个（池子里的来回用）。GC 工作量大减。

---

### 第二招：切片预分配（make 带 cap）

#### 问题

```go
// ❌ 没指定 cap，slice 按需增长
// 添加第 1 个元素 → 分配容量 1
// 添加第 2 个元素 → 容量不够！分配容量 2，老数据拷过去，老的扔掉（变垃圾）
// 添加第 3 个元素 → 容量不够！分配容量 4，又拷一遍又扔掉旧的
// 添加第 5 个元素 → 容量不够！分配容量 8，再拷再扔
// ...
// 1000 个元素，背后扩容了 ~10 次，产生了 ~10 份废弃内存（垃圾）
func buildList() []int {
    var result []int // cap=0
    for i := 0; i < 1000; i++ {
        result = append(result, i) // 每次都可能触发扩容
    }
    return result
}
```

#### 解决：一次分配到位

```go
// ✅ 提前给 cap："我要装 1000 个，一次给我分够"
func buildList() []int {
    //                      ↓ 长度 0，容量 1000
    result := make([]int, 0, 1000)
    // 内存一次分配到位，后面 append 直接往里填，不再扩容
    // → 没有反复分配 → 没有额外垃圾
    for i := 0; i < 1000; i++ {
        result = append(result, i)
    }
    return result
}
```

**效果**：1000 次 append，之前可能分配 10 几次，现在只分配 1 次。垃圾少了 GC 就轻松。

---

### 第三招：pprof 找分配大户

这是**诊断工具**，不是优化手段——先找出谁在疯狂分配，再针对它优化。

```bash
# 1. 程序运行时抓 heap（堆内存快照）
go tool pprof http://localhost:6060/debug/pprof/heap

# 2. 进入 pprof 交互界面后
(pprof) top 10
# 显示：哪些函数分配的内存最多

# 输出像这样：
#   flat  flat%   sum%    cum   cum%
# 125MB 45.2%  45.2%  125MB 45.2%  processPrice       ← 这个函数占了 45%！
#  80MB 28.9%  74.1%   80MB 28.9%  parseOrderBook
#  50MB 18.1%  92.2%   50MB 18.1%  buildResponse
#  15MB  5.4%  97.6%   15MB  5.4%  logTransaction

# 看到 processPrice 占了 45% 的分配量 → 它就是我们要优化的目标

# 3. 看具体哪一行
(pprof) list processPrice
# 显示 processPrice 函数里每一行各分配了多少内存
```

```go
// 在代码里启用 pprof（仅内网！不要暴露到公网）
import (
    _ "net/http/pprof" // 副作用导入，自动注册 pprof 路由
    "net/http"
)

func main() {
    go func() {
        // 只监听内网地址！
        http.ListenAndServe("localhost:6060", nil)
    }()

    // 你的业务代码...
}
```

**一句话**：pprof 就像体检报告——告诉你哪里有问题，但它不能治病。治病靠前两招（对象池 + 预分配）。

---

### 第四招：调 GOGC（最后手段）

#### GOGC 是什么？

```
GOGC = 100（默认）：堆内存增长 100% 就触发一次 GC
GOGC = 200：       堆内存增长 200% 才触发一次 GC（GC 频率降低）
GOGC = 50：        堆内存增长 50% 就触发 GC（GC 频率提高）
```

#### 怎么用？

```bash
# 启动程序时设置
GOGC=200 ./myapp

# 或者在代码里
debug.SetGCPercent(200)
```

#### 什么时候调？

```
GOGC 调高（如 200-400）：
  适用：延迟敏感的服务（交易所撮合引擎、行情推送）
  效果：GC 频率降低 → CPU 更多用于业务 → 延迟更低
  代价：堆内存占用变大（因为 GC 收得不勤快）
  风险：容器内存限制小的话可能 OOM

GOGC 调低（如 25-50）：
  适用：内存很紧张的容器环境
  效果：GC 更勤奋，内存占用保持低位
  代价：GC 消耗更多 CPU

规则：配合 GOMEMLIMIT 使用，防止 OOM
```

```go
// 最佳实践：GOGC + GOMEMLIMIT 配合
// 容器内存 512MB，留给程序 400MB
import "runtime/debug"

func init() {
    debug.SetGCPercent(200)               // 降低 GC 频率
    debug.SetMemoryLimit(400 * 1024 * 1024) // 但设置软上限，内存快满时强制 GC，防止 OOM
}
```

---

## 五、四招的优先级（重要！面试必说）

```
你的程序 GC 压力大
      ↓
第 1 步：pprof 诊断 → 找到分配大户（谁在疯狂创建对象？）
      ↓
第 2 步：sync.Pool → 高频对象池化，用完还回去
      ↓
第 3 步：make 带 cap → 一次性分配到位，避免扩容
      ↓
第 4 步（前 3 步做完还不行）：调 GOGC
      ↓
      ⚠️ 顺序不能反！
      前三招是"止血"（从根本上减少分配）
      调 GOGC 是"吃止痛药"（掩盖问题，不解决根本）
```

---

## 六、完整例子：优化前 vs 优化后

### 优化前（问题代码）

```go
package main

import (
    "fmt"
    "runtime"
    "time"
)

type PriceTick struct {
    Symbol string
    Price  float64
    Time   time.Time
}

// ❌ 问题代码：每秒 10000 次请求，每次 new 一个新对象
func processPricesBad(prices []float64) {
    for _, p := range prices {
        tick := &PriceTick{ // ← 每循环一次 new 一个
            Symbol: "BTC/USDT",
            Price:  p,
            Time:   time.Now(),
        }
        // 模拟发送
        _ = tick
    }
}

func main() {
    var m runtime.MemStats

    for round := 0; round < 5; round++ {
        prices := make([]float64, 10000) // 模拟 10000 次价格更新
        processPricesBad(prices)

        runtime.ReadMemStats(&m)
        fmt.Printf("第 %d 轮 | GC 次数: %d | 堆分配: %.2f MB\n",
            round+1, m.NumGC, float64(m.HeapAlloc)/1024/1024)
    }
}
```

### 优化后

```go
package main

import (
    "fmt"
    "runtime"
    "time"
)

type PriceTick struct {
    Symbol string
    Price  float64
    Time   time.Time
}

// ✅ 创建一个对象池
var tickPool = sync.Pool{
    New: func() any {
        return &PriceTick{}
    },
}

// ✅ 优化代码：从池子里借，用完还回去
func processPricesGood(prices []float64) {
    for _, p := range prices {
        tick := tickPool.Get().(*PriceTick) // 借（不是 new）
        tick.Symbol = "BTC/USDT"
        tick.Price = p
        tick.Time = time.Now()

        _ = tick

        tickPool.Put(tick) // 还
    }
}

// ✅ 优化2：切片预分配
func buildPriceList(count int) []float64 {
    // 提前给 cap，避免动态扩容
    prices := make([]float64, 0, count)
    for i := 0; i < count; i++ {
        prices = append(prices, float64(i))
    }
    return prices
}

func main() {
    var m runtime.MemStats

    for round := 0; round < 5; round++ {
        prices := buildPriceList(10000) // 预分配
        processPricesGood(prices)       // 对象池

        runtime.ReadMemStats(&m)
        fmt.Printf("第 %d 轮 | GC 次数: %d | 堆分配: %.2f MB\n",
            round+1, m.NumGC, float64(m.HeapAlloc)/1024/1024)
    }
}
```

**预期效果**：
- 分配量：从每秒 ~10000 次降到几乎为 0（池子里的对象循环用）
- GC 次数：显著减少
- 内存占用：更稳定，不会频繁波动

---

## 七、面试速记卡

```
GC 是什么？
  → 自动回收不再使用的内存，程序员不用手动 free
  → 三色标记法：标记活的 → 清的死的

GC 的问题？
  → 回收本身耗 CPU，太频繁会影响业务性能
  → 交易所行情剧烈波动时，GC 卡 200ms 用户就多花几百刀

优化四招（按顺序）：
  1. pprof 找谁在疯狂分配
  2. sync.Pool 对象池复用（共享充电宝）
  3. make 带 cap 预分配（一次分够）
  4. 调 GOGC + GOMEMLIMIT（最后手段）

原则：
  先减分配（前三招是止血）
  再调 GOGC（吃药）
  顺序不能反
```

---

## 八、扩展阅读

- **Go 官方博客**：[A Guide to the Go Garbage Collector](https://go.dev/doc/gc-guide)
- **调试工具**：`go tool pprof`、`go tool trace`
- **检测库**：`go.uber.org/goleak`（检测 Goroutine 泄露）
