# Gin 路由分组语法详解

## 1. 完整代码

```go
api := r.Group("/api/v1")
{
    wallets := api.Group("/wallets")
    {
        wallets.GET("/:address", handlers.Getwallet)
        wallets.GET("/:address/balance", handlers.GetBalance)
        wallets.POST("/:address/transfer", handlers.Transfer)
    }
}
```

## 2. 先看最终效果（这才是一眼该懂的）

上面那段代码，最终注册了这 3 个接口：

| 方法 | 完整 URL | 处理函数 |
|------|----------|----------|
| GET | `/api/v1/wallets/:address` | `Getwallet` |
| GET | `/api/v1/wallets/:address/balance` | `GetBalance` |
| POST | `/api/v1/wallets/:address/transfer` | `Transfer` |

**路由就是自动拼接出来的**：`api` 的前缀 + `wallets` 的前缀 + 具体路径。

---

## 3. 逐层拆解

### 第一层：创建版本分组

```go
api := r.Group("/api/v1")   // api 分组，前缀是 /api/v1
```

| 概念 | 说明 | Vue 类比 |
|------|------|----------|
| `r` | Gin 引擎实例 | Vue Router 实例 |
| `.Group("/api/v1")` | 创建一个路由组，前缀 `/api/v1` | `router.addRoute({ path: '/api/v1', children: [...] })` |
| `api` | 返回的子路由对象 | 一个带 basePath 的子路由 |

### 第二层：创建功能分组

```go
wallets := api.Group("/wallets")  // 继承 api 的 /api/v1，再加 /wallets
```

此时 `wallets` 的完整前缀是：`/api/v1` + `/wallets` = **`/api/v1/wallets`**

### 第三层：注册具体接口

```go
wallets.GET("/:address", handlers.Getwallet)
// 完整路径 = /api/v1/wallets + /:address = /api/v1/wallets/:address

wallets.GET("/:address/balance", handlers.GetBalance)
// 完整路径 = /api/v1/wallets + /:address/balance = /api/v1/wallets/:address/balance

wallets.POST("/:address/transfer", handlers.Transfer)
// 完整路径 = /api/v1/wallets + /:address/transfer = /api/v1/wallets/:address/transfer
```

---

## 4. 大括号 `{}` 是干什么的？（容易困惑的点）

```go
api := r.Group("/api/v1")
{                          // ← 这个大括号
    wallets := api.Group("/wallets")
    {                      // ← 这个大括号
        ...
    }
}
```

**答案：纯粹是为了好看，没有实际功能！**

Go 语言中 `{}` 单独使用，只表示一个**代码块（block）**，不影响任何逻辑。删掉大括号，代码功能完全一样：

```go
// 这样写功能完全一样，只是看起来乱
api := r.Group("/api/v1")
wallets := api.Group("/wallets")
wallets.GET("/:address", handlers.Getwallet)
wallets.GET("/:address/balance", handlers.GetBalance)
wallets.POST("/:address/transfer", handlers.Transfer)
```

### 那为什么加 `{}`？

视觉上表示**层级包含关系**，读代码时一眼看出：

```
api {                 ← api 下面"包着"这些东西
    wallets {         ← wallets 下面"包着"这些东西
        路由1
        路由2
        路由3
    }
}
```

> 就像 Vue 模板缩进让你看清 `<div>` 的父子关系一样，这里的 `{}` + 缩进让你看清路由的层级。

---

## 5. 对比 Vue Router（一对比就懂）

### Go Gin 写法

```go
api := r.Group("/api/v1")
{
    wallets := api.Group("/wallets")
    {
        wallets.GET("/:address", handlers.Getwallet)
    }
}
```

### Vue Router 等价写法

```js
const routes = [
  {
    path: '/api/v1',
    children: [
      {
        path: 'wallets',
        children: [
          // :address → 动态参数，相当于 /api/v1/wallets/0x123
          { path: ':address', name: 'Getwallet', method: 'GET' },
        ]
      }
    ]
  }
]
```

### 关键区别

| | Go Gin | Vue Router |
|------|--------|------------|
| 分组方式 | `.Group("/xxx")` | `children: [...]` |
| 动态参数 | `:address` | `:address`（一样） |
| 作用域 | `{}` 只是视觉分组 | `{ }` 是真正的嵌套对象 |
| 层级关系 | 前缀自动拼接 | 路径自动拼接 |

---

## 6. 路由拼接头脑练习

把前缀代进去算一遍：

```
r.Group("/api/v1")           → 前缀: /api/v1
  .Group("/wallets")         → 前缀: /api/v1/wallets
    .GET("/:address")        → 完整: /api/v1/wallets/:address
    .GET("/:address/balance")  → 完整: /api/v1/wallets/:address/balance
    .POST("/:address/transfer") → 完整: /api/v1/wallets/:address/transfer
```

---

## 7. 为什么这样设计？

| 好处 | 说明 |
|------|------|
| **避免重复** | 不用每个路由都写一遍 `/api/v1/wallets` |
| **统一前缀** | 改版本号只需改一处：`/api/v1` → `/api/v2` |
| **清晰层级** | URL 层级 = 代码层级，对得上 |
| **中间件复用** | 可以给整个分组加中间件（认证、限流等） |

```go
// 比如统一给 /api/v1 下面的接口加认证
api := r.Group("/api/v1", authMiddleware())
// 这个分组下所有接口都要登录后才能访问
```

---

## 8. 一句话总结

> `r.Group()` 每个层级加一段路径前缀，最终 URL = 所有前缀拼起来 + 最后的路径。`{}` 只是为了让代码看起来有层次感，没有实际逻辑作用。
