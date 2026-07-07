# 第22课：单元测试 + 性能优化

> 学习时间：3-4小时 | 难度：⭐⭐⭐

## 📋 本课目标

- 掌握 Go 表驱动测试（Table-Driven Tests）
- 学会使用 Mock 和依赖注入进行测试
- 掌握基准测试（Benchmark）
- 学会使用 pprof 进行性能分析
- 掌握常见性能优化技巧

## 1. Go 测试基础

### 测试文件命名和函数签名

```
测试规则：
- 测试文件必须以 _test.go 结尾
- 测试函数必须以 Test 开头
- 测试函数签名：func TestXxx(t *testing.T)
- 基准测试函数签名：func BenchmarkXxx(b *testing.B)
- 示例函数签名：func ExampleXxx()
```

### 第一个测试

```go
// mathutil.go
package mathutil

func Add(a, b int) int {
    return a + b
}

func Divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("除数不能为零")
    }
    return a / b, nil
}
```

```go
// mathutil_test.go
package mathutil

import "testing"

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    expected := 5

    if result != expected {
        t.Errorf("Add(2, 3) = %d; want %d", result, expected)
    }
}

func TestDivide(t *testing.T) {
    // 正常情况
    result, err := Divide(10, 2)
    if err != nil {
        t.Fatal("不应该返回错误:", err)
    }
    if result != 5.0 {
        t.Errorf("Divide(10, 2) = %f; want 5.0", result)
    }

    // 除零情况
    _, err = Divide(10, 0)
    if err == nil {
        t.Error("除零应该返回错误")
    }
}
```

## 2. 表驱动测试（Table-Driven Tests）

```go
func TestAdd_TableDriven(t *testing.T) {
    // 定义测试用例表
    tests := []struct {
        name     string // 测试用例名称
        a, b     int    // 输入
        expected int    // 期望输出
    }{
        {"正数相加", 2, 3, 5},
        {"零相加", 0, 0, 0},
        {"负数相加", -2, -3, -5},
        {"正负相加", -2, 5, 3},
        {"大数相加", 1000000, 2000000, 3000000},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d, %d) = %d; want %d", tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

## 3. Mock 和依赖注入

### 接口定义和 Mock

```go
// service.go
package service

import "context"

// BlockchainClient 区块链客户端接口
type BlockchainClient interface {
    GetBalance(ctx context.Context, address string) (string, error)
    GetBlockNumber(ctx context.Context) (uint64, error)
}

// BalanceService 余额服务
type BalanceService struct {
    client BlockchainClient // 依赖接口，不依赖具体实现
}

func NewBalanceService(client BlockchainClient) *BalanceService {
    return &BalanceService{client: client}
}

func (s *BalanceService) GetFormattedBalance(ctx context.Context, address string) (string, error) {
    balance, err := s.client.GetBalance(ctx, address)
    if err != nil {
        return "", fmt.Errorf("查询余额失败: %w", err)
    }
    return fmt.Sprintf("%s ETH", balance), nil
}
```

### Mock 实现

```go
// service_test.go
package service

import (
    "context"
    "errors"
    "testing"
)

// MockBlockchainClient 模拟区块链客户端
type MockBlockchainClient struct {
    GetBalanceFunc     func(ctx context.Context, address string) (string, error)
    GetBlockNumberFunc func(ctx context.Context) (uint64, error)
}

func (m *MockBlockchainClient) GetBalance(ctx context.Context, address string) (string, error) {
    if m.GetBalanceFunc != nil {
        return m.GetBalanceFunc(ctx, address)
    }
    return "0", nil
}

func (m *MockBlockchainClient) GetBlockNumber(ctx context.Context) (uint64, error) {
    if m.GetBlockNumberFunc != nil {
        return m.GetBlockNumberFunc(ctx)
    }
    return 0, nil
}

func TestGetFormattedBalance_Success(t *testing.T) {
    // 创建 Mock
    mockClient := &MockBlockchainClient{
        GetBalanceFunc: func(ctx context.Context, address string) (string, error) {
            return "1.5", nil
        },
    }

    service := NewBalanceService(mockClient)

    result, err := service.GetFormattedBalance(context.Background(), "0x123")
    if err != nil {
        t.Fatal("不应该返回错误:", err)
    }

    expected := "1.5 ETH"
    if result != expected {
        t.Errorf("GetFormattedBalance() = %s; want %s", result, expected)
    }
}

