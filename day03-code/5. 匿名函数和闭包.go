package main

import "fmt"

func main() {
	// 匿名函数
	add := func(a, b int) int {
		return a + b
	}

	result := add(10, 20)

	fmt.Println(result)

	// 立即执行函数
	func(name string) {
		fmt.Println(name)
	}("leaf")
}
