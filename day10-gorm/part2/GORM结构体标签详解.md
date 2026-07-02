# GORM 结构体标签详解

## 1. 回顾：三种标签一起用

```go
type User struct {
    ID       uint   `gorm:"primaryKey" json:"id"`
    Username string `gorm:"uniqueIndex;not null" json:"username" binding:"required,min=3"`
}
```

| 标签 | 谁在用 | 干什么 |
|------|--------|--------|
| `gorm:"..."` | **GORM** | 定义数据库表结构（约束、索引、关联） |
| `json:"..."` | **Gin / encoding/json** | 请求/响应时的 JSON 字段名 |
| `binding:"..."` | **Gin validator** | 请求参数校验 |

> 之前学的 `json` + `binding` 管"请求和响应"，现在新的 `gorm` 管"数据库表怎么建"——三个标签各干各的，互不冲突。

---

## 2. GORM 标签语法

```
gorm:"配置1;配置2:参数;配置3:参数1,参数2"
       ↑                    ↑
   用分号分隔           值用冒号，多参用逗号
```

```go
gorm:"uniqueIndex;not null"
//     ↑            ↑
//  唯一索引      非空约束

gorm:"foreignKey:UserID;references:ID"
//                  ↑
//          值和参数用冒号
```

---

## 3. 常用标签速查

### 3.1 主键相关

| 标签 | 效果 | 使用场景 |
|------|------|----------|
| `primaryKey` | 设为主键 | 默认用 `ID` 字段，其他字段需手动指定 |
| `autoIncrement` | 自增 | 整数主键自动加 |
| `default:值` | 默认值 | 没给值时填什么 |

```go
ID uint `gorm:"primaryKey;autoIncrement" json:"id"`
```

### 3.2 字段约束

| 标签 | 效果 | SQL 等价 |
|------|------|----------|
| `not null` | 不能为 NULL | `NOT NULL` |
| `unique` | 值不能重复 | `UNIQUE` |
| `uniqueIndex` | 不能重复 + 自动建索引 | `CREATE UNIQUE INDEX` |
| `index` | 建普通索引（查得快） | `CREATE INDEX` |
| `default:值` | 默认值 | `DEFAULT '值'` |
| `size:255` | 字段最大长度 | `VARCHAR(255)` |
| `type:text` | 指定数据库类型 | 手动控制 SQL 类型 |
| `comment:说明` | 字段注释 | `COMMENT '说明'` |

```go
Username string `gorm:"uniqueIndex;not null;default:未命名;comment:用户名" json:"username"`
```

### 3.3 关联关系

| 标签 | 作用 |
|------|------|
| `foreignKey:字段名` | 指定外键字段 |
| `references:字段名` | 指定对方表的哪个字段（默认 ID） |
| `constraint:OnDelete:CASCADE` | 级联删除 |
| `many2many:中间表名` | 多对多关联 |

---

## 4. `foreignKey:UserID` 详解

### 4.1 一个完整例子

```go
// 用户表（一）
type User struct {
    ID       uint    `gorm:"primaryKey" json:"id"`
    Username string  `json:"username"`
    Orders   []Order `gorm:"foreignKey:UserID" json:"orders,omitempty"`
}

// 订单表（多）
type Order struct {
    ID     uint    `json:"id"`
    Amount float64 `json:"amount"`
    UserID uint    `json:"user_id"` // ← 外键字段
}
```

### 4.2 数据库中的样子

```
users 表                      orders 表
┌────┬──────────┐            ┌────┬────────┬─────────┐
│ ID │ Username │            │ ID │ Amount │ UserID  │
├────┼──────────┤            ├────┼────────┼─────────┤
│ 1  │ 张三     │←───关联───│ 1  │ 100    │ 1       │
│ 2  │ 李四     │            │ 2  │ 200    │ 1       │
└────┴──────────┘            │ 3  │ 300    │ 2       │
                             └────┴────────┴─────────┘
                              ↑ UserID 列指向 users.ID
```

### 4.3 `foreignKey` 到底说了什么？

```
User struct 里写了：
    Orders []Order `gorm:"foreignKey:UserID"`

翻译成人话：
    "Order 表中的 UserID 列，指向 User 表的 ID"
```

### 4.4 完整版 vs 简写版

```go
// 简写（默认 references 是对方表的 ID）
Orders []Order `gorm:"foreignKey:UserID"`

// 完整版（显式写清楚谁指向谁）
Orders []Order `gorm:"foreignKey:UserID;references:ID"`
//                          Order.UserID  →  User.ID
```

### 4.5 Vue/JS 类比

```ts
// TypeScript 中这样理解
interface User {
    id: number
    username: string
    orders: Order[]   // ← 一对多关系
}

// Prisma / TypeORM 写法（类比）
class User {
    @OneToMany(() => Order, order => order.user)
    orders: Order[]
}

class Order {
    @Column()
    userId: number    // ← 外键
}
```

### 4.6 查询时的效果

```go
// 查用户时自动带出订单
var user User
db.Preload("Orders").First(&user, 1)
// 结果：
// {
//     "id": 1,
//     "username": "张三",
//     "orders": [
//         {"id": 1, "amount": 100},
//         {"id": 2, "amount": 200}
//     ]
// }
```

---

## 5. 完整实例对比

### 5.1 只有 JSON 标签（Gin 之前学的）

```go
type User struct {
    ID       uint   `json:"id"`
    Username string `json:"username"`
}
```

只控制前后端数据格式，不管数据库。

### 5.2 加上 GORM 标签（今天新学的）

```go
type User struct {
    ID       uint   `gorm:"primaryKey;autoIncrement" json:"id"`
    Username string `gorm:"uniqueIndex;not null;size:50" json:"username" binding:"required,min=3"`
}
```

现在有了三个标签：
- `gorm` → 建表：ID 主键自增，username 唯一、非空、最大 50 字符
- `json` → 请求响应：JSON key 名为 `"username"`
- `binding` → 参数校验：必填，至少 3 个字符

### 5.3 加上关联（外键）

```go
type User struct {
    ID     uint    `gorm:"primaryKey" json:"id"`
    Name   string  `json:"name"`
    Orders []Order `gorm:"foreignKey:UserID" json:"orders,omitempty"`
}
```

GORM 建 orders 表时就知道 `UserID` 是外键，查询时 `Preload("Orders")` 自动联表。

---

## 6. 一句话总结

> `gorm:"foreignKey:UserID"` = 告诉 GORM "对方的 `UserID` 列是外键，指向我的主键 `ID`"。查询用户时 GORM 会自动去找该用户的所有订单填进 `Orders` 字段。一对多关系 = 用户有多个订单，订单属于一个用户。`json:"user,omitempty"` 只管序列化，关联数据为空时字段不输出。
