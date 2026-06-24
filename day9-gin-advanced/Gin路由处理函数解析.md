# Gin 路由处理函数解析

## 完整代码

```go
r.GET("/users", func(c *gin.Context) {
    users := []string{"Alice", "Bob", "Charlie"}
    Success(c, gin.H{"users": users})
})
```

## 逐层拆解

### 第一层：注册路由

```go
r.GET("/users", 处理函数)
```

| 部分 | 含义 |
|------|------|
| `r` | Gin 引擎实例 |
| `.GET` | HTTP 方法，只响应 GET 请求 |
| `"/users"` | URL 路径 |
| `func(c *gin.Context) { ... }` | 匿名函数 = 处理函数 |

### 第二层：准备数据

```go
users := []string{"Alice", "Bob", "Charlie"}
```

Go 的切片（数组）字面量，Vue 开发者一看就懂：

```go
// Go
users := []string{"Alice", "Bob", "Charlie"}
```

```js
// JS 等价
const users = ["Alice", "Bob", "Charlie"]
```

### 第三层：调用 Success 返回

```go
Success(c, gin.H{"users": users})
```

两个参数拆开看：

| 参数 | 值 | 作用 |
|------|-----|------|
| `c` | Gin 的上下文对象 | 给 Success 提供发送 HTTP 响应的能力 |
| `gin.H{"users": users}` | 实际数据 | 传给 Response 的 `Data` 字段 |

### 什么是 `gin.H`？

本质就是 `map[string]any`（键值对），等价 JS 对象：

```go
gin.H{"users": users}
// 完全等于
map[string]interface{}{"users": users}
```

```js
// JS 等价
{ users: users }
```

---

## 最终返回 JSON

```json
{
    "code": 200,
    "message": "成功",
    "data": {
        "users": ["Alice", "Bob", "Charlie"]
    }
}
```

## 数据流

```
请求 GET /users
        │
        ▼
┌──────────────────────────┐
│  users := []string{      │  ← ① 准备数据
│    "Alice","Bob","Charlie"│
│  }                        │
└──────────────────────────┘
        │
        ▼
┌──────────────────────────┐
│  Success(c, gin.H{       │  ← ② 调用封装函数
│    "users": users         │
│  })                       │
└──────────────────────────┘
        │
        ▼
┌──────────────────────────┐
│  Response{               │  ← ③ Success 内部构造统一格式
│    Code:    200,          │
│    Message: "成功",        │
│    Data:    gin.H{...}    │
│  }                        │
└──────────────────────────┘
        │
        ▼
     JSON 响应
```

## 有 Success 之前 vs 之后

```go
// 没有 Success 时（每个接口要重复写 Response{...}）
r.GET("/users", func(c *gin.Context) {
    users := []string{"Alice", "Bob", "Charlie"}
    c.JSON(200, Response{
        Code:    200,
        Message: "成功",
        Data:    gin.H{"users": users},
    })
})

// 有了 Success 后（一行搞定）
r.GET("/users", func(c *gin.Context) {
    users := []string{"Alice", "Bob", "Charlie"}
    Success(c, gin.H{"users": users})
})
```

## 同模式的其他接口

```go
// 查询单个
r.GET("/user/:id", func(c *gin.Context) {
    id := c.Param("id")
    user := map[string]string{"id": id, "name": "张三"}
    Success(c, gin.H{"user": user})
})

// 创建
r.POST("/user", func(c *gin.Context) {
    Success(c, gin.H{"message": "创建成功"})
})

// Data 不是必须的（有 omitempty，为空时不输出）
r.DELETE("/user/:id", func(c *gin.Context) {
    Success(c, nil)  // 返回 {"code":200,"message":"成功"}
})
```

---

> **一句话**：`Success(c, 数据)` = 把数据包进 `{code:200, message:"成功", data:...}` 统一格式返回。`gin.H` 就是 `map[string]any`，用于快速构造 JSON 对象，等价 JS 的 `{}`。
