package main

import "fmt"

func main() {
	age := 23
	fmt.Println("age的值:", age)
	fmt.Println("age的地址:", &age)

	var ptr *int = &age
	fmt.Println("ptr的值(地址)", ptr)
	fmt.Println("ptr指向的值:", *ptr)

	*ptr = 24
	fmt.Println("修改后age的值:", age)
}

//简短解释 var ptr *int = &age 的含义和要点：
//
//含义：ptr 是一个变量，类型是“指向 int 的指针”（*int），它保存了 age 变量的内存地址。右边的 &age 表示取 age 的地址。
//组成：
//var ptr *int：声明一个指针变量 ptr，类型为 *int（指向 int）。
//= &age：把 age 的地址赋给 ptr，因此 ptr 指向 age。
//读取与修改：
//*ptr 表示“解引用”，得到指针 ptr 指向的那个 int 的值（等同于 age 的值）。
//给 *ptr 赋值（例如 *ptr = 24）会修改 ptr 指向的变量，也就是修改 age。
//快速记忆：
//&x = x 的地址（地址值），*p = 指针 p 指向的值。
//指针类型必须匹配（*int 只能指向 int）。
//其他注意：
//指针的零值是 nil，声明但不赋值的 var p *int 为 nil，此时不能解引用（会 panic）。
//可以用短声明 ptr := &age 更简洁。
//示例对应你代码的关系（说明输出含义）：
//
//fmt.Println("age的地址:", &age) 显示 age 的内存地址（ptr 的值）。
//fmt.Println("ptr的值(地址)", ptr) 会输出与上面相同的地址。
//fmt.Println("ptr指向的值:", *ptr) 输出指针指向的值（等于 age）。
//执行 *ptr = 24 后，age 的值被修改为 24。
