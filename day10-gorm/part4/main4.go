package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

	// 软删除（设置 DeletedAt）
	db.Delete(&models.User{}, 1)

	// 永久删除
	db.Unscoped().Delete(&models.User{}, 1)

	// 批量删除
	db.Where("age < ?", 18).Delete(&models.User{})

	// 查询包含软删除的记录
	var users []models.User
	db.Unscoped().Find(&users)
}
