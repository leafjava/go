# Gin 框架 — 结构体与参数校验详解

## 1. 完整的结构体定义

```go
type CreateUserRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Age   int    `json:"age" binding:"required,gte=18"`
}
```

## 2. 拆解结构（三部分）

每个字段由 **字段名 + 类型 + 标签** 组成：

```
Name  string  `json:"name"  binding:"required"`
 ↓       ↓         ↓              ↓
字段名  类型    JSON字段名      校验规则
```

| 部分 | 说明 |
|------|------|
| `Name` | Go 内部使用的字段名（首字母大写 = 公开） |
| `string` | 数据类型：字符串 |
| `` `json:"name"` `` | 前后端交互时 JSON 的 key 名 |
| `` `binding:"required"` `` | Gin 校验规则 |

## 3. 标签的作用（struct tag）

反引号 `` ` `` 包裹的内容称为 **结构体标签**，本质是给字段挂的元数据注解。

### 3.1 `json:"xxx"` — JSON 映射

前后端通信使用 JSON 格式，这个标签控制字段的序列化/反序列化：

```go
Name string `json:"name"`
//                ↑
//   前端发 { "name": "张三" } → Go 结构体 Name 字段 = "张三"
//   返回时 Name = "张三" → 前端收到 { "name": "张三" }
```

> **类比 Vue**：相当于定义了接口的 key 名
> ```ts
> interface User { name: string }  // ← "name" 是前后端约定的字段名
> ```

### 3.2 `binding:"xxx"` — 参数校验

Gin 框架会自动读取 `binding` 标签，在绑定请求数据时执行校验，不通过则返回 400 错误。

> **类比 Vue**：相当于 Element Plus 表单的 rules 属性
> ```js
> rules: { name: [{ required: true }] }
> ```

## 4. 字段逐一解析

### 4.1 Name 字段

```go
Name string `json:"name" binding:"required"`
```

- **含义**：姓名字段，必填
- **前端发送**：`{ "name": "张三" }`
- **Vue 等价写法**：
  ```js
  name: [{ required: true, message: '请填写姓名' }]
  ```

### 4.2 Email 字段

```go
Email string `json:"email" binding:"required,email"`
```

- **含义**：邮箱字段，必填 + 必须符合邮箱格式
- **前端发送**：`{ "email": "abc@example.com" }`
- **Vue 等价写法**：
  ```js
  email: [
    { required: true, message: '请填写邮箱' },
    { type: 'email', message: '邮箱格式不正确' }
  ]
  ```

### 4.3 Age 字段

```go
Age int `json:"age" binding:"required,gte=18"`
```

- **含义**：年龄字段，必填 + 必须 ≥ 18
- **前端发送**：`{ "age": 20 }`
- **Vue 等价写法**：
  ```js
  age: [
    { required: true, message: '请填写年龄' },
    { type: 'number', min: 18, message: '必须年满18岁' }
  ]
  ```

## 5. 校验规则速览

| 规则 | 含义 | Vue 等价 |
|------|------|----------|
| `required` | 必填，不能为空 | `required: true` |
| `email` | 必须是邮箱格式 | `type: 'email'` |
| `min=n` | 字符串最少 n 个字符 | `minlength: n` |
| `max=n` | 字符串最多 n 个字符 | `maxlength: n` |
| `len=n` | 字符串恰好 n 个字符 | 自定义校验 |
| `gte=n` | `≥ n`（greater than or equal） | `min: n` |
| `gt=n` | `> n`（greater than） | 无直接对应 |
| `lte=n` | `≤ n`（less than or equal） | `max: n` |
| `lt=n` | `< n`（less than） | 无直接对应 |
| `eq=n` | 等于某值 | 无直接对应 |
| `ne=n` | 不等于某值 | 无直接对应 |
| `oneof=a b c` | 必须是 a、b、c 之一 | 枚举校验 |
| `url` | 必须是 URL 格式 | `type: 'url'` |
| `ip` | 必须是 IP 地址 | 自定义 |

## 6. 组合使用

多个规则用逗号分隔：

```go
Name string `json:"name" binding:"required,min=2,max=20"`
//                         必填    最少2个字符  最多20个字符
```

## 7. 在路由中使用结构体

```go
r.POST("/users", func(c *gin.Context) {
	var req CreateUserRequest

	// ShouldBindJSON 会读取 JSON 请求体，并按 binding 标签自动校验
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 校验通过，req 里就是合法数据
	c.JSON(http.StatusCreated, gin.H{
		"name":  req.Name,
		"email": req.Email,
		"age":   req.Age,
	})
})
```

## 8. 一句话总结

> Go 结构体标签 = **TypeScript 接口定义 + Element Plus 表单校验规则** 二合一，Gin 框架自动完成 JSON 解析和参数校验，不需要手动写 `if` 判断。
