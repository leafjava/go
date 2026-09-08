package main

import "fmt"

//func main() {
//	var name string = "林燊"
//	var age int = 23
//	var isStudent bool = true
//
//	fmt.Println(name, age, isStudent)
//
//	var city string
//	city = "广州"
//	fmt.Println(city)
//}

//func main() {
//	var name = "林燊"
//	var age = 23
//	var score = 95.5
//
//	fmt.Printf("%T,%T,%T\n", name, age, score)
//}

//func main() {
//	name := "林燊"
//	age := 23
//	city := "广州"
//
//	x, y, z := 1, 2, 3
//
//	fmt.Println(name, age, city)
//	fmt.Println(x, y, z)
//}

var (
	name   string = "林燊"
	age    int    = 23
	city   string = "广州"
	salary float64
)

func main() {
	salary = 15000.0
	fmt.Println(name, age, city, salary)
}
