# Gin 中间件（Middleware）详解

## 1. 完整代码

```go
// 使用中间件
auth := r.Group("/api")
auth.Use(AuthMiddleware())

// 定义中间件
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")

        if token == "" {
            c.JSON(http.StatusUnauthorized, gin.H{
                "error": "缺少认证令牌",
            })
        }
    }
}
```

## 2. Vue 类比（先建立直觉）

| Go Gin 概念 | Vue 类比 |
|---|---|
| 中间件 `Use(AuthMiddleware())` | Vue Router 的 **导航守卫** `router.beforeEach()` |
| `c.GetHeader("Authorization")` | 从请求头取 token，相当于 `localStorage.getItem('token')` |
| `c.Next()` | `next()` — 放行，继续下一步 |
| `c.Abort()` | `return false` — 阻止，到此为止 |
| 给分组加中间件 | 给某组路由统一加守卫 |

```js
// Vue Router 等价写法
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (!token) {
    return next('/login')   // 没 token，踢去登录
  }
  next()                    // 有 token，放行
})
```

---

## 3. 中间件的本质

**中间件就是一个"安检口"**，请求到达处理函数之前，先过中间件这一关。

```
请求 → [中间件1] → [中间件2] → [处理函数] → 响应
         ↓              ↓
    检查 token      记日志
```

---

## 4. 逐行拆解

### 4.1 `.Use()` — 装上中间件

```go
auth := r.Group("/api")     // auth 是 /api 开头的路由组
auth.Use(AuthMiddleware())  // 给这个组装上认证中间件
```

`.Use()` 的意思是：**之后这个组里的所有接口，请求来了都要先过这个中间件**。

```
/api
├── /api/users      ← 需要认证
├── /api/wallet     ← 需要认证
└── /api/transfer   ← 需要认证
```

> 类比：Vue 中给某组路由统一加 `meta: { requiresAuth: true }`。

### 4.2 函数嵌套 — 为什么要 `return func()`？

```go
func AuthMiddleware() gin.HandlerFunc {   // 外层：工厂函数
    return func(c *gin.Context) {         // 内层：真正的中间件逻辑
        // 在这里写校验代码
    }
}
```

这是一个**闭包**模式，分两层理解：

| 层 | 作用 | 执行时机 |
|----|------|----------|
| 外层 `AuthMiddleware()` | 创建中间件（可传配置参数） | 程序启动时执行一次 |
| 内层 `func(c *gin.Context)` | 真正的请求处理逻辑 | **每个请求来**都执行一遍 |

**如果不需要传参，其实可以简写：**

```go
// 复杂写法（可传参）
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) { ... }
}

// 简单写法（不需要传参时）
func AuthMiddleware(c *gin.Context) {
    // 直接用
}
```

但写成 `return func()` 是为了以后方便加参数，比如：

```go
func AuthMiddleware(requiredRole string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 可以用 requiredRole 这个参数
    }
}

// 使用时传参
auth.Use(AuthMiddleware("admin"))
```

### 4.3 `c.GetHeader("Authorization")` — 取请求头

```go
token := c.GetHeader("Authorization")
```

等同于前端 axios 请求中设置的：

```js
// 前端发送时
axios.get('/api/users', {
  headers: { Authorization: 'Bearer xxx-token-xxx' }
})

// 后端这里取出来就是 'Bearer xxx-token-xxx'
```

`c` 是 Gin 的上下文对象，提供了这些取参数的常用方法：

| 方法 | 取什么 | Vue 类比 |
|------|--------|----------|
| `c.Param("id")` | URL 路径参数 `/user/:id` | `route.params.id` |
| `c.Query("name")` | URL 查询参数 `?name=张三` | `route.query.name` |
| `c.GetHeader("Authorization")` | 请求头 | `config.headers.Authorization` |
| `c.ShouldBindJSON(&req)` | 请求体 JSON | `request.body` |

---

## 5. 关键问题：这段代码少了什么？

当前代码**有 bug**：校验不通过时返回了错误 JSON，但**没有阻止请求继续往下走**！

### ❌ 错误写法（会报错）

```go
if token == "" {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证令牌"})
    // 少了 c.Abort()！请求还会继续走到处理函数！
}
```

### ✅ 正确写法

```go
if token == "" {
    c.JSON(http.StatusUnauthorized, gin.H{"error": "缺少认证令牌"})
    c.Abort()  // ← 阻止！不要继续往下走了
    return     // ← 函数也立即返回
}
```

`c.Abort()` 的作用就是告诉 Gin：**这个请求不合格，别再调用后面的中间件和处理函数了**。

---

## 6. 完整流程图

```
请求进来了
    │
    ▼
┌─────────────────────┐
│  AuthMiddleware     │
│                     │
│  token 是否存在？    │
│   ├── 有 → c.Next() │────→ 下一个中间件 → 处理函数 → 响应
│   └── 无 → c.Abort()│────→ 直接返回 401
└─────────────────────┘
```

---

## 7. `c.Next()` vs `c.Abort()` 对比

| 方法 | 作用 | 类比 Vue |
|------|------|----------|
| `c.Next()` | 放行，继续执行后续中间件 | `next()` |
| `c.Abort()` | 拦截，跳过所有后续步骤 | `next('/login')` 阻止导航 |

```go
func MyMiddleware(c *gin.Context) {
    // 请求进来时做的事（前处理）
    log.Println("请求来了")

    c.Next()  // ← 执行后续中间件 + 处理函数

    // 响应返回时做的事（后处理）
    log.Println("响应已返回")
}
```

---

## 8. 中间件执行顺序记忆口诀

```
请求 → 前置 → c.Next() → 处理函数 → 后置 → 响应
         ↓                  ↓          ↓
     检查 token         业务逻辑    记录日志
```

`c.Next()` 就像一条分界线：
- **上面**的代码：请求进来时先做
- **下面**的代码：处理完返回时再做

---

## 9. 一句话总结

> `.Use()` = 给路由装上安检装置。每个请求到达处理函数之前，必须经过中间件的检查。`c.Next()` 放行，`c.Abort()` 拦截。类比 Vue Router 的 `beforeEach` 导航守卫。
