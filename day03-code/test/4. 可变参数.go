package main

import "fmt"

func sum(numbers ...int) int {
	total := 0
	for _, num := range numbers {
		total += num
	}
	return total
}

func log(level string, messages ...string) {
	fmt.Printf("[%s] ", level)
	for _, msg := range messages {
		fmt.Print(msg, " ")
	}
	fmt.Println()
}

func main() {
	fmt.Println(sum(1, 2, 3))
	nums := []int{10, 20, 30}
	fmt.Println(sum(nums...))

	log("INFO", "服务启动", "端口:8080")
	log("ERROR", "连接失败", "重试中...")
}
