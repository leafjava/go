package main

import "fmt"

func add(a int, b int) int {
	return a + b
}

func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, fmt.Errorf("除数不能为0")
	}
	return a / b, nil
}

func main() {
	sum := add(2, 2)
	fmt.Println(sum)

	sum2, err := divide(10, 2)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(sum2)
	}
}
