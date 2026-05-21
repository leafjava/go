package main

import (
	"fmt"
	"time"
)

// TODO: 实现函数执行时间统计（使用 defer）
func measureTime(funcName string) func() {
	// 返回一个函数，用于计算执行时间
	// 提示：记录开始时间，返回的函数中计算时间差
	startTime := time.Now()
	fmt.Printf("[%s] 开始执行\n", funcName)

	return func() {
		duration := time.Since(startTime)
		fmt.Printf("[%s] 执行完成，耗时: %v\n", funcName, duration)
	}
}

// TODO: 实现资源清理函数
func processTransaction(txHash string) error {
	// 使用 defer 确保资源被正确清理
	// 1. 打印 "开始处理交易"
	fmt.Println("开始处理交易:", txHash)

	// 2. defer 打印 "清理资源"
	defer fmt.Println("清理资源")

	// 3. defer 打印 "关闭连接"
	defer fmt.Println("关闭连接")

	// 4. 模拟处理逻辑
	fmt.Println("验证交易...")
	time.Sleep(50 * time.Millisecond)
	fmt.Println("广播交易...")

	return nil
}

func main() {
	// 测试时间统计
	defer measureTime("main")()

	// 模拟耗时操作
	time.Sleep(100 * time.Millisecond)

	// 测试资源清理
	processTransaction("0xabc123")
}
