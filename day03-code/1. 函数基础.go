package main

import "fmt"

func greet(name string) string {
	return "Hello," + name
}

func add(a, b int) int {
	return a + b
}

func createUser(name string, age int, isVIP bool) {
	fmt.Printf("用户:%s,年龄:%d,VIP:%t\n", name, age, isVIP)
}

func main() {
	msg := greet("leaf")
	fmt.Println(msg)

	sum := add(1, 2)
	fmt.Println(sum)

	createUser("leaf", 23, true)
}
