package database

import (
	"fmt"
	"skillsbay-api/config"
	"skillsbay-api/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var BD *gorm.DB

func InitDB() error {
	var err error

	dsn := config.AppConfig.GetDatabaseDSN()

	switch config.Appconfig.Database.Driver {
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
