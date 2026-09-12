package main

import (
	"errors"
	"fmt"
)

func divide(a, b int) (int, error) {
	if b == 0 {
		return 0, errors.New("除数不能为0")
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
		fmt.Println("错误:", err)
	} else {
		fmt.Println("结果:", result)
	}

	name, _, _ := getUserInfo(1)
	fmt.Println("用户名:", name)
}
