# Go 方法接收者（Method Receiver）详解

## 完整代码

```go
func (e *AppError) Error() string {
    return e.Message
}
```

## 逐词拆解

```
func  (e *AppError)  Error()  string  { return e.Message }
 1        2     3       4        5            6
```

| 位置 | 代码 | 含义 | 说明 |
|------|------|------|------|
| ① | `func` | 关键字 | 定义函数 |
| ② | `(e *AppError)` | **接收者** | 把方法挂到 AppError 类型上，`e` 相当于 `this` |
| ③ | `Error()` | 方法名 | 叫 Error |
| ④ | `string` | 返回类型 | 返回一个字符串 |
| ⑤ | `string` | 返回类型 | 返回字符串 |
| ⑥ | `e.Message` | 函数体 | 返回当前实例的 Message 字段 |

---

## 核心概念：接收者 = Go 的 `this`

```go
// 普通函数 — 不属于任何类型
func SayHello() string {
    return "Hello"
}

// 方法 — 属于 AppError 类型
func (e *AppError) Error() string {
    return e.Message    // e = 当前实例，相当于 JS 的 this
}
```

### 三语言对比

```go
// Go（方法接收者）
type AppError struct {
    Message string
}
func (e *AppError) Error() string {
    return e.Message          // e = this
}
```

```ts
// TypeScript（class）
class AppError {
    message: string
    error(): string {
        return this.message   // this = 当前实例
    }
}
```

```js
// Vue 组件中
methods: {
    error() {
        return this.message   // this = 当前组件实例
    }
}
```

| Go | TS/JS | 含义 |
|------|------|------|
| `(e *AppError)` | `class AppError { ... }` | 方法属于 AppError |
| `e` | `this` | 当前实例对象 |
| `e.Message` | `this.message` | 访问实例字段 |

---

## `*` 是什么？接收者分两种

### 值接收者

```go
func (e AppError) Error() string {   // 不加 *
    e.Message = "改了"                // ❌ 改的是副本，不影响原值
    return e.Message
}
```

### 指针接收者（推荐）

```go
func (e *AppError) Error() string {  // 加 *
    e.Message = "改了"                // ✅ 改的是原值
    return e.Message
}
```

| | 值接收者 `(e T)` | 指针接收者 `(e *T)` |
|------|------|------|
| 能否修改原值 | ❌ 不能，改的是副本 | ✅ 能，直接改原值 |
| 性能 | 大结构体拷贝慢 | 只传地址，快 |
| 使用场景 | 极小的只读结构体 | **默认选这个** |

> 和之前 `ShouldBindJSON(&req)` 需要传 `&` 是一个道理：Go 默认值传递，想改原变量就用指针。

---

## 为什么要实现 `Error()` 方法？

Go 内置了 `error` 接口：

```go
// Go 标准库中的定义
type error interface {
    Error() string
}
```

**只要你的类型实现了 `Error() string` 方法，它就自动实现 `error` 接口，能当错误使用。**

```go
type AppError struct {
    Code    int
    Message string
}

func (e *AppError) Error() string {
    return e.Message
}

// 现在 AppError 是合法的 error 类型了！
func doSomething() error {
    return &AppError{Code: 400, Message: "参数错误"}
    //    ↑ 返回类型声明为 error，实际返回 AppError
}
```

使用方完全不用关心是哪种错误：

```go
err := doSomething()
if err != nil {
    fmt.Println(err.Error())  // "参数错误"
}
```

---

## 完整 AppError 示例

```go
// 定义自定义错误类型
type AppError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

// 实现 error 接口（系统要求）
func (e *AppError) Error() string {
    return e.Message
}

// 在 handler 中使用
r.GET("/user/:id", func(c *gin.Context) {
    user, err := findUser(c.Param("id"))
    if err != nil {
        // err 可能是 AppError，也可能是其他 error
        var appErr *AppError
        if errors.As(err, &appErr) {
            // 是我们自定义的 AppError
            Error(c, appErr.Code, appErr.Message)
        } else {
            // 其他未知错误
            Error(c, 500, "服务器内部错误")
        }
        return
    }
    Success(c, user)
})
```

---

## 总结一句话

> `(e *AppError)` 是 Go 的方法接收者 = JS 的 `this`，表示"这个方法挂在 AppError 类型上"。`Error()` 是实现 Go 标准 error 接口的方法，实现后 AppError 就能当 error 类型使用。`*` 表示指针接收者，能修改原值且性能更好。
