package main

import (
	"errors"
	"fmt"
)

func divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, errors.New("divide by zero")
	}
	return a / b, nil
}

func getUserInfo(id int) (name string, age int, err error) {
	if id <= 0 {
		err = errors.New("无效的用户ID")
		return
	}

	name = "leaf"
	age = 23
	return
}

func main() {
	result, err := divide(10, 2)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(result)
	}
}
