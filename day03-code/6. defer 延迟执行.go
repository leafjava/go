package main

import "fmt"

// defer：函数返回前执行（类似 Java 的 finally）
func processFile() {
	fmt.Println("1.打开文件")
	defer fmt.Println("4. 关闭文件")

	fmt.Println("2. 读取文件")
	fmt.Println("3. 处理数据")
}

// 多个 defer 按 LIFO（后进先出）顺序执行
func multipleDefer() {
	defer fmt.Println("第一个 defer")
	defer fmt.Println("第二个 defer")
	defer fmt.Println("第三个 defer")

	fmt.Println("函数体")
}

// Web3 实战：确保释放资源
func connectBlockchain() error {
	fmt.Println("连接区块链节点...")
	defer fmt.Println("断开区块链连接")

	// 模拟操作
	fmt.Println("查询余额...")
	fmt.Println("发送交易...")

	return nil
}

func main() {
	processFile()
	multipleDefer()
}
