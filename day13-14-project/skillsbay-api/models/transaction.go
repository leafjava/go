package models

import "gorm.io/gorm"

type Transaction struct {
	gorm.Model
	WalletID uint    `gorm:"index" json:"wallet_id"`
	Hash     string  `gorm:"uniqueIndex;not null" json:"hash"`
	From     string  `gorm:"not null" json:"from"`
	To       string  `gorm:"not null" json:"to"`
	Amount   float64 `gorm:"not null" json:"amount"`
	GasUsed  int64   `json:"gas_used"`
	GasPrice float64 `json:"gas_price"`
	Status   string  `gorm:"default:'pending'" json:"status"`
	BlockNum int64   `json:"block_num"`
	Wallet   Wallet  `gorm:"foreignKey:WalletID" json:"wallet,omitempty"`
}
