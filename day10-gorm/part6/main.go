package main

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

	// 方式1：手动事务
	tx := db.Begin()

	if err := tx.Create(&models.User{Username: "test"}).Error; err != nil {
		tx.Rollback()
		return
	}

	if err := tx.Create(&models.Wallet{Address: "0x..."}).Error; err != nil {
		tx.Rollback()
		return
	}

	tx.Commit()

	// 方式2：自动事务
	err := db.Transaction(func(tx *gorm.DB) error {
		// 创建用户
		user := models.User{Username: "linshen"}
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		// 创建钱包
		wallet := models.Wallet{
			userID:  user.ID,
			Address: "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
			Network: "Ethereum",
		}
		if err := tx.Create(&wallet).Error; err != nil {
			return err
		}

		// 返回 nil 提交事务
		return nil

	})

	if err != nil {
		// 事务回滚
		return
	}

}
