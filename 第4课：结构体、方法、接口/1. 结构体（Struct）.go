package main

import "fmt"

type User struct {
	ID       int
	Name     string
	Email    string
	Age      int
	IsActive bool
}

type Address struct {
	City    string
	Country string
}

type UserWithAddress struct {
	ID      int
	Name    string
	Address Address
}

func main() {
	user1 := User{
		ID:       1,
		Name:     "leaf",
		Email:    "linshen@example.com",
		Age:      20,
		IsActive: true,
	}

	user2 := User{2, "张三", "zhangsan@example.com", 25, true}

	user3 := User{
		ID:   3,
		Name: "李四",
	}

	fmt.Println(user1)
	fmt.Println(user2)
	fmt.Println(user3)

	fmt.Println("用户名:", user1.Name)
	fmt.Println("邮箱:", user1.Email)

	user1.Age = 24
	fmt.Println("新年龄:", user1.Age)

}
