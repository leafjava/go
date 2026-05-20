package main

import "fmt"

//func main() {
//	var name string = "leaf"
//
//	var age = 23
//
//	city := "广州"
//
//	fmt.Println(name, age, city)
//}

func add(a int, b int) int {
	return a + b
}

func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("Cannot divide by 0")
	}
	return a / b, nil
}

func main() {
	sum := add(10, 20)
	fmt.Println("10+20=", sum)

	result, err := divide(10, 2)

	if err != nil {
		fmt.Println("错误:", err)
	} else {
		fmt.Println("1=/2=", result)
	}
}
