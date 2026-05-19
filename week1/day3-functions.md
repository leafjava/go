# 第3课：函数、错误处理

> 学习时间：2-3小时 | 难度：⭐⭐

## 📋 本课目标

- 掌握 Go 函数的定义和调用
- 理解多返回值机制
- 掌握 Go 的错误处理模式
- 学会使用匿名函数和闭包

## 1. 函数基础

### 基本语法

```go
package main

import "fmt"

// 函数定义：func 函数名(参数列表) 返回类型 { 函数体 }
func greet(name string) string {
    return "Hello, " + name
}

// 多个参数（相同类型可以合并）
func add(a, b int) int {
    return a + b
}

// 多个不同类型参数
func createUser(name string, age int, isVIP bool) {
    fmt.Printf("用户: %s, 年龄: %d, VIP: %t\n", name, age, isVIP)
}

func main() {
    msg := greet("林燊")
    fmt.Println(msg)
    
    sum := add(10, 20)
    fmt.Println("Sum:", sum)
    
    createUser("林燊", 23, true)
}
```

## 2. 多返回值（Go 特色）⭐

### 基本用法

```go
package main

import (
    "fmt"
    "errors"
)

// 返回多个值
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, errors.New("除数不能为0")
    }
    return a / b, nil
}

// 命名返回值
func getUserInfo(id int) (name string, age int, err error) {
    if id <= 0 {
        err = errors.New("无效的用户ID")
        return  // 自动返回 name, age, err
    }
    
    name = "林燊"
    age = 23
    return  // 等价于 return name, age, nil
}

func main() {
    // 接收多个返回值
    result, err := divide(10, 2)
    if err != nil {
        fmt.Println("错误:", err)
    } else {
        fmt.Println("结果:", result)
    }
    
    // 忽略某个返回值（使用 _）
    name, _, _ := getUserInfo(1)
    fmt.Println("用户名:", name)
}
```

### Web3 实战示例

```go
package main

import (
    "errors"
    "fmt"
    "strings"
)

// 验证以太坊地址并返回格式化地址
func validateEthAddress(address string) (string, error) {
    // 检查长度
    if len(address) != 42 {
        return "", errors.New("地址长度必须为42个字符")
    }
    
    // 检查前缀
    if !strings.HasPrefix(address, "0x") {
        return "", errors.New("地址必须以0x开头")
    }
    
    // 转换为小写（标准格式）
    normalized := strings.ToLower(address)
    return normalized, nil
}

// 计算 Gas 费用
func calculateGasFee(gasLimit int64, gasPriceGwei float64) (ethCost float64, usdCost float64, err error) {
    if gasLimit <= 0 {
        err = errors.New("Gas Limit 必须大于0")
        return
    }
    
    if gasPriceGwei <= 0 {
        err = errors.New("Gas Price 必须大于0")
        return
    }
    
    // 计算 ETH 费用
    ethCost = float64(gasLimit) * gasPriceGwei / 1e9
    
    // 假设 ETH 价格 $2000
    ethPrice := 2000.0
    usdCost = ethCost * ethPrice
    
    return
}

func main() {
    // 测试地址验证
    addr, err := validateEthAddress("0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb")
    if err != nil {
        fmt.Println("地址验证失败:", err)
    } else {
        fmt.Println("有效地址:", addr)
    }
    
    // 测试 Gas 计算
    ethCost, usdCost, err := calculateGasFee(21000, 50)
    if err != nil {
        fmt.Println("计算失败:", err)
    } else {
        fmt.Printf("Gas 费用: %.6f ETH ($%.2f)\n", ethCost, usdCost)
    }
}
```

## 3. 错误处理（重要）⭐⭐⭐

### Go 的错误处理哲学

```go
package main

import (
    "errors"
    "fmt"
)

// Go 没有 try-catch，使用返回值传递错误

// 方式1：返回 error
func connectDatabase(host string) error {
    if host == "" {
        return errors.New("主机地址不能为空")
    }
    
    // 模拟连接
    fmt.Println("连接数据库:", host)
    return nil
}

// 方式2：自定义错误类型
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("字段 %s 验证失败: %s", e.Field, e.Message)
}

func validateUser(name string, age int) error {
    if name == "" {
        return &ValidationError{
            Field:   "name",
            Message: "用户名不能为空",
        }
    }
    
    if age < 18 {
        return &ValidationError{
            Field:   "age",
            Message: "年龄必须大于18岁",
        }
    }
    
    return nil
}

// 方式3：包装错误（Go 1.13+）
func processTransaction(txHash string) error {
    if txHash == "" {
        return errors.New("交易哈希为空")
    }
    
    // 模拟错误
    err := errors.New("网络超时")
    if err != nil {
        return fmt.Errorf("处理交易失败 %s: %w", txHash, err)
    }
    
    return nil
}

func main() {
    // 错误处理模式
    if err := connectDatabase("localhost"); err != nil {
        fmt.Println("错误:", err)
        return
    }
    
    // 类型断言检查错误类型
    if err := validateUser("", 20); err != nil {
        if ve, ok := err.(*ValidationError); ok {
            fmt.Printf("验证错误 - 字段: %s, 消息: %s\n", ve.Field, ve.Message)
        } else {
            fmt.Println("未知错误:", err)
        }
    }
    
    // 错误包装
    if err := processTransaction("0xabc123"); err != nil {
        fmt.Println("错误:", err)
    }
}
```

