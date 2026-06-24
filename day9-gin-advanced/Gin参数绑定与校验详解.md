# Gin 参数绑定与校验详解（ShouldBindJSON）

## 1. 完整流程

```go
var req CreateUserRequest

if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, gin.H{"error": err.Error()})
    return
}

// 到这里 req 已经填好值了，直接用
c.JSON(201, gin.H{"user": req})
```

一行 `ShouldBindJSON(&req)`，背后做了 **两件事**：

```
c.ShouldBindJSON(&req)
        │
        ├── ① 绑定（Bind）：读请求体 JSON → 填入结构体字段
        │
        └── ② 校验（Should）：按规则逐个验证字段
```

---

## 2. 绑定：JSON 怎么填入结构体？

### 结构体定义

```go
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Email    string `json:"email" binding:"required,email"`
	Age      int    `json:"age" binding:"required,gte=18,lte=100"`
	Password string `json:"password" binding:"required,min=6"`
}
```

### 映射关系

```
前端发来的 JSON                     Go 结构体字段
─────────────────                   ──────────────
{                                    CreateUserRequest {
  "username": "张三丰",  ──→  json:"username"    Username = "张三丰"
  "email": "a@b.com",    ──→  json:"email"       Email    = "a@b.com"
  "age": 25,             ──→  json:"age"          Age      = 25
  "password": "123456"   ──→  json:"password"    Password = "123456"
}                                    }
```

**规则**：`json` 标签决定从哪里取值，JSON key 和标签一一对应。没标 `json` 标签的字段不受影响。

---

## 3. 校验：`binding` 标签怎么工作？

绑定完成后，Gin 按 `binding` 标签的规则逐个字段检查：

```
字段      值         规则                           结果
─────    ─────      ────                           ────
Username  "张三丰"   required, min=3, max=20         ✅ 3 个字符，在 3~20 之间
Email     "a@b.com" required, email                 ✅ 合法邮箱格式
Age       25        required, gte=18, lte=100       ✅ 25 在 18~100 之间
Password  "123456"  required, min=6                 ✅ 6 个字符，≥6
                                                    全部通过 → err == nil
```

任何一个不通过，立刻返回错误，不会继续：

```
Age: 17  →  gte=18 不满足  →  err != nil  →  返回 400
                  ↓
{"error": "Key: 'CreateUserRequest.Age' Error:Field validation for 'Age' failed on the 'gte' tag"}
```

---

## 4. 为什么要加 `&`？

```go
c.ShouldBindJSON(&req)
//               ↑
//         传指针，不是传值
```

### Go 语言基础：值传递 vs 指针传递

| 写法 | 效果 |
|------|------|
| `ShouldBindJSON(req)` | ❌ 传副本 → 函数改了副本 → `req` 还是空的 |
| `ShouldBindJSON(&req)` | ✅ 传地址 → 函数直接修改 `req` 本身 |

```go
// 反例（如果不用 &）
var req CreateUserRequest
c.ShouldBindJSON(req)   // 传的是拷贝
// req 在这里还是 Username="" Email="" Age=0 Password=""
```

Go 中 `&` 是取地址符，得到的值称为**指针（pointer）**。传到函数里，函数就能通过地址找到原始变量，直接修改它。

> **Vue 开发者注意**：JS 对象天然是引用传递，所以不需要这个操作。Go 的结构体默认是**值传递**（复制一份），必须显式传 `&` 才能让函数修改原变量。

---

## 5. `ShouldBindJSON` 内部做了什么？（伪代码）

```go
func ShouldBindJSON(obj any) error {
    // 第 1 步：读取请求体原始字节
    bodyBytes := 读取请求体()
    // bodyBytes = {"username":"张三丰", "email":"a@b.com", ...}

    // 第 2 步：JSON 反序列化 → 按 json 标签填入结构体
    json.Unmarshal(bodyBytes, obj)

    // 第 3 步：按 binding 标签逐字段校验
    校验器.验证(obj)

    // 第 4 步：返回结果
    if 有字段不通过 {
        return 错误信息
    }
    return nil  // nil = 没错误 = 校验通过
}
```

---

## 6. `ShouldBindJSON` vs 其他绑定方法的区别

| 方法 | 数据来源 | 示例 |
|------|----------|------|
| `c.ShouldBindJSON(&obj)` | 请求体 JSON | POST/PUT/PATCH 请求 |
| `c.ShouldBindQuery(&obj)` | URL 查询参数 | `?name=张三&age=18` |
| `c.ShouldBindUri(&obj)` | URL 路径参数 | `/user/:id` |
| `c.ShouldBind(&obj)` | 自动判断来源 | 根据 Content-Type 决定 |

---

## 7. Should vs Must 的区别

| 前缀 | 错误处理 | 适用场景 |
|------|----------|----------|
| `Should...` | 返回 error，你手动处理 | **推荐**，自定义错误格式 |
| `Must...` | 失败自动返回 400，你不处理 | 简单接口，不需要自定义错误 |

```go
// Should：你控制错误格式
if err := c.ShouldBindJSON(&req); err != nil {
    c.JSON(400, gin.H{"error": err.Error()})  // 你自己决定返回什么
}

// Must：框架自动返回 400，格式固定
c.MustBindWith(&req, binding.JSON)  // 出错自动拦截，后面代码不执行
```

> **建议**：始终用 `Should` 系列，能控制错误返回的格式和信息。

---

## 8. 完整流程图

```
客户端 POST /users {"username":"张三丰","email":"a@b.com","age":25,"password":"123456"}
        │
        ▼
┌───────────────────────────────┐
│  var req CreateUserRequest   │  ← 声明空结构体（全是零值）
│  req = {Username:"" Email:"" │
│         Age:0 Password:""}   │
└───────────────────────────────┘
        │
        ▼
┌───────────────────────────────┐
│  c.ShouldBindJSON(&req)      │
│                               │
│  ① 读 JSON 请求体             │
│  ② json.Unmarshal → 填值      │  ← json:"username" 等标签起作用
│  ③ 校验 validator → 逐字段检查  │  ← binding:"required,min=3" 等标签起作用
└───────────────────────────────┘
        │
        ├── err != nil ──→ 返回 400，停止
        │
        └── err == nil ──→ req 已填好，继续往下
                │
                ▼
┌───────────────────────────────┐
│  req.Username = "张三丰"       │
│  req.Email    = "a@b.com"     │
│  req.Age      = 25            │
│  req.Password = "123456"      │
│                               │
│  直接用！                      │
└───────────────────────────────┘
        │
        ▼
  c.JSON(201, gin.H{"user": req})
```

---

## 9. 一句话总结

> `c.ShouldBindJSON(&req)` = **JSON 反序列化 + 表单校验**一步完成。`json` 标签负责填值，`binding` 标签负责验证，`&` 确保改的是原变量。校验不通过时返回 error，你在 `if err != nil` 里自己决定怎么响应。