func TestGetFormattedBalance_Error(t *testing.T) {
    mockClient := &MockBlockchainClient{
        GetBalanceFunc: func(ctx context.Context, address string) (string, error) {
            return "", errors.New("RPC 连接失败")
        },
    }

    service := NewBalanceService(mockClient)

    _, err := service.GetFormattedBalance(context.Background(), "0x123")
    if err == nil {
        t.Error("应该返回错误")
    }
}
```

## 4. 子测试（Subtests）和测试辅助函数

```go
// TestWalletValidation 钱包地址验证测试
func TestWalletValidation(t *testing.T) {
    // Setup - 在所有子测试之前执行
    validator := NewAddressValidator()

    // Teardown - 在所有子测试之后执行
    t.Cleanup(func() {
        // 清理资源
    })

    t.Run("ValidEthereumAddress", func(t *testing.T) {
        valid := validator.ValidateEthereum("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
        if !valid {
            t.Error("有效的以太坊地址应该返回 true")
        }
    })

    t.Run("InvalidEthereumAddress", func(t *testing.T) {
        valid := validator.ValidateEthereum("invalid")
        if valid {
            t.Error("无效的以太坊地址应该返回 false")
        }
    })

    t.Run("ValidTONAddress", func(t *testing.T) {
        valid := validator.ValidateTON("EQCkR1cGZxYbG6QrGqKJ8HxSPfI3NZ9sNjCpHlLHyVTGc5Gq")
        if !valid {
            t.Error("有效的 TON 地址应该返回 true")
        }
    })
}

// TestHelper 测试辅助函数
func createTestBalanceService(t *testing.T) *BalanceService {
    t.Helper() // 标记为辅助函数，错误会报告调用方位置

    mockClient := &MockBlockchainClient{
        GetBalanceFunc: func(ctx context.Context, address string) (string, error) {
            return "10.0", nil
        },
    }

    return NewBalanceService(mockClient)
}
```

## 5. 基准测试（Benchmark）

```go
// benchmark_test.go
package service

import (
    "context"
    "fmt"
    "testing"
)

// BenchmarkBalanceFormat 基准测试格式化余额
func BenchmarkBalanceFormat(b *testing.B) {
    ctx := context.Background()
    
    mockClient := &MockBlockchainClient{
        GetBalanceFunc: func(ctx context.Context, address string) (string, error) {
            return "1.5", nil
        },
    }

    service := NewBalanceService(mockClient)

    // b.N 由测试框架自动调整
    for i := 0; i < b.N; i++ {
        service.GetFormattedBalance(ctx, "0x123")
    }
}

// BenchmarkStringConcat 对比字符串拼接方式
func BenchmarkStringConcat(b *testing.B) {
    addresses := []string{"0x123", "0x456", "0x789", "0xabc", "0xdef"}

    b.Run("用+拼接", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            result := ""
            for _, addr := range addresses {
                result += addr + ","
            }
        }
    })

    b.Run("用fmt.Sprintf", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = fmt.Sprintf("%s,%s,%s,%s,%s", addresses[0], addresses[1], addresses[2], addresses[3], addresses[4])
        }
    })

    b.Run("用strings.Builder", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            var builder strings.Builder
            for _, addr := range addresses {
                builder.WriteString(addr)
                builder.WriteByte(',')
            }
            _ = builder.String()
        }
    })

    b.Run("用strings.Join", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            _ = strings.Join(addresses, ",")
        }
    })
}

// 运行基准测试：
// go test -bench=. -benchmem
// go test -bench=BenchmarkStringConcat -benchmem -count=5
```

## 6. 性能分析（pprof）

### CPU 性能分析

```go
package main

import (
    "log"
    "os"
    "runtime/pprof"
)

func main() {
    // 创建 CPU profile 文件
    f, err := os.Create("cpu.prof")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    // 开始 CPU 分析
    pprof.StartCPUProfile(f)
    defer pprof.StopCPUProfile()

    // 你的程序逻辑
    runBusinessLogic()
}

func runBusinessLogic() {
    // 模拟耗时操作
    for i := 0; i < 100000; i++ {
        _ = fmt.Sprintf("block_%d", i)
    }
}
```

### 内存性能分析

```go
package main

import (
    "log"
    "os"
    "runtime"
    "runtime/pprof"
)

func main() {
    // 运行程序
    runBusinessLogic()

    // 创建内存 profile 文件
    f, err := os.Create("mem.prof")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    // 强制 GC
    runtime.GC()

    // 写入内存分析
    if err := pprof.WriteHeapProfile(f); err != nil {
        log.Fatal(err)
    }
}
```

### 使用 pprof 分析

```bash
# 运行测试并生成 profile
go test -cpuprofile=cpu.prof -memprofile=mem.prof -bench=.

# 使用 pprof 分析
go tool pprof cpu.prof

# 常用 pprof 命令：
# top     - 显示占用最高的函数
# list    - 显示具体函数代码
# web     - 生成调用图（需要 graphviz）
# peek    - 查看特定函数

# 启动 Web UI（推荐）
go tool pprof -http=:8080 cpu.prof
```

## 7. 常见性能优化技巧

### 1. 避免不必要的内存分配

```go
// ❌ 不好：每次调用都分配新内存
func BadFormatBalance(balance float64) string {
    return fmt.Sprintf("余额: %.4f ETH", balance)
}

// ✅ 好：使用 strings.Builder 减少分配
func GoodFormatBalance(balance float64) string {
    var builder strings.Builder
    builder.WriteString("余额: ")
    builder.WriteString(strconv.FormatFloat(balance, 'f', 4, 64))
    builder.WriteString(" ETH")
    return builder.String()
}
```

### 2. sync.Pool 复用对象

```go
// 对象池（适用于频繁创建/销毁的对象）
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}

