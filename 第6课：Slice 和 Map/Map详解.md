# Map 详解

## 什么是 Map

Map 是 Go 的**键值对集合**——通过 key 直接查到 value，时间复杂度 O(1)。

## 类型写法

```go
map[KeyType]ValueType

map[string]int        // 字符串 → 整数
map[string]float64    // 字符串 → 浮点数
map[int]string        // 整数 → 字符串
map[string][]int      // 字符串 → 整数切片
map[string]map[string]float64  // 嵌套 Map
```

## 创建 Map 的 4 种方式

```go
// ① var 声明 —— nil map，不能写！
var m1 map[string]int
// m1["a"] = 1  // ❌ panic

// ② 字面量空 map
m2 := map[string]int{}
m2["a"] = 1  // ✅

// ③ 字面量初始化
m3 := map[string]int{"Alice": 100, "Bob": 200}

// ④ make 创建 —— 最常用
m4 := make(map[string]int)
m4["a"] = 1  // ✅
```

## 基本操作

```go
m := make(map[string]int)

// 写（添加/修改）
m["Alice"] = 100   // 添加
m["Alice"] = 150   // 覆盖

// 读（取值 + 判断是否存在）
v := m["Alice"]           // 只取值（不存在返回 int 的零值 0）
v, ok := m["David"]       // v=0, ok=false → 键不存在
v, ok := m["Bob记录0"]    // v=0, ok=true  → 键存在，值就是 0

// 惯用一行写法
if v, ok := m["Alice"]; ok {
    fmt.Println("值:", v)
}

// 删除
delete(m, "Alice")  // 键不存在也不会报错

// 长度
len(m)  // 返回键值对数量
```

## 两个返回值的关键点

`ok` 用来区分"值就是零值"和"键不存在"：

```go
m := map[string]int{"Bob": 0}

v, ok := m["Bob"]    // v=0, ok=true   → 存在，值就是 0
v, ok = m["David"]   // v=0, ok=false  → 不存在
```

## 遍历（无序）

```go
for key, value := range m { }  // 键+值
for key := range m { }         // 只要键
for _, value := range m { }    // 只要值
```

> Map 遍历顺序是随机的，需要有序时先把 key 取到 Slice 里排序再遍历。

## nil Map 陷阱

```go
var m map[string]int  // nil map
_ = m["a"]            // ✅ 读取不存在的 key 返回零值，不 panic
m["a"] = 1            // ❌ panic: assignment to entry in nil map
```

**结论**：声明后记得 `make`，或者直接用 `map[string]int{}`。
