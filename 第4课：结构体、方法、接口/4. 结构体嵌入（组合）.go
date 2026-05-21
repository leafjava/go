package main

import "fmt"

type BaseModel struct {
	ID        int
	CreatedAt int64
	UpdatedAt int64
}

type User struct {
	BaseModel
	Name  string
	Email string
}

type Product struct {
	BaseModel
	Name  string
	Price float64
}

func main() {
	//user := User{
	//	BaseModel: BaseModel{
	//		ID:        1,
	//		CreatedAt: 1704067200,
	//		UpdatedAt: 1704067200,
	//	},
	//	Name:  "Leaf",
	//	Email: "leaf@qq.com",
	//}

	user := User{
		BaseModel: BaseModel{
			ID:        1,
			CreatedAt: 1704067200,
			UpdatedAt: 1704067200,
		},
		Name:  "John Doe",
		Email: "leaf@qq.com",
	}

	fmt.Println("用户ID：", user.ID)
	fmt.Println("用户名:", user.Name)
	fmt.Println("创建时间:", user.CreatedAt)
}