func ProcessWithPool(data []byte) {
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf[:0]) // 重置后放回

    buf = append(buf, data...)
    // 处理 buf...
}
```

### 3. 预分配切片容量

```go
// ❌ 不预分配：多次扩容
func BadCollectAddresses(count int) []string {
    var addresses []string
    for i := 0; i < count; i++ {
        addresses = append(addresses, fmt.Sprintf("0x%x", i))
    }
    return addresses
}

// ✅ 预分配容量：一次分配
func GoodCollectAddresses(count int) []string {
    addresses := make([]string, 0, count) // 预分配容量
    for i := 0; i < count; i++ {
        addresses = append(addresses, fmt.Sprintf("0x%x", i))
    }
    return addresses
}
```

### 4. 使用 strings.Builder 替代 + 拼接

```go
// ❌ 字符串拼接（大量分配）
func BadJoinStrings(strs []string) string {
    result := ""
    for _, s := range strs {
        result += s
    }
    return result
}

// ✅ strings.Builder
func GoodJoinStrings(strs []string) string {
    var builder strings.Builder
    for _, s := range strs {
        builder.WriteString(s)
    }
    return builder.String()
}
```

### 5. 并发安全的缓存

```go
// 使用 sync.Map 或分段锁减少锁竞争
type SafeCache struct {
    mu    sync.RWMutex
    items map[string]interface{}
}

func (c *SafeCache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    val, ok := c.items[key]
    return val, ok
}

func (c *SafeCache) Set(key string, value interface{}) {
    c.mu.Lock()
    defer c.mu.Unlock()

    c.items[key] = value
}
```

## 8. 数据竞争检测

```go
// 使用 race detector 检测竞态条件
// go test -race ./...
// go run -race main.go

// 示例：存在数据竞争的代码
type Counter struct {
    value int
}

// ❌ 有数据竞争
func (c *Counter) BadIncrement() {
    c.value++ // 非原子操作，多 goroutine 同时访问会出错
}

// ✅ 使用互斥锁
type SafeCounter struct {
    mu    sync.Mutex
    value int
}

func (c *SafeCounter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

// ✅ 使用原子操作
type AtomicCounter struct {
    value int64
}

func (c *AtomicCounter) Increment() {
    atomic.AddInt64(&c.value, 1)
}
```

## 9. 测试覆盖率

```bash
# 运行测试并生成覆盖率报告
go test -coverprofile=coverage.out ./...

# 查看覆盖率
go tool cover -func=coverage.out

# 生成 HTML 报告
go tool cover -html=coverage.out -o coverage.html

# 设置覆盖率阈值（在 CI 中使用）
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total | awk '{print $3}'
```

## 10. 性能优化清单

| 优化项 | 说明 | 预期收益 |
|--------|------|---------|
| 预分配切片容量 | `make([]T, 0, cap)` | 减少内存分配和拷贝 |
| strings.Builder | 替代 + 和 fmt.Sprintf | 大量拼接场景 10x+ |
| sync.Pool | 复用临时对象 | 减少 GC 压力 |
| 减少 JSON 编解码 | 使用 json.RawMessage 延迟解析 | 跳过不必要的字段 |
| 批量 RPC 调用 | 合并多次查询为一次 | 减少网络往返 |
| Redis Pipeline | 批量执行 Redis 命令 | 减少网络 IO |
| 数据库索引 | 为查询列添加索引 | 查询速度 10-100x |
| 连接池复用 | 合理配置 HTTP/DB 连接池 | 减少连接开销 |
| 减少锁粒度 | 读写锁 + 分段锁 | 提高并发性能 |

## 📝 作业

### 作业1：为区块链服务编写测试

```go
// TODO: 为 day20-21 的区块链服务编写完整测试
// 1. 为 BalanceService 编写表驱动测试
// 2. Mock Ethereum 和 TON 客户端
// 3. 测试缓存命中/未命中场景
// 4. 测试错误处理路径
// 5. 达到 80% 以上测试覆盖率
```

### 作业2：性能基准测试

```go
// TODO: 为区块链服务编写基准测试
// 1. Benchmark 余额查询（带缓存 vs 不带缓存）
// 2. Benchmark JSON 序列化/反序列化
// 3. Benchmark 字符串拼接方式对比
// 4. 使用 pprof 分析性能瓶颈
```

### 作业3：Gin 中间件测试

```go
// TODO: 为 Gin 中间件编写测试
// 1. 测试 JWT 认证中间件
// 2. 测试限流中间件
// 3. 使用 httptest 模拟 HTTP 请求
func TestJWTAuthMiddleware(t *testing.T) {
    // TODO: 实现
}
```

## 🎯 检查点

- ✅ 掌握表驱动测试
- ✅ 能够编写 Mock 和依赖注入
- ✅ 掌握基准测试
- ✅ 能够使用 pprof 分析性能
- ✅ 了解常见性能优化技巧
- ✅ 掌握数据竞争检测

## ⏭️ 下一课

[第23课：Docker + CI/CD](./day23-docker.md)
