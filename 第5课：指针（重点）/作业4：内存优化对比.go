package main

import (
	"fmt"
	"time"
)

// LargeData 模拟一个大结构体（10万个 int，每个 8 字节 ≈ 800KB）
type LargeData struct {
	Data [100000]int // 数组字段，整体大小约 800KB
}

// processByValue 值传递：每次调用都会拷贝完整的 800KB 数据
func processByValue(data LargeData) {
	// 模拟处理
	_ = data.Data[0] // 仅读取第一个元素，但这之前已经拷贝了整个结构体
}

// processByPointer 指针传递：每次调用只传 8 字节的指针地址
func processByPointer(data *LargeData) {
	// 模拟处理
	_ = data.Data[0] // 通过指针间接读取，无数据拷贝
}

func main() {
	data := LargeData{}                              // 创建一个约 800KB 的结构体

	// —————— 值传递性能测试 ——————
	fmt.Println("===== 大结构体：值传递 vs 指针传递 =====\n")
	start := time.Now()                              // time.Now() 获取当前时间
	for i := 0; i < 1000; i++ {                      // 循环调用 1000 次
		processByValue(data)                         // 每次拷贝整个 800KB
	}
	valueTime := time.Since(start)                   // Since 返回从 start 到现在的耗时
	fmt.Printf("值传递耗时:  %v  （每次拷贝 800KB）\n", valueTime)

	// —————— 指针传递性能测试 ——————
	start = time.Now()
	for i := 0; i < 1000; i++ {
		processByPointer(&data)                      // &data 取地址，每次只传 8 字节
	}
	pointerTime := time.Since(start)
	fmt.Printf("指针传递耗时: %v  （每次只传 8 字节指针）\n", pointerTime)

	// —————— 性能倍数 ——————
	fmt.Println("\n===== 性能倍数 =====")
	ratio := float64(valueTime) / float64(pointerTime) // 值传递时间是指针的多少倍
	fmt.Printf("值传递 / 指针传递 ≈ %.2f 倍\n", ratio)
	fmt.Println("👉 数据越大，指针优势越明显")

	// —————— 小结构体对比（值传递 vs 指针传递差距不大）——————
	fmt.Println("\n===== 小结构体对比（int，8 字节）=====\n")
	testSmallStruct()

	// —————— 逃逸分析示例 ——————
	fmt.Println("===== 逃逸分析示意 =====\n")
	testEscape()

	// —————— 最佳实践总结 ——————
	Summary()
}

// ========== 辅助函数 ==========

// SmallData 小结构体：只有一个 int 字段，大小仅 8 字节
type SmallData struct {
	Value int // 8 字节，和指针一样大
}

func processSmallValue(d SmallData) { _ = d.Value }   // 值拷贝 8 字节，直接在栈上
func processSmallPointer(d *SmallData) { _ = d.Value } // 指针 8 字节 + 解引用开销

func testSmallStruct() {
	d := SmallData{Value: 42}

	start := time.Now()
	for i := 0; i < 1000000; i++ {                   // 100 万次循环
		processSmallValue(d)                          // 值拷贝，每次 8 字节
	}
	fmt.Printf("小结构体-值传递:  %v\n", time.Since(start))

	start = time.Now()
	for i := 0; i < 1000000; i++ {
		processSmallPointer(&d)                       // 指针传递 + 每次解引用
	}
	fmt.Printf("小结构体-指针传递: %v\n", time.Since(start))
	fmt.Println("👉 小结构体两者差距不大，值传递可能更快（无解引用开销 + 栈分配）")
}

// ========== 逃逸分析 ==========

// returnValue 返回结构体值：数据在栈上创建，不逃逸
func returnValue() Person {
	return Person{Name: "Alice", Age: 20}            // 值拷贝返回，生命周期在函数内，栈上分配
}

// returnPointer 返回指针：数据逃逸到堆，靠 GC 回收
func returnPointer() *Person {
	return &Person{Name: "Bob", Age: 30}             // & 取地址返回→编译器发现"外部还要用"→堆分配
}

type Person struct {
	Name string // 字符串字段
	Age  int    // 年龄
}

func testEscape() {
	_ = returnValue()                                // 值返回：栈分配，函数结束即释放
	_ = returnPointer()                              // 指针返回：堆分配，GC 负责回收

	fmt.Println("returnValue()  → 栈分配（快，函数结束自动回收）")
	fmt.Println("returnPointer() → 堆分配（慢，依赖 GC 扫描回收）")
	fmt.Println()
	fmt.Println("Go 编译器逃逸分析规则：")
	fmt.Println("  · 函数返回指针 → 逃逸到堆")
	fmt.Println("  · 函数内用完 → 留在栈上")
	fmt.Println("👉 不要为了'省拷贝'盲目返回指针，注意逃逸带来的 GC 开销")
}

// ========== 最佳实践总结 ==========

func Summary() {
	fmt.Println("\n==================== 最佳实践 ====================\n")
	fmt.Println(" 场景                  | 推荐      | 原因")
	fmt.Println("-----------------------|-----------|-----------------------")
	fmt.Println(" 大结构体（> 100 字节）  | 指针传递   | 避免大量内存拷贝")
	fmt.Println(" 需要修改接收者          | 指针接收者  | 值接收者改不了原值")
	fmt.Println(" 小结构体 / 基本类型     | 值传递     | 栈分配更快，无 GC 压力")
	fmt.Println(" 并发场景               | 指针+Mutex | 同一块内存加锁保护")
	fmt.Println(" map / slice / channel  | 直接用    | 它们本身已是引用类型")
	fmt.Println(" 返回局部计算结果        | 值返回     | 避免不必要的堆逃逸")
	fmt.Println()
	fmt.Println("口诀：小用值，大用针，要改原值必用针，返回优先返回值")
}
