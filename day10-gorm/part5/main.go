package main

import (
	"log"
	"skillsbay-api4/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

//### 一对多关系

func main() {
	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

	// 预加载关联数据
	var wallet models.Wallet
	db.Preload("Transactions").First(&wallet, 1)

	log.Printf("钱包地址: %s, 交易数: %d\n", wallet.Address, len(wallet.Transactions))

	// 条件预加载
	db.Preload("Transactions", "status = ?", "confirmed").First(&wallet, 1)

	// 嵌套预加载
	db.Preload("Transactions.Wallet").First(&wallet, 1)

}
