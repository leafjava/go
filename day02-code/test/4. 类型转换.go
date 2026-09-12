package main

import (
	"fmt"
	"strconv"
)

func main() {
	var a int = 100
	var b float64 = float64(a)
	var c int32 = int32(a)

	fmt.Println(a, b, c)

	str := "123"
	num, err := strconv.Atoi(str)
	if err != nil {
		fmt.Println("转换失败:", err)
	} else {
		fmt.Println("数字:", num)
	}

	age := 23
	ageStr := strconv.Itoa(age)
	fmt.Println("年龄字符串:", ageStr)

	priceStr := "99.99"
	price, _ := strconv.ParseFloat(priceStr, 64)
	fmt.Println("价格:", price)

	boolStr := "true"
	boolVal, _ := strconv.ParseBool(boolStr)
	fmt.Println("布尔值:", boolVal)
}
