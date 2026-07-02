package main

import (
	"log"
	"part4/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ### 查询（Read）
func main() {
	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

	// 查询单条记录
	var user models.User

	// 根据主键查询
	db.First(&user, 1) // SELECT * FROM users WHERE id = 1;

	//根据条件查询
	db.Where("username = ?", "linshen").First(&user)

	// 查询多条记录
	var users []models.User
	db.Find(&users)

	// 条件查询
	db.Where("age > ?", 20).Find(&users)

	// 复杂查询
	db.Where("age > ? AND is_active = ?", 20, true).
		Order("created_at desc").
		Limit(10).
		Find(&users)

	//查询指定字段
	db.Select("username", "email").Find(&users)

	//统计
	var count int64
	db.Model(&models.User{}).Where("age > ?", 20).Count(&count)
	log.Printf("符合条件的用户数: %d\n", count)

}