### Go vs Java 错误处理对比

```go
// Java 方式
/*
try {
    result = divide(10, 0);
} catch (ArithmeticException e) {
    System.out.println("错误: " + e.getMessage());
}
*/

// Go 方式
result, err := divide(10, 0)
if err != nil {
    fmt.Println("错误:", err)
}
```

## 4. 可变参数

```go
package main

import "fmt"

// 可变参数（类似 Java 的 String... args）
func sum(numbers ...int) int {
    total := 0
    for _, num := range numbers {
        total += num
    }
    return total
}

// 格式化日志
func log(level string, messages ...string) {
    fmt.Printf("[%s] ", level)
    for _, msg := range messages {
        fmt.Print(msg, " ")
    }
    fmt.Println()
}

func main() {
    // 传递任意数量的参数
    fmt.Println(sum(1, 2, 3))           // 6
    fmt.Println(sum(1, 2, 3, 4, 5))     // 15
    
    // 传递切片
    nums := []int{10, 20, 30}
    fmt.Println(sum(nums...))           // 60
    
    log("INFO", "服务启动", "端口:8080")
    log("ERROR", "连接失败", "重试中...")
}
```

## 5. 匿名函数和闭包

### 匿名函数

```go
package main

import "fmt"

func main() {
    // 匿名函数
    add := func(a, b int) int {
        return a + b
    }
    
    result := add(10, 20)
    fmt.Println("结果:", result)
    
    // 立即执行函数
    func(name string) {
        fmt.Println("Hello,", name)
    }("林燊")
}
```

### 闭包（重要）⭐

```go
package main

import "fmt"

// 闭包：函数 + 引用环境
func counter() func() int {
    count := 0
    return func() int {
        count++
        return count
    }
}

// Web3 实战：创建交易计数器
func createTxCounter(initialCount int) func(txHash string) int {
    count := initialCount
    
    return func(txHash string) int {
        count++
        fmt.Printf("处理交易 #%d: %s\n", count, txHash)
        return count
    }
}

func main() {
    // 计数器示例
    c1 := counter()
    fmt.Println(c1())  // 1
    fmt.Println(c1())  // 2
    fmt.Println(c1())  // 3
    
    c2 := counter()
    fmt.Println(c2())  // 1（独立的计数器）
    
    // 交易计数器
    txCounter := createTxCounter(0)
    txCounter("0xabc123")
    txCounter("0xdef456")
    txCounter("0xghi789")
}
```

## 6. defer 延迟执行

```go
package main

import "fmt"

// defer：函数返回前执行（类似 Java 的 finally）
func processFile() {
    fmt.Println("1. 打开文件")
    defer fmt.Println("4. 关闭文件")  // 最后执行
    
    fmt.Println("2. 读取文件")
    fmt.Println("3. 处理数据")
}

// 多个 defer 按 LIFO（后进先出）顺序执行
func multipleDefer() {
    defer fmt.Println("第一个 defer")
    defer fmt.Println("第二个 defer")
    defer fmt.Println("第三个 defer")
    
    fmt.Println("函数体")
}

// Web3 实战：确保释放资源
func connectBlockchain() error {
    fmt.Println("连接区块链节点...")
    defer fmt.Println("断开区块链连接")
    
    // 模拟操作
    fmt.Println("查询余额...")
    fmt.Println("发送交易...")
    
    return nil
}

func main() {
    processFile()
    fmt.Println("---")
    multipleDefer()
    fmt.Println("---")
    connectBlockchain()
}

// 输出：
// 1. 打开文件
// 2. 读取文件
// 3. 处理数据
// 4. 关闭文件
// ---
// 函数体
// 第三个 defer
// 第二个 defer
// 第一个 defer
// ---
// 连接区块链节点...
// 查询余额...
// 发送交易...
// 断开区块链连接
```

## 📝 作业

### 作业1：以太坊交易验证器

创建 `homework/day3/tx_validator.go`：

