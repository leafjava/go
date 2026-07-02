package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ### 多对多关系
type User struct {
	ID    uint
	Name  string
	Roles []Role `gorm:"many2many:user_roles;"`
}

type Role struct {
	ID    uint
	Name  string
	Users []User `gorm:"many2many:user_roles;"`
}

// 使用
func main() {
	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

	// 创建用户和角色
	user := User{Name: "linshen"}
	role1 := Role{Name: "admin"}
	role2 := Role{Name: "user"}

	db.Create(&user)
	db.Create(&role1)
	db.Create(&role2)

	// 关联
	db.Model(&user).Association("Roles").Append(&role1, &role2)

	// 查询
	var u User
	db.Preload("Roles").First(&u, user.ID)
}
