# 第2课：变量、常量、基础类型

> 学习时间：2-3小时 | 难度：⭐⭐

## 📋 本课目标

- 掌握 Go 的变量声明方式
- 理解 Go 的基础数据类型
- 学会使用常量和 iota
- 掌握类型转换

## 1. 变量声明（4种方式）

### 方式1：var 完整声明

```go
package main

import "fmt"

func main() {
    // 声明并初始化
    var name string = "林燊"
    var age int = 23
    var isStudent bool = true
    
    fmt.Println(name, age, isStudent)
    
    // 声明后赋值
    var city string
    city = "广州"
    fmt.Println(city)
}
```

### 方式2：类型推断

```go
func main() {
    var name = "林燊"        // 自动推断为 string
    var age = 23            // 自动推断为 int
    var score = 95.5        // 自动推断为 float64
    
    fmt.Printf("%T, %T, %T\n", name, age, score)
    // 输出：string, int, float64
}
```

### 方式3：短声明（最常用）⭐

```go
func main() {
    // := 只能在函数内使用
    name := "林燊"
    age := 23
    city := "广州"
    
    // 一次声明多个
    x, y, z := 1, 2, 3
    
    fmt.Println(name, age, city)
    fmt.Println(x, y, z)
}
```

### 方式4：批量声明

```go
var (
    name   string = "林燊"
    age    int    = 23
    city   string = "广州"
    salary float64
)

func main() {
    salary = 15000.0
    fmt.Println(name, age, city, salary)
}
```

## 2. 基础数据类型

### 整数类型

```go
package main

import "fmt"

func main() {
    // 有符号整数
    var a int8 = 127          // -128 ~ 127
    var b int16 = 32767       // -32768 ~ 32767
    var c int32 = 2147483647  // -2^31 ~ 2^31-1
    var d int64 = 9223372036854775807
    
    // 无符号整数
    var e uint8 = 255         // 0 ~ 255
    var f uint16 = 65535      // 0 ~ 65535
    var g uint32 = 4294967295
    var h uint64 = 18446744073709551615
    
    // int 和 uint（根据系统自动选择 32 或 64 位）
    var i int = 100
    var j uint = 200
    
    fmt.Println(a, b, c, d, e, f, g, h, i, j)
}
```

### 浮点数类型

```go
func main() {
    var price float32 = 99.99
    var balance float64 = 10000.123456789
    
    fmt.Printf("价格: %.2f\n", price)
    fmt.Printf("余额: %.2f\n", balance)
    
    // 科学计数法
    var gasPrice float64 = 1.5e-9  // 1.5 * 10^-9
    fmt.Println("Gas Price:", gasPrice)
}
```

### 字符串类型

```go
func main() {
    // 双引号：解释字符串
    var name string = "林燊\n广东财经大学"
    fmt.Println(name)
    
    // 反引号：原始字符串（不转义）
    var address string = `0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb`
    fmt.Println(address)
    
    // 字符串拼接
    greeting := "Hello, " + "Go!"
    fmt.Println(greeting)
    
    // 字符串长度
    fmt.Println("长度:", len(name))  // 字节数，不是字符数
}
```

### 布尔类型

```go
func main() {
    var isWeb3Developer bool = true
    var hasExperience bool = false
    
    // 逻辑运算
    canApply := isWeb3Developer && hasExperience
    fmt.Println("可以申请:", canApply)
    
    // 比较运算
    age := 23
    isAdult := age >= 18
    fmt.Println("是否成年:", isAdult)
}
```

## 3. 常量

### 基本常量

```go
package main

import "fmt"

// 常量：编译时确定，不可修改
const PI = 3.14159
const AppName = "SkillsBay"

func main() {
    const MaxRetry = 3
    
    fmt.Println(PI, AppName, MaxRetry)
    
    // PI = 3.14  // 错误：不能修改常量
}
```

### iota 枚举器（重要）⭐

```go
package main

import "fmt"

// iota：自动递增的常量生成器
const (
    Sunday = iota     // 0
    Monday            // 1
    Tuesday           // 2
    Wednesday         // 3
    Thursday          // 4
    Friday            // 5
    Saturday          // 6
)

// 区块链网络枚举
const (
    Ethereum = iota + 1  // 1
    BSC                  // 2
    Polygon              // 3
    TON                  // 4
)

// 权限枚举（位运算）
const (
    ReadPermission   = 1 << iota  // 1 (二进制: 001)
    WritePermission               // 2 (二进制: 010)
    DeletePermission              // 4 (二进制: 100)
)

func main() {
    fmt.Println("今天是:", Wednesday)
    fmt.Println("网络:", TON)
    
    // 权限组合
    userPermission := ReadPermission | WritePermission
    fmt.Printf("用户权限: %b\n", userPermission)  // 输出: 11
}
```

## 4. 类型转换

### 基本类型转换

```go
package main

import (
    "fmt"
    "strconv"
)

func main() {
    // 数值类型转换（必须显式转换）
    var a int = 100
    var b float64 = float64(a)
    var c int32 = int32(a)
    
    fmt.Println(a, b, c)
    
    // 字符串转数字
    str := "123"
    num, err := strconv.Atoi(str)
    if err != nil {
        fmt.Println("转换失败:", err)
    } else {
        fmt.Println("数字:", num)
    }
    
    // 数字转字符串
    age := 23
    ageStr := strconv.Itoa(age)
    fmt.Println("年龄字符串:", ageStr)
    
    // 字符串转浮点数
    priceStr := "99.99"
    price, _ := strconv.ParseFloat(priceStr, 64)
    fmt.Println("价格:", price)
    
    // 布尔值转换
    boolStr := "true"
    boolVal, _ := strconv.ParseBool(boolStr)
    fmt.Println("布尔值:", boolVal)
}
```

