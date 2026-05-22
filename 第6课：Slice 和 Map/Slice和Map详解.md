# 第6课：Slice 和 Map 详解

> Go 语言中最重要的两种集合类型——掌握它们才算真正入门

---

## 目录

1. [前置概念：值类型 vs 引用类型](#1-前置概念值类型-vs-引用类型)
2. [Array（数组）—— Slice 的底层](#2-array数组--slice-的底层)
3. [Slice（切片）—— 动态数组](#3-slice切片--动态数组)
4. [Map（映射）—— 键值对集合](#4-map映射--键值对集合)
5. [Slice 和 Map 组合使用](#5-slice-和-map-组合使用)
6. [常见错误与陷阱](#6-常见错误与陷阱)
7. [Web3 实战场景](#7-web3-实战场景)

---

## 1. 前置概念：值类型 vs 引用类型

在理解 Slice 和 Map 之前，必须先搞清楚 Go 里两种类型传递方式的区别。

### 值类型

变量存的是**数据本身**，赋值给另一个变量会**拷贝一份**。

```go
a := 10
b := a     // b = 10，拷贝了 a 的值
b = 20     // 修改 b 不影响 a
fmt.Println(a) // 10
fmt.Println(b) // 20
```

值类型包括：`int`、`float64`、`string`、`bool`、数组、结构体

### 引用类型

变量存的是**指向底层数据的指针/引用**，赋值给另一个变量只拷贝指针，**底层共享同一份数据**。

```go
s1 := []int{1, 2, 3}
s2 := s1        // s2 和 s1 指向同一个底层数组
s2[0] = 999     // 修改 s2
fmt.Println(s1) // [999 2 3] —— s1 也变了！
```

引用类型包括：**切片 (Slice)**、**映射 (Map)**、**通道 (Channel)**、**接口 (Interface)**、**指针 (\*T)**

```text
s1 ──→ 底层数组 [999, 2, 3]
s2 ──→ 同一个底层数组 ↗
```

> **核心区别**：值类型赋值 = 复印一份；引用类型赋值 = 共享同一份

---

## 2. Array（数组）—— Slice 的底层

数组是固定长度的同类型元素集合，长度是**类型的一部分**。

### 2.1 创建数组

```go
var arr1 [5]int                                  // [0 0 0 0 0]，零值初始化
arr2 := [3]string{"Alice", "Bob", "Charlie"}     // 初始化+赋值
arr3 := [...]int{1, 2, 3, 4, 5}                 // ... 让编译器自动数长度
```

### 2.2 关键特性

```go
arr4 := [5]int{1, 2, 3, 4, 5}
fmt.Println(len(arr4)) // 5 —— 内置函数 len() 获取长度

// ⚠️ 长度是类型的一部分
var a [3]int
var b [5]int
// a = b  // ❌ 编译错误：类型不匹配！[3]int 和 [5]int 是不同的类型
```

### 2.3 数组是值类型

```go
arr1 := [3]int{1, 2, 3}
arr2 := arr1      // 完全拷贝了一份
arr2[0] = 999
fmt.Println(arr1) // [1 2 3] —— arr1 不受影响
```

> 这就是为什么 Go 里几乎不用数组，而是用 Slice——数组一赋值就全量拷贝，大数组性能很差。

---

## 3. Slice（切片）—— 动态数组

Slice 是 Go 里最常用的数据结构。它是对底层数组的一个"视图"。

### 3.1 Slice 的内存结构

每个 Slice 由三部分组成（24 字节）：

```text
┌──────────┬──────────┬──────────┐
│  ptr     │  len     │  cap     │
│ (指针)    │ (长度)   │ (容量)   │
│ 8 bytes  │ 8 bytes  │ 8 bytes  │
└──────────┴──────────┴──────────┘
         ↘ 指向底层数组
```

| 字段 | 含义 | 类比 |
|------|------|------|
| **ptr** | 指向底层数组的指针 | 书签指向书页 |
| **len** | 当前元素个数 | 已经写了多少页 |
| **cap** | 从 ptr 开始到底层数组末尾的元素个数 | 书还剩多少空白页 |

```go
s := make([]int, 3, 5)

// len = 3，cap = 5
//
//   ptr → [0, 0, 0, _, _]
//          ← len=3 →←cap剩余→
//          ←──── cap=5 ────→
//
// 可直接访问的下标：0, 1, 2
// append 时：直接用第 3、4 的位置，不需要扩容
```

### 3.2 创建 Slice 的 5 种方式

```go
// ① var 声明 —— nil 切片（len=0, cap=0, ptr=nil）
var s1 []int
fmt.Println(s1 == nil) // true

// ② 字面量空切片 —— 非 nil（len=0, cap=0, ptr 指向某个地址）
s2 := []int{}
fmt.Println(s2 == nil) // false

// ③ 字面量初始化
s3 := []int{1, 2, 3, 4, 5} // len=5, cap=5

// ④ make 指定长度 —— 长度 = 容量
s4 := make([]int, 5)        // len=5, cap=5 → [0, 0, 0, 0, 0]

// ⑤ make 分别指定长度和容量
s5 := make([]int, 5, 10)    // len=5, cap=10 → [0, 0, 0, 0, 0]（前5个已初始化，后5个预留给 append）
```

**make 参数详解**：

```text
make([]int, 5)       →  只给2个参数：长度=容量=5
make([]int, 5, 10)   →  给3个参数：长度5，容量10
                         ↑ 第2个是 len    ↑ 第3个是 cap
```

**什么时候用哪个？**

| 场景 | 推荐写法 |
|------|----------|
| 不确定有没有数据（可能是 nil） | `var s []int` |
| 明确需要空切片（JSON 序列化返回 `[]` 而非 `null`） | `s := []int{}` |
| 已知初始值 | `s := []int{1,2,3}` |
| 已知长度，要直接通过下标赋值 | `s := make([]int, 5)` |
| 已知大概容量，要频繁 append | `s := make([]int, 0, 100)` |

### 3.3 nil 切片 vs 空切片

```go
var s1 []int       // nil 切片：ptr = nil
s2 := []int{}      // 空切片：ptr 指向一个零长度数组
s3 := make([]int, 0) // 空切片

fmt.Println(s1 == nil) // true
fmt.Println(s2 == nil) // false
fmt.Println(s3 == nil) // false

fmt.Println(len(s1)) // 0 —— 都能正常用 len()
fmt.Println(len(s2)) // 0

// 实际中两者可以互换使用——append、len、range 对 nil 切片都不会 panic
s1 = append(s1, 1) // ✅ nil 切片也能正常 append
```

> **唯一注意**：JSON 序列化时，nil 切片 → `null`，空切片 → `[]`

### 3.4 append —— 核心操作

```go
s := []int{1, 2, 3}
s = append(s, 4)       // [1 2 3 4]
s = append(s, 5, 6, 7) // [1 2 3 4 5 6 7]
```

**扩容机制**：当 len 达到 cap 时，append 会自动：
1. 分配一个更大的底层数组（通常是 2 倍增长）
2. 把旧数据拷贝过去
3. 返回指向新数组的 Slice

```go
s := make([]int, 0, 2)
fmt.Println(len(s), cap(s)) // 0 2

s = append(s, 1)
s = append(s, 2)
fmt.Println(len(s), cap(s)) // 2 2

s = append(s, 3)            // 触发扩容！
fmt.Println(len(s), cap(s)) // 3 4 —— cap 翻倍了
```

> 如果你能预估元素数量，用 `make([]int, 0, 预估数量)` 预分配容量，避免多次扩容拷贝。

### 3.5 切片操作（Slicing）

```go
s := []int{1, 2, 3, 4, 5}

s[1:4]  // [2 3 4]     从下标 1 开始，到 4 之前（包头不包尾）
s[:3]   // [1 2 3]      从开头，等价于 s[0:3]
s[3:]   // [4 5]        到末尾，等价于 s[3:5]
s[:]    // [1 2 3 4 5]  完整拷贝...但不是深拷贝！
```

**⚠️ 切片表达式不会拷贝数据**，新旧 Slice 共享同一个底层数组：

```go
original := []int{1, 2, 3, 4, 5}
sub := original[1:3]   // [2 3]，和 original 共享底层数组
sub[0] = 999
fmt.Println(original)  // [1 999 3 4 5] —— original 也被改了！
```

**如果真要独立拷贝，用 `copy`**：

```go
original := []int{1, 2, 3}
copied := make([]int, len(original))
copy(copied, original)  // copy 的目标必须先用 make 分配好长度
copied[0] = 999
fmt.Println(original) // [1 2 3] —— 不受影响
```

### 3.6 删除元素

Go 没有 `remove` 方法，通过切片拼接实现（跳过要删的那个）：

```go
s := []string{"A", "B", "C", "D"}
indexToRemove := 1                     // 要删掉 "B"
s = append(s[:indexToRemove], s[indexToRemove+1:]...)
//         ↑ 前半部分 [A]       ↑ 后半部分 [C D]

// 拆解：
// s[:1]  = [A]
// s[2:]  = [C D]
// append([A], [C D]...) → [A C D]

fmt.Println(s) // [A C D]
```

### 3.7 遍历

```go
for i, v := range s {  // i=索引, v=元素值（拷贝！）
    fmt.Println(i, v)
}

for _, v := range s {  // 只要值，不要索引
}

for i := range s {     // 只要索引
}

for i := 0; i < len(s); i++ { // 传统 for 循环
}
```

> ⚠️ `range` 出来的 `v` 是元素的**拷贝**，修改 `v` 不影响原切片。要改原值必须用 `s[i]`。

### 3.8 常见操作速查

```go
// 拷贝
dst := make([]int, len(src))
copy(dst, src)

// 清空（保留底层数组，len 归零）
s = s[:0]

// 完全释放（让 GC 回收底层数组）
s = nil

// 反转（Go 1.22+ 无内置，手写）
for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
    s[i], s[j] = s[j], s[i]
}
```

---

## 4. Map（映射）—— 键值对集合

Map 是无序的键值对集合，底层是哈希表。

### 4.1 创建 Map 的 4 种方式

```go
// ① var 声明 —— nil map（读写都会 panic！！）
var m1 map[string]int
// m1["key"] = 1  // ❌ panic: assignment to entry in nil map
fmt.Println(m1 == nil) // true

// ② 字面量空 map
m2 := map[string]int{}
m2["key"] = 1 // ✅ 可以写入

// ③ 字面量初始化
m3 := map[string]int{
    "Alice": 100,
    "Bob":   200,  // 最后的逗号不能省略
}

// ④ make 创建 —— 最常用
m4 := make(map[string]int)
m4["Charlie"] = 300 // ✅ 安全
```

> **核心规则**：nil map（`var m map[K]V`）可以读，但不能写！要用 `make` 或 `{}` 初始化后才能写。

### 4.2 基本 CRUD

```go
m := make(map[string]int)

// Create / Update
m["Alice"] = 100     // 新键 → 添加
m["Alice"] = 150     // 已存在的键 → 覆盖

// Read
value := m["Alice"]  // 150
value2 := m["David"] // 0 —— 不存在的键返回该类型的零值

// 如何区分"值为0"和"键不存在"？
value, exists := m["David"] // 第二个返回值是 bool
if exists {
    fmt.Println("找到了，值为:", value)
} else {
    fmt.Println("键不存在")
}

// 惯用简写：一行判断
if v, ok := m["Alice"]; ok {
    fmt.Println("Alice 余额:", v)
}

// Delete
delete(m, "Alice")   // 删除键（即使键不存在也不会报错）
```

### 4.3 遍历

```go
for key, value := range m {
    fmt.Printf("%s: %d\n", key, value)
}

for key := range m {     // 只要键
}

for _, value := range m { // 只要值
}
```

> ⚠️ Map 遍历顺序是**随机**的——每次 `range` 顺序都可能不同。需要有序时，先把 key 取到 Slice 里排序再遍历。

### 4.4 嵌套 Map

```go
// 用户地址 → 代币名称 → 余额
userBalances := make(map[string]map[string]float64)

// ⚠️ 内层 map 也需要初始化，否则是 nil map！
user1 := "0xABC..."
userBalances[user1] = make(map[string]float64)    // 先 make 内层
userBalances[user1]["ETH"] = 10.0
userBalances[user1]["USDT"] = 1000.0
```

---

## 5. Slice 和 Map 组合使用

### 5.1 Slice 装 Map（多行表数据）

```go
// 交易记录列表
txs := []map[string]interface{}{
    {"hash": "0xaaa", "amount": 1.5, "status": "confirmed"},
    {"hash": "0xbbb", "amount": 2.0, "status": "pending"},
}
```

### 5.2 Map 装 Slice（一对多关系）

```go
// 用户 → 拥有的 NFT ID 列表
userNFTs := make(map[string][]int)
userNFTs["Alice"] = []int{1, 2, 3}
userNFTs["Bob"] = []int{4, 5}
```

### 5.3 有序遍历 Map

```go
m := map[string]int{"C": 3, "A": 1, "B": 2}

// 提取 key 到 Slice
keys := make([]string, 0, len(m))
for k := range m {
    keys = append(keys, k)
}

// 排序（Go 标准库 sort.Strings）
sort.Strings(keys)

// 按 key 顺序遍历
for _, k := range keys {
    fmt.Println(k, m[k])
}
// A 1
// B 2
// C 3
```

---

## 6. 常见错误与陷阱

### 错误1：nil Map 写入

```go
var m map[string]int // nil map
m["key"] = 1         // ❌ panic: assignment to entry in nil map
// 解决：m := make(map[string]int)
```

### 错误2：Slice 切片的 append 陷阱

```go
s1 := []int{1, 2, 3}
s2 := s1[:2]         // s2 = [1 2]，cap = 3（共享底层）
s2 = append(s2, 999) // len=3，cap 够，写入了底层第3个位置
fmt.Println(s1)      // [1 2 999] —— s1 被偷偷改了！

// 解决：使用完整切片表达式限制容量
s2 := s1[:2:2]       // s[low:high:max]，cap = max-low = 2
s2 = append(s2, 999) // 触发扩容，新底层数组，不影响 s1
```

### 错误3：range 中修改 v 无效

```go
type User struct { Name string }
users := []User{{"Alice"}, {"Bob"}}

for _, u := range users {
    u.Name = "Xxx"      // u 是拷贝，改不了原数据
}
fmt.Println(users) // [{Alice} {Bob}] —— 没变！

// 方案1：用索引
for i := range users {
    users[i].Name = "Xxx" // ✅
}

// 方案2：装指针
ptrUsers := []*User{{"Alice"}, {"Bob"}}
for _, u := range ptrUsers {
    u.Name = "Xxx"       // ✅ 指针指向原数据
}
```

### 错误4：Map 遍历中删除

```go
for k := range m {
    delete(m, k) // ✅ 遍历中删除当前键是安全的
}
```

### 错误5：Slice 的 copy 只拷贝 min(len(src), len(dst)) 个

```go
src := []int{1, 2, 3}
dst := make([]int, 0)     // len(dst)=0
copy(dst, src)            // 拷贝了 min(0, 3) = 0 个
fmt.Println(dst)          // [] —— 空的！
// 解决：dst := make([]int, len(src))
```

---

## 7. Web3 实战场景

### 场景1：交易池——Pending + Confirmed 队列

```go
type TxPool struct {
    Pending   []*Transaction // Slice：有序排队，按 Gas 排序
    Confirmed map[string]*Transaction // Map：哈希 → 交易，O(1) 查找
}
```

**选型理由**：
- Pending 用 Slice：需要按顺序打包、按 Gas 排序
- Confirmed 用 Map：用户查 `0xhash...` 需要 O(1) 响应

### 场景2：地址簿

```go
// Map：名称 → 地址，快速查询
type AddressBook struct {
    Entries map[string]string // "张三" → "0x742d..."
}
```

### 场景3：代币余额

```go
// 嵌套 Map：用户 → (代币 → 余额)
// 两层 Map，第一层 O(1) 找到用户，第二层 O(1) 找到代币
type TokenManager struct {
    Balances map[string]map[string]float64
}
```

### 场景4：NFT 集合

```go
// 双索引：Map 用于快速查找，Slice 用于有序遍历
type NFTCollection struct {
    nfts     map[int]*NFT    // TokenID → NFT（O(1) 查找）
    tokenIDs []int           // 维护插入顺序
}
```

---

## 总结：Slice 还是 Map？

| 需求 | 选 Slice | 选 Map |
|------|----------|--------|
| 按索引访问 | ✅ `s[3]` | ❌ |
| 按 key 查找 | ❌ O(n) 遍历 | ✅ O(1) |
| 保持插入顺序 | ✅ | ❌ 无序 |
| 去重判断 | ❌ 需要遍历 | ✅ key 天然唯一 |
| 排序 | ✅ `sort.Slice()` | ❌ 需先转 Slice |
| 内存占用 | 较小（24B 头+数组） | 较大（哈希表有空洞） |
| 并发安全 | ❌ 都不是线程安全的 | ❌ 都不是线程安全的 |

> 简单原则：**有顺序用 Slice，有查找用 Map，两者都配指针才是标准 Web3 写法。**

---

## 🎯 检查点

- [ ] 能说出 Slice 的三个组成部分（ptr / len / cap）
- [ ] 理解 `make([]int, 5)` 和 `make([]int, 5, 10)` 的区别
- [ ] 知道 nil Map 为什么不能写入
- [ ] 理解切片的 append 扩容机制
- [ ] 能解释为什么 range 里修改 v 无效
- [ ] 能组合使用 Slice + Map 解决问题
