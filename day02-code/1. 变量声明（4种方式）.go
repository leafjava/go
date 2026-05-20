package main

//方式1：var 完整声明
//func main() {
//	var name string = "leaf"
//	var age int = 23
//	var isStudent bool = true
//
//	fmt.Println(name, age, isStudent)
//
//	var city string
//	city = "广州"
//	fmt.Println(city)
//
//}

//方式2：类型推断
//func main() {
//	var name = "leaf"
//	var age = 23
//	var score = 95.5
//
//	fmt.Printf("%T,%T,%T\n", name, age, score)
//}

// 方式3：短声明（最常用）⭐
//func main() {
//	name := "leaf"
//	age := 23
//	city := "guangzhou"
//
//	x, y, z := 1, 2, 3
//
//	fmt.Println(name, age, city)
//	fmt.Println(x, y, z)
//}

// 方式4：批量声明
//var (
//	name   string = "leaf"
//	age    int    = 23
//	city   string = "guangzhou"
//	salary float64
//)
//
//func main() {
//	salary = 15000.0
//	fmt.Println(name, age, city, salary)
//}
