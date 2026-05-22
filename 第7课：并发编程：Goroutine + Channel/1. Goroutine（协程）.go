package main

import (
	"fmt"
	"time"
)

func sayHello(name string) {
	for i := 0; i < 3; i++ {
		fmt.Printf("Hello,%s!(%d)\n", name, i)
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	sayHello("Alice")

	go sayHello("Bob")
	go sayHello("Charlie")

	time.Sleep(500 * time.Millisecond)
	fmt.Println("主函数结束")
}
