package main

//主要用途：
//
//字符串转数字
//
//strconv.Atoi(str) - 字符串转整数（ASCII to Integer）
//strconv.ParseFloat(str, 64) - 字符串转浮点数
//strconv.ParseBool(str) - 字符串转布尔值
//数字转字符串
//
//strconv.Itoa(num) - 整数转字符串（Integer to ASCII）

//为什么需要 strconv？
//
//Go 是强类型语言，不能直接用类型转换语法（如 string(123)）来转换字符串和数字，必须用 strconv 包。这样设计是为了：
//
//处理转换可能失败的情况（比如 "abc" 转数字会报错）
//提供更多转换选项（进制、精度等）

//func main() {
//	var a int = 100
//	var b float64 = float64(a)
//	var c int32 = int32(a)
//
//	fmt.Println(a, b, c)
//
//	str := "123"
//	num, err := strconv.Atoi(str)
//	if err != nil {
//		fmt.Println(err)
//	} else {
//		fmt.Println("数字:", num)
//	}
//
//	age := 23
//	ageStr := strconv.Itoa(age) //数字转字符串
//	fmt.Println(ageStr)
//
//	PriceStr := "99.99"
//	price, _ := strconv.ParseFloat(PriceStr, 64)
//	fmt.Println(price)
//
//	boolStr := "true"
//	boolVal, _ := strconv.ParseBool(boolStr)
//	fmt.Println(boolVal)
//}

//func main() {
//	var i int
//	var f float64
//	var b bool
//	var s string
//
//	fmt.Printf("int: %d\n", i)
//	fmt.Printf("float64: %f\n", f)
//	fmt.Printf("bool: %t\n", b)
//	fmt.Printf("string: %s\n", s)
//
//}
