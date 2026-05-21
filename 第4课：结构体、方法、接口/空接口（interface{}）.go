package main

import "fmt"

func printAnything(v interface{}) {
	fmt.Printf("类型: %T, 值: %v\n", v, v)
}

func processValue(v interface{}) {
	switch val := v.(type) {
	case int:
		fmt.Println("整数:", val*2)
	case string:
		fmt.Println("字符串", val+"World")
	case float64:
		fmt.Printf("浮点数:%.2f\n", val)
	default:
		fmt.Println("未知类型")
	}
}

func main() {
	printAnything(42)
	printAnything("hello")
	printAnything(3.14)
	printAnything(true)

	fmt.Println("---")

	processValue(100)
	processValue("Hello")
	processValue(99.99)
}
