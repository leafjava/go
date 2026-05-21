package main

import "fmt"

// 可变参数（类似 Java 的 String... args）
func sum(numbers ...int) int {
	total := 0
	for _, num := range numbers {
		total += num
	}
	return total
}

// 格式化日志
func log(level string, messages ...string) {
	fmt.Printf("[%s]", level)
	for _, msg := range messages {
		fmt.Print(msg, " ")
	}
	fmt.Println()
}

func main() {
	fmt.Println(sum(1, 2, 3))
	fmt.Println(sum(1, 2, 3, 4, 5))

	nums := []int{10, 20, 30}
	fmt.Println(sum(nums...))

	log("INFO", "服务启动", "端口:8080")
}
