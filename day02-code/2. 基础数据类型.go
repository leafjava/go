package main

import "fmt"

//整数类型
//func main() {
//	var a int8 = 127
//	var b int16 = 32767
//	var c int32 = 2147483647
//	var d int64 = 9223372036854775807
//
//	// 无符号整数
//	var e uint8 = 255    // 0 ~ 255
//	var f uint16 = 65535 // 0 ~ 65535
//	var g uint32 = 4294967295
//	var h uint64 = 18446744073709551615
//
//	// int 和 uint（根据系统自动选择 32 或 64 位）
//	var i int = 100
//	var j uint = 200
//
//	fmt.Println(a, b, c, d, e, f, g, h, i, j)
//}

// 浮点数类型
//func main() {
//	var price float32 = 99.99
//	var balance float64 = 10000.123456789
//
//	fmt.Printf("价格: %.2f\n", price)
//	fmt.Printf("余额: %.2f\n", balance)
//
//	var gasPrice float64 = 1.5e-9
//	fmt.Println(gasPrice)
//}

// 字符串类型
func main() {
	var name string = "leaf"
	fmt.Println(name)

	var address string = `0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb`
	fmt.Println(address)

	greeting := "Hello, world!"
	fmt.Println(greeting)

	
}
