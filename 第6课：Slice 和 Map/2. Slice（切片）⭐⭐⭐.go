package main

import "fmt"

func main() {
	//var s1 []int
	//s2 := []int{}
	s3 := []int{1, 2, 3, 4, 5}
	//s4 := make([]int, 5)
	//s5 := make([]int, 5, 10)

	fmt.Println("s3[1:4]:", s3[1:4])
	fmt.Println("s3[:3]:", s3[:3])
	fmt.Println("s3[3:]:", s3[3:])

	fmt.Println("长度:", len(s3), "容量:", cap(s3))
}
