package main

import "fmt"

func main() {
	var p1 *int
	var p2 *string
	var p3 *float64

	fmt.Println(p1, p2, p3)

	name := "leaf"
	p2 = &name

	balance := 100.5

	p3 = &balance

	p1 = new(int)
	*p1 = 42

	fmt.Println(*p1, *p2, *p3)
}
