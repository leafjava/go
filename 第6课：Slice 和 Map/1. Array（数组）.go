package main

import "fmt"

func main() {
	// 数组：固定长度
	var arr1 [5]int
	arr1[0] = 10
	arr1[1] = 20

	// 初始化数组
	arr2 := [3]string{"Alice", "Bob", "Charlie"}

	// 自动推断长度
	arr3 := [...]int{1, 2, 3, 4, 5}
	fmt.Println(arr1, arr2, arr3)
	fmt.Println("长度:", len(arr3))
}
