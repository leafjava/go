package main

import "fmt"

func add(a, b float64) float64 {
	return a + b
}

func subtract(a, b float64) float64 {
	return a - b
}

func multiply(a, b float64) float64 {
	return a * b
}

func divide(a, b float64) (float64, error) {
	return a / b, nil
}

func main() {
	fmt.Println(add(10, 20))
	fmt.Println(subtract(10, 20))
	fmt.Println(multiply(10, 20))

	result, err := divide(10, 0)
	if err != nil {
		fmt.Println("错误:", err)
	} else {
		fmt.Println("10/0=", result)
	}
}
