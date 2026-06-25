package main

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("连接数据库失败:", err)
	}

	log.Println("数据库连接成功")
}
