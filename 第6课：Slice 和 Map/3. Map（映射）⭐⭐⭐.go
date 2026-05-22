package main

import "fmt"

func main() {
	// 创建 Map
	//var m1 map[string]int
	//m2 := map[string]int{}
	//m3 := map[string]int{
	//	"Alice": 100,
	//	"Bob":   200,
	//}
	m4 := make(map[string]int)

	// 添加/修改
	m4["Charlie"] = 300
	m4["Alice"] = 150

	// 读取
	value := m4["Alice"]
	fmt.Println("Alice:", value)

	// 检查键是否存在
	value, exists := m4["David"]
	if exists {
		fmt.Println("David:", value)
	} else {
		fmt.Println("Not exists")
	}

	// 删除
	delete(m4, "Alice")

	// 遍历
	for key, value := range m4 {
		fmt.Printf("%s: %d\n", key, value)
	}

	// 长度
	fmt.Println("Map 长度:", len(m4))
}
