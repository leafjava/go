package main

import "fmt"

func add(a int, b int) int {
	return a + b
}

func divide(a int, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("Cannot divide by 0")
	}
	return a / b, nil
}

func main() {
	sum := add(1, 2)
	fmt.Println(sum)

	result, err := divide(10, 2)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(result)
	}
}
