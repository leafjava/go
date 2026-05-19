# 第1课：环境搭建 + Hello World

> 学习时间：1-2小时 | 难度：⭐

## 📋 本课目标

- 安装 Go 开发环境
- 配置 VS Code / GoLand
- 理解 Go 项目结构
- 写出第一个 Go 程序

## 1. 安装 Go

### Windows 安装

```bash
# 1. 下载 Go 安装包（1.21+）
# https://go.dev/dl/

# 2. 双击安装，默认路径：C:\Go

# 3. 验证安装
go version
# 输出：go version go1.21.x windows/amd64
```

### 配置环境变量

```bash
# 查看 Go 环境
go env

# 设置 GOPROXY（国内加速）
go env -w GOPROXY=https://goproxy.cn,direct

# 设置 Go Module（推荐）
go env -w GO111MODULE=on
```

## 2. 安装 IDE

### 方案1：VS Code（推荐新手）

```bash
# 1. 安装 VS Code
# https://code.visualstudio.com/

# 2. 安装 Go 扩展
# 搜索 "Go" by Go Team at Google

# 3. 安装 Go 工具
# Ctrl+Shift+P → Go: Install/Update Tools → 全选安装
```

### 方案2：GoLand（推荐专业开发）

```bash
# JetBrains GoLand
# https://www.jetbrains.com/go/
# 学生可免费申请
```

## 3. 第一个 Go 程序

### 创建项目

```bash
# 1. 创建项目目录
cd D:\webProject\go
mkdir hello
cd hello

# 2. 初始化 Go Module
go mod init hello

# 3. 创建 main.go
```

### hello/main.go

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Go!")
    fmt.Println("我是林燊，准备征服 Go 语言！")
}
```

### 运行程序

```bash
# 方式1：直接运行
go run main.go

# 方式2：编译后运行
go build
./hello.exe  # Windows
./hello      # Linux/Mac

# 输出：
# Hello, Go!
# 我是林燊，准备征服 Go 语言！
```

## 4. Go 项目结构

```
hello/
├── go.mod          # 依赖管理文件（类似 package.json）
├── go.sum          # 依赖版本锁定（类似 package-lock.json）
├── main.go         # 主程序入口
└── README.md       # 项目说明
```

### go.mod 文件解析

```go
module hello        // 模块名称

go 1.21            // Go 版本

// 依赖包会自动添加到这里
// require (
//     github.com/gin-gonic/gin v1.9.1
// )
```

## 5. Go 基础语法预览

### 变量声明

```go
package main

import "fmt"

func main() {
    // 方式1：完整声明
    var name string = "林燊"
    
    // 方式2：类型推断
    var age = 23
    
    // 方式3：短声明（最常用）
    city := "广州"
    
    fmt.Println(name, age, city)
}
```

### 函数定义

```go
package main

import "fmt"

// 函数：参数类型在后面，返回值类型在最后
func add(a int, b int) int {
    return a + b
}

// 多个返回值（Go 特色）
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, fmt.Errorf("除数不能为0")
    }
    return a / b, nil
}

func main() {
    sum := add(10, 20)
    fmt.Println("10 + 20 =", sum)
    
    result, err := divide(10, 2)
    if err != nil {
        fmt.Println("错误:", err)
    } else {
        fmt.Println("10 / 2 =", result)
    }
}
```

## 6. Go vs Java 对比

| 特性 | Java | Go |
|------|------|-----|
| 包声明 | `package com.example;` | `package main` |
| 导入 | `import java.util.*;` | `import "fmt"` |
| 主函数 | `public static void main(String[] args)` | `func main()` |
| 变量声明 | `String name = "林燊";` | `name := "林燊"` |
| 错误处理 | `try-catch` | `if err != nil` |
| 并发 | `Thread / ExecutorService` | `goroutine / channel` |

## 📝 作业

### 作业1：个人信息卡片

创建 `homework/day1/profile.go`，输出你的个人信息：

```go
package main

import "fmt"

func main() {
    // TODO: 输出你的姓名、年龄、学校、技能栈
    // 要求：使用不同的变量声明方式
    
    // 示例输出：
    // ========== 个人信息 ==========
    // 姓名: 林燊
    // 年龄: 23
    // 学校: 广东财经大学
    // 技能: Java, Vue, Solidity, Go
    // ==============================
}
```

### 作业2：简单计算器

创建 `homework/day1/calculator.go`，实现加减乘除：

```go
package main

import "fmt"

// TODO: 实现以下函数
func add(a, b float64) float64 {
    // 返回 a + b
}

func subtract(a, b float64) float64 {
    // 返回 a - b
}

func multiply(a, b float64) float64 {
    // 返回 a * b
}

func divide(a, b float64) (float64, error) {
    // 如果 b == 0，返回错误
    // 否则返回 a / b
}

func main() {
    // 测试你的函数
    fmt.Println("10 + 5 =", add(10, 5))
    fmt.Println("10 - 5 =", subtract(10, 5))
    fmt.Println("10 * 5 =", multiply(10, 5))
    
    result, err := divide(10, 0)
    if err != nil {
        fmt.Println("错误:", err)
    } else {
        fmt.Println("10 / 0 =", result)
    }
}
```

### 作业3：Web3 地址验证器

创建 `homework/day1/address_validator.go`：

```go
package main

import (
    "fmt"
    "strings"
)

// 验证以太坊地址格式（0x开头，42个字符）
func isValidEthAddress(address string) bool {
    // TODO: 实现验证逻辑
    // 1. 检查是否以 "0x" 开头
    // 2. 检查长度是否为 42
    // 3. 检查是否只包含十六进制字符
    return false
}

func main() {
    addresses := []string{
        "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
        "0xInvalidAddress",
        "742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
    }
    
    for _, addr := range addresses {
        if isValidEthAddress(addr) {
            fmt.Printf("✓ %s 是有效地址\n", addr)
        } else {
            fmt.Printf("✗ %s 是无效地址\n", addr)
        }
    }
}
```

## 🎯 检查点

完成本课后，你应该能够：
- ✅ 成功安装 Go 并配置开发环境
- ✅ 创建并运行 Go 程序
- ✅ 理解 Go 的基本语法结构
- ✅ 声明变量和定义函数
- ✅ 处理简单的错误

## 📚 扩展阅读

- [Go 官方教程 - Tour of Go](https://go.dev/tour/welcome/1)
- [Go by Example - Hello World](https://gobyexample.com/hello-world)
- [Go 语言圣经 - 第1章](https://gopl-zh.github.io/ch1/ch1-01.html)

## ⏭️ 下一课

[第2课：变量、常量、基础类型](./day2-basics.md)

---

**💪 加油！你已经迈出了 Go 语言学习的第一步！**
