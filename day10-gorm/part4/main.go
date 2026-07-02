package main

import (
	"log"
	"your-project/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ### 创建（Create）
func main() {
	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

	// 创建单条记录
	user := models.User{
		Username: "linshen",
		Email:    "linshen@example.com",
		Password: "hashed_password",
		Age:      23,
	}

	result := db.Create(&user)
	if result.Error != nil {
		log.Fatal("创建失败:", result.Error)
	}

	log.Printf("创建成功，ID: %d, 影响行数: %d\n", user.ID, result.RowsAffected)

	// 批量创建
	users := []models.User{
		{Username: "alice", Email: "alice@example.com", Password: "pass1", Age: 25},
		{Username: "bob", Email: "bob@example.com", Password: "pass2", Age: 30},
	}

	db.Create(&users)
}
