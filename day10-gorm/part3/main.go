package main

import (
	"log"
	"your-project/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	err = db.AutoMigrate(
		&models.User{},
		&models.Wallet{},
		&models.Transaction{},
		&models.NFT{},
	)
	if err != nil {
		log.Fatal("迁移失败:", err)
	}

	log.Println("数据库迁移成功")
}
