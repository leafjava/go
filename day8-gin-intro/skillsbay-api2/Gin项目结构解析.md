# Gin 项目结构解析（skillsbay-api）

## 目录结构总览

```
skillsbay-api/
├── main.go              ← 入口文件：启动服务器、注册路由
├── models/
│   └── wallet.go        ← 数据模型：定义数据结构（相当于 Vue 的 TS 接口）
├── handlers/
│   └── wallet.go        ← 处理函数：接口的业务逻辑（相当于 Vue 的 methods）
└── go.mod               ← 模块定义：包名 + 依赖列表（相当于 package.json）
```

## 逐文件解析

---

### 1. `go.mod` — 模块定义文件

**作用**：相当于前端的 `package.json`，定义项目名称和所有依赖。

```go
module skillsbay-api   // ← 项目模块名，import 时的根路径

go 1.25                // ← Go 版本

require (              // ← 依赖列表（相当于 dependencies）
    github.com/gin-gonic/gin v1.12.0
)
```

| 功能 | Go | Vue 前端对比 |
|------|-----|-------------|
| 项目名称 | `module skillsbay-api` | `"name": "skillsbay-api"` 在 `package.json` |
| 依赖管理 | `require (...)` | `"dependencies": { ... }` |
| 安装依赖 | `go mod tidy` | `npm install` |
| 锁版本 | `go.sum` | `package-lock.json` |

---

### 2. `main.go` — 应用入口

**作用**：相当于 `main.js` / `App.vue`，负责**启动服务 + 注册路由**。

```go
package main

import "github.com/gin-gonic/gin"

func main() {
	r := gin.Default()          // 创建 Gin 实例（相当于 createApp()）

	r.GET("/users", handler)    // 注册路由（相当于 router.get('/users', ...)）

	r.Run(":8080")              // 启动服务器（相当于 app.listen(8080)）
}
```

| 功能 | Go | Vue 前端对比 |
|------|-----|-------------|
| 程序入口 | `func main()` | `main.ts` 入口文件 |
| 创建实例 | `gin.Default()` | `createApp()` |
| 注册路由 | `r.GET("/users", ...)` | `router.addRoute(...)` |
| 启动监听 | `r.Run(":8080")` | `app.listen(8080)` |

---

### 3. `models/wallet.go` — 数据模型层

**作用**：定义数据结构，相当于 Vue 中的 **TypeScript 接口 + 表单校验规则**。

```go
package models

type Wallet struct {
	Address string  `json:"address"`
	Balance float64 `json:"balance"`
}
```

| 功能 | Go | Vue 前端对比 |
|------|-----|-------------|
| 定义数据形状 | `type Wallet struct { ... }` | `interface Wallet { address: string }` |
| 字段校验 | `` `binding:"required"` `` | `rules: [{ required: true }]` |
| JSON 映射 | `` `json:"address"` `` | `{ address: "0x..." }` |
| 文件作用 | 纯数据结构，不含业务逻辑 | 相当于 `types/` 目录下的 TS 文件 |

> **职责单一原则**：models 层只负责"长什么样"，不负责"做什么"。

---

### 4. `handlers/wallet.go` — 处理函数层

**作用**：编写具体的业务处理逻辑，相当于 Vue 组件中的 **methods / 事件处理函数**。

```go
package handlers

import "github.com/gin-gonic/gin"

func GetWallet(c *gin.Context) {
	address := c.Param("address")

	c.JSON(200, gin.H{
		"address": address,
		"balance": 10.5,
	})
}
```

| 功能 | Go | Vue 前端对比 |
|------|-----|-------------|
| 处理请求 | `func GetWallet(c *gin.Context)` | `const submit = () => { ... }` |
| 获取参数 | `c.Param("id")` | `route.params.id` |
| 返回响应 | `c.JSON(200, data)` | `return Response.json(data)` |
| 文件作用 | 每个函数对应一个接口 | 每个函数对应一个事件处理 |

> **职责单一原则**：handlers 层只负责"怎么处理请求和返回响应"，不定义数据结构。

---

## 四层对比总结

```
前端 Vue                           后端 Go (Gin)
───────────                        ─────────────
package.json       ←→    go.mod           （依赖管理）
main.ts / App.vue  ←→    main.go          （入口 + 路由注册）
types/index.ts     ←→    models/          （数据结构定义）
api/*.ts           ←→    handlers/        （业务处理逻辑）
```

## 数据流向

```
用户请求
  │
  ▼
main.go（路由匹配）  →  哪个 URL 对应哪个处理函数？
  │
  ▼
handlers/（处理逻辑） →  收到请求后干什么？
  │
  ▼
models/（数据结构）   →  请求/响应的数据长什么样？
  │
  ▼
JSON 响应  →  返回给前端
```

## 为什么这样分层？

| 原因 | 说明 |
|------|------|
| **职责分离** | 每个文件只做一件事，好找好改 |
| **代码复用** | models 可以被多个 handler 共用 |
| **易于维护** | 新人一看目录就知道代码在哪 |
| **方便测试** | 每一层可以独立编写测试 |
