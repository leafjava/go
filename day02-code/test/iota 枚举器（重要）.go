package main

import "fmt"

const (
	Sunday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

const (
	Ethereum = iota + 1
	BSC
	Polygon
	TON
)

const (
	ReadPermission = 1 << iota
	WritePermission
	DeletePermission
)

func main() {
	fmt.Println("今天是:", Wednesday)
	fmt.Println("网络:", TON)

	userPermission := ReadPermission | WritePermission
	fmt.Printf("用户权限: %b\n", userPermission) // 输出: 11
}