## 5. 零值（默认值）

```go
package main

import "fmt"

func main() {
    var i int        // 0
    var f float64    // 0.0
    var b bool       // false
    var s string     // ""
    
    fmt.Printf("int: %d\n", i)
    fmt.Printf("float64: %f\n", f)
    fmt.Printf("bool: %t\n", b)
    fmt.Printf("string: '%s'\n", s)
}
```

## 6. Go vs Java 类型对比

| Java | Go | 说明 |
|------|-----|------|
| `int` | `int` | 整数（Go 根据系统自动选择 32/64 位）|
| `long` | `int64` | 64位整数 |
| `float` | `float32` | 32位浮点数 |
| `double` | `float64` | 64位浮点数 |
| `boolean` | `bool` | 布尔值 |
| `String` | `string` | 字符串 |
| `final` | `const` | 常量 |
| `enum` | `const + iota` | 枚举 |

## 📝 作业

### 作业1：Web3 钱包信息

创建 `homework/day2/wallet.go`：

```go
package main

import "fmt"

func main() {
    // TODO: 声明以下变量
    // 1. 钱包地址（string）
    // 2. ETH 余额（float64）
    // 3. USDT 余额（float64）
    // 4. 是否已验证（bool）
    // 5. 交易次数（int）
    
    // TODO: 输出钱包信息
    // 示例输出：
    // ========== 钱包信息 ==========
    // 地址: 0x742d35Cc...
    // ETH 余额: 1.5 ETH
    // USDT 余额: 1000.00 USDT
    // 已验证: true
    // 交易次数: 42
    // ==============================
}
```

### 作业2：Gas 费计算器

创建 `homework/day2/gas_calculator.go`：

```go
package main

import "fmt"

func main() {
    // Gas 参数
    gasLimit := 21000              // Gas 限制
    gasPrice := 50.0               // Gwei
    ethPrice := 2000.0             // USD
    
    // TODO: 计算以下内容
    // 1. Gas 费用（ETH）= gasLimit * gasPrice / 1e9
    // 2. Gas 费用（USD）= Gas费用(ETH) * ethPrice
    // 3. 输出结果，保留4位小数
    
    // 示例输出：
    // Gas Limit: 21000
    // Gas Price: 50.0 Gwei
    // Gas 费用: 0.0011 ETH
    // Gas 费用: $2.1000
}
```

### 作业3：区块链网络枚举

创建 `homework/day2/blockchain_enum.go`：

```go
package main

import "fmt"

// TODO: 使用 iota 定义区块链网络枚举
const (
    // Ethereum = ?
    // BSC = ?
    // Polygon = ?
    // Arbitrum = ?
    // Optimism = ?
    // TON = ?
)

// TODO: 使用 iota 定义交易状态枚举
const (
    // Pending = ?
    // Confirmed = ?
    // Failed = ?
)

func getNetworkName(network int) string {
    // TODO: 根据网络ID返回网络名称
    return ""
}

func getStatusName(status int) string {
    // TODO: 根据状态ID返回状态名称
    return ""
}

func main() {
    // 测试你的枚举
    fmt.Println("网络:", getNetworkName(TON))
    fmt.Println("状态:", getStatusName(Confirmed))
}
```

### 作业4：类型转换练习

创建 `homework/day2/type_conversion.go`：

```go
package main

import (
    "fmt"
    "strconv"
)

func main() {
    // 场景1：用户输入的金额（字符串）转换为数字
    amountStr := "1000.50"
    // TODO: 转换为 float64 并计算手续费（0.1%）
    
    // 场景2：区块高度（int64）转换为字符串
    blockHeight := int64(18500000)
    // TODO: 转换为字符串并输出
    
    // 场景3：Gas Price（float64）转换为 Wei（int64）
    gasPriceGwei := 50.5
    // TODO: 转换为 Wei（1 Gwei = 1e9 Wei）
    
    // 场景4：十六进制地址转换
    hexStr := "0x1a2b3c"
    // TODO: 使用 strconv.ParseInt 转换为十进制
}
```

## 🎯 检查点

完成本课后，你应该能够：
- ✅ 熟练使用4种变量声明方式
- ✅ 理解 Go 的基础数据类型
- ✅ 使用 const 和 iota 定义常量和枚举
- ✅ 进行类型转换
- ✅ 理解零值概念

## 💡 重点提示

1. **短声明 `:=` 最常用**，但只能在函数内使用
2. **Go 不支持隐式类型转换**，必须显式转换
3. **iota 在 Web3 开发中常用于定义网络、状态枚举**
4. **字符串是不可变的**，修改需要转换为 []byte

## 📚 扩展阅读

- [Go by Example - Variables](https://gobyexample.com/variables)
- [Go by Example - Constants](https://gobyexample.com/constants)
- [Go 语言圣经 - 第2章](https://gopl-zh.github.io/ch2/ch2.html)

## ⏭️ 下一课

[第3课：函数、错误处理](./day3-functions.md)
