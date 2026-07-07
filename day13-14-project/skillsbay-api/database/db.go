package database

import (
	"fmt"
	"skillsbay-api/config"
	"skillsbay-api/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() error {
	var err error

	dsn := config.AppConfig.GetDatabaseDSN()

	switch config.AppConfig.Database.Driver {
	case "sqlite":
		DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	default:
		return fmt.Errorf("不支持的数据库驱动: %s", config.AppConfig.Database.Driver)
	}

	if err != nil {
		return err
	}

	err = DB.AutoMigrate(
		&models.User{},
		&models.Wallet{},
		&models.Transaction{},
		&models.NFT{},
	)

	return err
}
