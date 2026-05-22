# Q7: GORM 事务处理 + 复杂关联查询

## 目录

- [GORM 事务处理](#gorm-事务处理)
  - [手动事务](#手动事务)
  - [闭包事务（推荐）](#闭包事务推荐)
  - [嵌套事务（SavePoint）](#嵌套事务savepoint)
  - [事务隔离级别](#事务隔离级别)
- [GORM 关联查询](#gorm-关联查询)
  - [模型定义与 Tag 标签](#模型定义与-tag-标签)
  - [Preload 预加载](#preload-预加载)
  - [Joins 连接查询](#joins-连接查询)
  - [关联模式（Associations）](#关联模式associations)
- [实战：电商下单场景](#实战电商下单场景)
- [常见坑与最佳实践](#常见坑与最佳实践)

---

## GORM 事务处理

### 手动事务

需要手动 `Begin`、`Commit`、`Rollback`，适合需要在事务中做额外判断的场景。

```go
func TransferMoney(fromID, toID uint, amount float64) error {
    db := model.DB // 假设已经初始化好了 *gorm.DB

    // 开启事务
    tx := db.Begin()
    if tx.Error != nil {
        return tx.Error
    }

    // 确保最终一定会 Commit 或 Rollback
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
        }
    }()

    // 操作一：扣款
    result1 := tx.Model(&Account{}).
        Where("id = ? AND balance >= ?", fromID, amount).
        Update("balance", gorm.Expr("balance - ?", amount))
    if result1.Error != nil || result1.RowsAffected == 0 {
        tx.Rollback()
        return errors.New("扣款失败：余额不足或账户不存在")
    }

    // 操作二：加款
    result2 := tx.Model(&Account{}).
        Where("id = ?", toID).
        Update("balance", gorm.Expr("balance + ?", amount))
    if result2.Error != nil || result2.RowsAffected == 0 {
        tx.Rollback()
        return errors.New("加款失败：账户不存在")
    }

    // 提交事务
    return tx.Commit().Error
}
```

关键点：

- **必须用同一个 `tx` 实例**，用 `db` 的话就不在事务里了
- `RowsAffected == 0` 也要回滚，说明数据不满足条件
- 用 `defer + recover` 防止 panic 导致事务悬挂

### 闭包事务（推荐）

GORM 的 `Transaction` 方法自动处理 Commit 和 Rollback，返回 error 就自动回滚。

```go
func TransferMoney(fromID, toID uint, amount float64) error {
    return model.DB.Transaction(func(tx *gorm.DB) error {
        // 扣款
        result1 := tx.Model(&Account{}).
            Where("id = ? AND balance >= ?", fromID, amount).
            Update("balance", gorm.Expr("balance - ?", amount))
        if result1.Error != nil || result1.RowsAffected == 0 {
            return errors.New("扣款失败")
        }

        // 加款
        result2 := tx.Model(&Account{}).
            Where("id = ?", toID).
            Update("balance", gorm.Expr("balance + ?", amount))
        if result2.Error != nil || result2.RowsAffected == 0 {
            return errors.New("加款失败")
        }

        return nil // 返回 nil 自动 Commit
    })
}
```

两种方式对比：

| | 手动事务 | 闭包事务 |
|------|---------|---------|
| 编码量 | 多，需要手动 Rollback | 少，return error 就回滚 |
| panic 保护 | 需要自己 defer recover | GORM 内部自动 recover |
| 灵活性 | 高，可以在中间做外部调用后决定 | 一般场景够用 |
| 适用场景 | 事务中途需要做复杂判断或调用外部服务 | **推荐首选** |

### 嵌套事务（SavePoint）

GORM 支持嵌套事务，底层用数据库的 SavePoint 实现。

```go
func ComplexOrderFlow(orderID uint) error {
    return model.DB.Transaction(func(tx *gorm.DB) error {
        // 外层事务：创建订单
        if err := tx.Create(&Order{ID: orderID}).Error; err != nil {
            return err
        }

        // 内层事务：扣库存
        err := tx.Transaction(func(tx2 *gorm.DB) error {
            // tx2 实际上是同一个 tx，GORM 在这里创建了一个 SavePoint
            result := tx2.Model(&Product{}).
                Where("id = ? AND stock > 0", 1).
                Update("stock", gorm.Expr("stock - 1"))
            if result.RowsAffected == 0 {
                return errors.New("库存不足")
            }
            return nil
        })
        if err != nil {
            // 内层失败，回滚到 SavePoint，外层可以继续或返回
            return err // 返回 error，整个外层也回滚
        }

        // 外层继续：扣优惠券
        result := tx.Model(&Coupon{}).
            Where("id = ? AND status = 'unused'", 1).
            Update("status", "used")
        if result.RowsAffected == 0 {
            return errors.New("优惠券无效")
        }

        return nil
    })
}
```

SavePoint 的执行过程：

```
BEGIN
  -- 外层操作
  INSERT INTO orders ...

  SAVEPOINT sp1       -- 内层事务开始
  UPDATE products SET stock = stock - 1 ...
  RELEASE SAVEPOINT sp1 -- 内层成功

  -- 外层继续
  UPDATE coupons SET status = 'used' ...

  -- 如果内层失败：
  ROLLBACK TO SAVEPOINT sp1 -- 只回滚内层，外层不受影响
COMMIT
```

### 事务隔离级别

GORM 支持设置隔离级别：

```go
// 可重复读（MySQL 默认）
err := db.Transaction(func(tx *gorm.DB) error {
    // ...
}, &sql.TxOptions{
    Isolation: sql.LevelRepeatableRead,
})

// 读已提交
err := db.Transaction(func(tx *gorm.DB) error {
    // ...
}, &sql.TxOptions{
    Isolation: sql.LevelReadCommitted,
})
```

四种隔离级别对比：

| 隔离级别 | 脏读 | 不可重复读 | 幻读 | 并发性能 |
|---------|------|----------|------|---------|
| Read Uncommitted | 可能 | 可能 | 可能 | 最高 |
| Read Committed | 避免 | 可能 | 可能 | 高 |
| Repeatable Read | 避免 | 避免 | 可能 | 中 |
| Serializable | 避免 | 避免 | 避免 | 最低 |

---

## GORM 关联查询

### 模型定义与 Tag 标签

以一个电商数据库为例：

```go
// 用户：一个用户有多个订单
type User struct {
    ID       uint   `gorm:"primaryKey"`
    Username string `gorm:"size:64;not null"`
    Orders   []Order `gorm:"foreignKey:UserID"` // Has Many
}

// 订单：一个订单属于一个用户，且包含多个订单项
type Order struct {
    ID          uint      `gorm:"primaryKey"`
    UserID      uint      `gorm:"index;not null"` // 外键
    OrderNo     string    `gorm:"size:32;uniqueIndex"`
    TotalAmount float64
    Status      string    `gorm:"default:pending"`
    User        User      `gorm:"foreignKey:UserID"` // Belongs To
    Items       []OrderItem `gorm:"foreignKey:OrderID"` // Has Many
}

// 订单项：一个订单项属于一个订单，也属于一个商品
type OrderItem struct {
    ID        uint    `gorm:"primaryKey"`
    OrderID   uint    `gorm:"index;not null"`
    ProductID uint    `gorm:"index;not null"`
    Quantity  int     `gorm:"not null"`
    Price     float64 `gorm:"not null"`
    Order     Order   `gorm:"foreignKey:OrderID"`
    Product   Product `gorm:"foreignKey:ProductID"`
}

// 商品：一个商品有多个标签（多对多）
type Product struct {
    ID       uint    `gorm:"primaryKey"`
    Name     string  `gorm:"size:128;not null"`
    Price    float64
    Stock    int
    Tags     []Tag   `gorm:"many2many:product_tags;"` // Many To Many
}

// 标签
type Tag struct {
    ID       uint      `gorm:"primaryKey"`
    Name     string    `gorm:"size:32"`
    Products []Product `gorm:"many2many:product_tags;"`
}
```

常用 GORM Tag 速查：

| Tag | 说明 | 示例 |
|-----|------|------|
| `primaryKey` | 主键 | `gorm:"primaryKey"` |
| `foreignKey:UserID` | 指定外键字段 | `gorm:"foreignKey:UserID"` |
| `references:ID` | 指定参照字段 | `gorm:"references:ID"` |
| `many2many:table` | 多对多中间表 | `gorm:"many2many:product_tags;"` |
| `constraint:OnDelete:CASCADE` | 级联删除 | `gorm:"constraint:OnDelete:CASCADE;"` |
| `index` | 创建索引 | `gorm:"index"` |
| `uniqueIndex` | 唯一索引 | `gorm:"uniqueIndex"` |
| `default:value` | 默认值 | `gorm:"default:pending"` |
| `not null` | 非空约束 | `gorm:"not null"` |
| `size:128` | 字段长度 | `gorm:"size:128"` |

### Preload 预加载

Preload 是 GORM 解决 N+1 查询的核心方法，分成多条 SQL 查询，然后在内存中组装。

```go
// ❌ 不预加载：N+1 问题
var orders []Order
db.Find(&orders)
for _, order := range orders {
    // 每条订单又查一次数据库拿 Items
    db.Where("order_id = ?", order.ID).Find(&order.Items)
}
// 输出：
// SELECT * FROM orders                           -- 1 次
// SELECT * FROM order_items WHERE order_id = 1  -- N 次
// SELECT * FROM order_items WHERE order_id = 2  -- N 次
// ...

// ✅ Preload：只有 2 条 SQL
db.Preload("Items").Find(&orders)
// SELECT * FROM orders
// SELECT * FROM order_items WHERE order_id IN (1,2,3,...)
```

**多层嵌套预加载：**

```go
// 查订单 → 带出订单项 → 再带出订单项中的商品
db.Preload("Items.Product").Find(&orders)
```

**带条件的预加载：**

```go
// 只预加载价格 > 100 的订单项
db.Preload("Items", "price > ?", 100).Find(&orders)

// 用闭包实现更复杂的条件
db.Preload("Items", func(db *gorm.DB) *gorm.DB {
    return db.Where("price > ?", 100).Order("price DESC")
}).Find(&orders)
```

**预加载所有关联：**

```go
db.Preload(clause.Associations).Find(&orders)
// 会预加载 User、Items 等所有关联，谨慎使用
```

### Joins 连接查询

Joins 用一条 SQL 完成关联查询，适合需要跨表条件的场景。

**基础 Joins：**

```go
// 查询订单，同时 JOIN 用户表做条件过滤
var orders []Order
db.Joins("User").Where("User.username = ?", "张三").Find(&orders)
// SQL: SELECT orders.* FROM orders
//      LEFT JOIN users ON orders.user_id = users.id
//      WHERE users.username = '张三'
```

**根据关联表字段筛选后预加载：**

```go
// 先 JOIN 筛选，再 Preload 填充
var users []User
db.Joins("Orders").
    Where("Orders.total_amount > ?", 500).
    Preload("Orders").
    Find(&users)
// 只查出有订单金额 > 500 的用户，并预加载他们的所有订单
```

**Joins 和 Preload 的区别：**

| | Joins | Preload |
|------|-------|---------|
| SQL 数量 | 1 条 JOIN 查询 | 2+ 条独立查询 |
| 筛选关联数据 | 可以 | 可以（带条件 Preload） |
| 填充关联字段 | 默认不填充 | 默认填充到结构体 |
| 结果集 | 可能产生笛卡尔积 | 各自独立查询，内存组装 |
| 典型场景 | "查有高额订单的用户" | "查用户并带出他的所有订单" |

### 关联模式（Associations）

关联模式用于操作关联数据，追加、替换、删除关联。

```go
var order Order
db.First(&order, 1)

// 追加关联：给订单添加订单项
db.Model(&order).Association("Items").Append(&OrderItem{
    ProductID: 1,
    Quantity:  2,
    Price:     99.0,
})

// 替换关联：用新的一组订单项替换旧的
db.Model(&order).Association("Items").Replace([]OrderItem{...})

// 删除关联：删除指定的关联记录
db.Model(&order).Association("Items").Delete(&item1, &item2)

// 清空关联：删除所有关联
db.Model(&order).Association("Items").Clear()

// 统计关联数量
count := db.Model(&order).Association("Items").Count()

// 查询关联数据
var items []OrderItem
db.Model(&order).Association("Items").Find(&items)
```

---

## 实战：电商下单场景

把事务和关联查询结合起来，实现一个完整的下单流程。

### 需求

1. 下单时：扣库存 → 创建订单 → 创建订单项 → 扣优惠券
2. 四个操作必须在同一个事务中
3. 查询订单详情时，要带出订单项和商品信息

### 下单代码

```go
type CreateOrderReq struct {
    UserID    uint   `json:"user_id"`
    CouponID  *uint  `json:"coupon_id"` // 可选
    Items     []struct {
        ProductID uint `json:"product_id"`
        Quantity  int  `json:"quantity"`
    } `json:"items"`
}

func CreateOrder(req CreateOrderReq) (*Order, error) {
    var order *Order

    err := model.DB.Transaction(func(tx *gorm.DB) error {
        // 1. 锁定商品并扣库存
        for _, item := range req.Items {
            result := tx.Model(&Product{}).
                Where("id = ? AND stock >= ?", item.ProductID, item.Quantity).
                Update("stock", gorm.Expr("stock - ?", item.Quantity))
            if result.Error != nil || result.RowsAffected == 0 {
                return fmt.Errorf("商品 %d 库存不足", item.ProductID)
            }

            // 2. 查询商品价格（在事务内查，保证读到的价格和扣库存一致）
            var product Product
            if err := tx.First(&product, item.ProductID).Error; err != nil {
                return err
            }

            // 计算总价
            totalPrice := product.Price * float64(item.Quantity)

            // 3. 创建订单项
            orderItem := OrderItem{
                ProductID: item.ProductID,
                Quantity:  item.Quantity,
                Price:     totalPrice,
            }
            // 这里暂时不需要 orderID，创建订单后再关联
        }

        // 4. 创建订单
        order = &Order{
            UserID:      req.UserID,
            OrderNo:     generateOrderNo(),
            Status:      "pending",
            TotalAmount: calculateTotal(req.Items), // 上面循环中累加
        }
        if err := tx.Create(order).Error; err != nil {
            return err
        }

        // 5. 更新订单项的 OrderID
        // （实际开发中这一步可以在创建订单时一并处理）

        // 6. 扣优惠券（如果有）
        if req.CouponID != nil {
            result := tx.Model(&Coupon{}).
                Where("id = ? AND user_id = ? AND status = 'unused'",
                    *req.CouponID, req.UserID).
                Update("status", "used")
            if result.RowsAffected == 0 {
                return errors.New("优惠券无效")
            }
        }

        return nil
    })

    return order, err
}

func generateOrderNo() string {
    return fmt.Sprintf("ORD%s", time.Now().Format("20060102150405"))
}
```

### 查询订单详情（关联查询）

```go
func GetOrderDetail(orderID uint) (*Order, error) {
    var order Order
    err := model.DB.
        Preload("Items.Product.Tags"). // 订单项 → 商品 → 标签
        Preload("User").               // 用户信息
        First(&order, orderID).Error
    return &order, err
}
```

执行时 GORM 会生成 4 条 SQL：

```sql
-- 1. 查订单
SELECT * FROM orders WHERE id = ?

-- 2. 查订单项
SELECT * FROM order_items WHERE order_id IN (?)

-- 3. 查商品
SELECT * FROM products WHERE id IN (?, ?, ?)

-- 4. 查标签（多对多中间表）
SELECT * FROM tags
INNER JOIN product_tags ON product_tags.tag_id = tags.id
WHERE product_tags.product_id IN (?, ?, ?)
```

4 条 SQL 而不是 N+1 条，性能远好于循环查。

---

## 常见坑与最佳实践

### 坑 1：忘记用 tx 而用了 db

```go
// ❌ 错误：Create 用的是 db 而不是 tx，不在事务里
db.Transaction(func(tx *gorm.DB) error {
    tx.Model(&Product{}).Update("stock", gorm.Expr("stock - 1"))
    db.Create(&Order{}) // 这个没在事务里！
    return nil
})

// ✅ 正确：统一用闭包参数的 tx
db.Transaction(func(tx *gorm.DB) error {
    tx.Model(&Product{}).Update("stock", gorm.Expr("stock - 1"))
    tx.Create(&Order{})
    return nil
})
```

### 坑 2：Preload 和 Joins 混用时字段名

```go
// ❌ Joins 的参数是关联名，不是表名
db.Joins("Product")   // ✅ 用结构体中定义的关联名
db.Joins("products")  // ❌ 用数据库表名，不生效
```

### 坑 3：多对多中间表自动命名

GORM 多对多中间表名默认按字母排序后拼接：

```go
// User 和 Role → 中间表 "user_roles"
// Product 和 Tag   → 中间表 "product_tags"
// 如果不一致，用 many2many tag 显式指定
Tags []Tag `gorm:"many2many:custom_table;"`
```

### 坑 4：事务中查询用 FOR UPDATE 防并发

高并发场景下扣库存，仅靠 `WHERE stock >= ?` 不够：

```go
// ❌ 两个事务可能同时读到 stock=5，都认为够扣
tx.Where("id = ? AND stock >= ?", id, qty).Update(...)

// ✅ 用 FOR UPDATE 锁定行，串行化读写
tx.Clauses(clause.Locking{Strength: "UPDATE"}).
    Where("id = ? AND stock >= ?", id, qty).
    First(&product)
// 第二个事务必须等第一个提交后才能读
```

### 最佳实践总结

1. **事务尽量短小**：事务内不要做外部 API 调用、文件读写等耗时操作
2. **优先用闭包事务**：除非有特殊需求，否则 `db.Transaction(func(tx *gorm.DB) error {...})` 是最安全的选择
3. **Preload 优于手动循环**：用 `Preload` 解决 N+1，比手写 `for range` 清晰且不易出错
4. **Joins 用于跨表条件**：需要"查有高额订单的用户"这种跨表筛选时用 Joins
5. **高并发用 FOR UPDATE**：库存扣减、余额变动等场景必须加行锁
6. **关联名一致**：`Joins("User")`、`Preload("User")` 中的名字对应结构体字段名，不是表名
