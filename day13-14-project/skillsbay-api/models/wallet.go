package models

import (
	"gorm.io/gorm"
)

type Wallet struct {
	gorm.Model
	UserID       uint          `gorm:"index" json:"user_id"`
	Address      string        `gorm:"uniqueIndex;not null" json:"address"`
	Balance      float64       `gorm:"default:0" json:"balance"`
	Network      string        `gorm:"not null" json:"network"`
	User         User          `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Transactions []Transaction `gorm:"foreignKey:WalletID" json:"transactions,omitempty"`
}