```go
package main

import (
    "errors"
    "fmt"
    "strings"
)

// TODO: 实现交易验证函数
// 返回：是否有效、错误信息
func validateTransaction(from, to string, amount float64) (bool, error) {
    // 1. 验证发送地址（42字符，0x开头）
    // 2. 验证接收地址
    // 3. 验证金额（必须大于0）
    // 如果全部通过，返回 true, nil
    // 否则返回 false, error
    return false, nil
}

// TODO: 实现 Gas 估算函数
// 返回：预估 Gas、ETH 费用、USD 费用、错误
func estimateGas(txType string, dataSize int) (gasLimit int64, ethCost float64, usdCost float64, err error) {
    // txType: "transfer" (21000 gas), "contract" (50000 + dataSize*68 gas)
    // gasPrice: 50 Gwei
    // ethPrice: $2000
    return 0, 0, 0, nil
}

func main() {
    // 测试交易验证
    valid, err := validateTransaction(
        "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
        "0x8ba1f109551bD432803012645Ac136ddd64DBA72",
        1.5,
    )
    
    if err != nil {
        fmt.Println("验证失败:", err)
    } else if valid {
        fmt.Println("✓ 交易有效")
    }
    
    // 测试 Gas 估算
    gas, eth, usd, err := estimateGas("transfer", 0)
    if err != nil {
        fmt.Println("估算失败:", err)
    } else {
        fmt.Printf("Gas: %d, ETH: %.6f, USD: $%.2f\n", gas, eth, usd)
    }
}
```

### 作业2：区块链查询器

创建 `homework/day3/blockchain_query.go`：

```go
package main

import (
    "errors"
    "fmt"
)

// 模拟区块数据
type Block struct {
    Number       int64
    Hash         string
    Transactions int
}

// TODO: 实现获取最新区块函数
func getLatestBlock() (*Block, error) {
    // 模拟返回最新区块
    // 10% 概率返回错误（模拟网络问题）
    return nil, nil
}

// TODO: 实现获取指定区块函数
func getBlockByNumber(number int64) (*Block, error) {
    // 如果 number <= 0，返回错误
    // 否则返回模拟区块数据
    return nil, nil
}

// TODO: 实现批量查询函数（使用可变参数）
func getBlocks(numbers ...int64) ([]*Block, error) {
    // 查询多个区块
    // 如果任何一个查询失败，返回错误
    return nil, nil
}

func main() {
    // 测试你的函数
}
```

### 作业3：钱包管理器（闭包练习）

创建 `homework/day3/wallet_manager.go`：

```go
package main

import "fmt"

// TODO: 创建钱包管理器（使用闭包）
func createWallet(initialBalance float64) (
    deposit func(float64) float64,
    withdraw func(float64) (float64, error),
    getBalance func() float64,
) {
    // 实现存款、取款、查询余额三个闭包函数
    // balance 变量被三个函数共享
    
    return nil, nil, nil
}

func main() {
    // 创建钱包
    deposit, withdraw, getBalance := createWallet(100.0)
    
    // 测试存款
    fmt.Println("存款后余额:", deposit(50))
    
    // 测试取款
    newBalance, err := withdraw(30)
    if err != nil {
        fmt.Println("取款失败:", err)
    } else {
        fmt.Println("取款后余额:", newBalance)
    }
    
    // 测试查询
    fmt.Println("当前余额:", getBalance())
}
```

### 作业4：defer 实战

创建 `homework/day3/defer_practice.go`：

```go
package main

import (
    "fmt"
    "time"
)

// TODO: 实现函数执行时间统计（使用 defer）
func measureTime(funcName string) func() {
    // 返回一个函数，用于计算执行时间
    // 提示：记录开始时间，返回的函数中计算时间差
    return nil
}

// TODO: 实现资源清理函数
func processTransaction(txHash string) error {
    // 使用 defer 确保资源被正确清理
    // 1. 打印 "开始处理交易"
    // 2. defer 打印 "清理资源"
    // 3. defer 打印 "关闭连接"
    // 4. 模拟处理逻辑
    return nil
}

func main() {
    // 测试时间统计
    defer measureTime("main")()
    
    // 模拟耗时操作
    time.Sleep(100 * time.Millisecond)
    
    // 测试资源清理
    processTransaction("0xabc123")
}
```

## 🎯 检查点

完成本课后，你应该能够：
- ✅ 定义和调用函数
- ✅ 使用多返回值处理结果和错误
- ✅ 正确处理错误（Go 的核心模式）
- ✅ 使用可变参数
- ✅ 理解和使用闭包
- ✅ 使用 defer 管理资源

## 💡 重点提示

1. **Go 的错误处理是显式的**，不要忽略 error 返回值
2. **多返回值是 Go 的特色**，充分利用它
3. **defer 常用于资源清理**（关闭文件、数据库连接等）
4. **闭包在 Web3 开发中常用于状态管理**

## 📚 扩展阅读

- [Go by Example - Functions](https://gobyexample.com/functions)
- [Go by Example - Errors](https://gobyexample.com/errors)
- [Go by Example - Defer](https://gobyexample.com/defer)
- [Go by Example - Closures](https://gobyexample.com/closures)

## ⏭️ 下一课

[第4课：结构体、方法、接口](./day4-structs.md)
