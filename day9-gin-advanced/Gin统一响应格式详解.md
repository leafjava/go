# Gin 统一响应格式详解

## 1. 什么是统一响应？

前端收到的每个接口返回，格式都一模一样：

```json
{
    "code": 200,
    "message": "成功",
    "data": { ... }
}
```

不管查用户、创建订单、转账，都是这个结构。前端解析时不需要每接口写一套逻辑。

## 2. 结构体定义

```go
type Response struct {
    Code    int         `json:"code"`
    Message string      `json:"message"`
    Data    interface{} `json:"data,omitempty"`
}
```

## 3. 逐词拆解

```go
type  Response  struct { ... }
 ↓       ↓        ↓
 1       2        3
```

| 单词 | 中文 | 作用 | Vue/TS 类比 |
|------|------|------|-------------|
| `type` | 声明类型 | Go 关键字，表示"我要创建一个新类型" | `type`（TS 中一模一样） |
| `Response` | 类型名 | 自己起的名字，叫啥都行 | `interface Response { }` 的 Response |
| `struct` | 结构体 | 多个字段组合在一起的数据容器 | `interface` / `class` |

```go
// Go
type Response struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}
```

```ts
// TypeScript 等价写法
interface Response {
    code: number
    message: string
}
```

## 4. 三个字段解析

### 4.1 `Code int` — 业务状态码

```go
Code int `json:"code"`
```

| 说明 | 内容 |
|------|------|
| **类型** | `int`（整数） |
| **作用** | 告诉前端"结果是什么" |
| **和 HTTP 状态码的关系** | 不一样！HTTP 的 200/404 是传输层，`code` 是业务层 |

常用约定：

| code 值 | 含义 |
|--------|------|
| `200` | 成功 |
| `400` | 参数错误 |
| `401` | 未登录 |
| `403` | 无权限 |
| `404` | 数据不存在 |
| `500` | 服务器错误 |

### 4.2 `Message string` — 提示信息

```go
Message string `json:"message"`
```

| 说明 | 内容 |
|------|------|
| **类型** | `string`（字符串） |
| **作用** | 给人看的错误/成功提示 |

```json
// 成功时
{ "code": 200, "message": "操作成功" }

// 失败时
{ "code": 400, "message": "用户名不能为空" }
```

> 前端可以直接把 `message` 弹给用户看，不用自己再写一遍文案。

### 4.3 `Data interface{}` — 实际数据

```go
Data interface{} `json:"data,omitempty"`
```

| 说明 | 内容 |
|------|------|
| **类型** | `interface{}`（空接口 = 任意类型 = TS 的 `any`） |
| **作用** | 放真正的返回数据 |
| **`omitempty`** | 如果为空（nil），JSON 里不出现这个字段 |

```json
// 有数据时
{ "code": 200, "message": "成功", "data": { "username": "张三", "age": 25 } }

// Data 为空时，omitempty 让它直接消失
{ "code": 200, "message": "操作成功" }
```

## 5. `omitempty` 详解

| Data 的值 | 不加 omitempty | 加了 omitempty |
|-----------|----------------|-----------------|
| `"hello"` | `"data": "hello"` | `"data": "hello"` |
| `nil` | `"data": null` 丑 😫 | **字段消失** 干净 ✅ |
| 有值的对象 | `"data": {...}` | `"data": {...}` |

前后端对接时，`"data": null` 会让前端代码出 bug（比如 `.data.items.length` 报空指针）。`omitempty` 直接不输出，前端判断 `if (res.data)` 就行。

## 6. `interface{}` 为什么能放任何类型？

Go 的 `interface{}`（空接口）相当于 TS 的 `any`：

```go
// 都能放
resp.Data = "字符串"
resp.Data = 123
resp.Data = map[string]int{"a": 1}
resp.Data = User{Name: "张三", Age: 25}
```

```ts
// TypeScript 等价
interface Response {
    data: any  // ← 什么都能放
}
```

## 7. 完整使用示例

### 7.1 定义成功/失败辅助函数

```go
// 成功响应
func Success(c *gin.Context, data interface{}) {
    c.JSON(http.StatusOK, Response{
        Code:    200,
        Message: "成功",
        Data:    data,
    })
}

// 错误响应
func Error(c *gin.Context, code int, message string) {
    c.JSON(http.StatusOK, Response{   // 注意 HTTP 状态码仍然是 200
        Code:    code,                // 业务错误码放在 body 里
        Message: message,
    })
}
```

### 7.2 在 handler 中使用

```go
// 原来要写一堆
c.JSON(http.StatusOK, gin.H{
    "code":    200,
    "message": "成功",
    "data":    result,
})

// 现在一行搞定
Success(c, result)
```

### 7.3 前端统一处理

```js
// axios 拦截器中统一处理
axios.interceptors.response.use(res => {
    const { code, message, data } = res.data

    if (code === 200) {
        return data              // 成功，直接拿数据
    } else if (code === 401) {
        router.push('/login')    // 未登录，跳登录页
        return Promise.reject(message)
    } else {
        alert(message)           // 其他错误，弹窗提示
        return Promise.reject(message)
    }
})
```

## 8. 一句话总结

> `Response` 是一个包装壳，所有接口都用它返回。`Code` 表示状态（成功/失败类型），`Message` 是给人读的提示，`Data` 携带实际数据。`omitempty` 让空数据不出现，`interface{}` 让 Data 字段什么类型都能装。
