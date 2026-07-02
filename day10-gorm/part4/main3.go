package main

import (
	"your-project/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ### 更新（Update）
func mian() {
	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

	// 更新单个字段
	db.Model(&models.User{}).Where("id = ?", 1).Update("age", 24)

	// 更新多个字段
	db.Model(&models.User{}).Where("id = ?", 1).Updates(map[string]interface{}{
		"age":   24,
		"email": "newemail@example.com",
	})

	// 使用结构体更新
	var user models.User
	db.First(&user, 1)
	user.Age = 25
	user.Email = "updated@example.com"
	db.Save(&user)

	// 批量更新
	db.Model(&models.User{}).Where("age < ?", 18).Update("is_active", false)
}
